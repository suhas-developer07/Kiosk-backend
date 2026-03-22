package Fileservice

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	apperrors "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/errors"
	domain "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/files"
	"github.com/suhas-developer07/Kiosk-backend/src/internals/domain/subjects"
	db "github.com/suhas-developer07/Kiosk-backend/src/internals/repository/files_repo"

	"github.com/suhas-developer07/Kiosk-backend/src/pkg/filestore"
	"github.com/suhas-developer07/Kiosk-backend/src/pkg/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type FileService struct {
	FileRepo *db.FilesRepo
	Storage  filestore.FileStorage
	Logger   *zap.SugaredLogger
}

func NewFileService(repo *db.FilesRepo, storage filestore.FileStorage, Logger *zap.SugaredLogger) *FileService {
	return &FileService{
		FileRepo: repo,
		Storage:  storage,
		Logger:   Logger,
	}
}

func (s *FileService) GetFileByGradeAndSubjectService(ctx context.Context, grade string, subject string) ([]domain.File, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	grade = strings.TrimSpace(strings.ToUpper(grade))
	subject = strings.TrimSpace(strings.ToLower(subject))

	if grade != "1PUC" && grade != "2PUC" {
		return nil, apperrors.ErrInvalidGrade
	}

	if subject == "" {
		return nil, apperrors.ErrInvalidSubject
	}

	if !subjects.IsValidSubject(subject) {
		return nil, apperrors.ErrInvalidSubject
	}

	s.Logger.Infof("fetching files: grade=%s subject=%s", grade, subject)

	files, err := s.FileRepo.GetFileByGradeAndSubject(ctx, grade, subject)

	if err != nil {
		return nil, fmt.Errorf("service: get files: %w", err)
	}

	if len(files) == 0 {
		return []domain.File{}, nil
	}

	return files, nil
}

func (s *FileService) CreatePrintJobService(ctx context.Context, req domain.PrintJobPayload) error {

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	s.Logger.Infof(
		"Creating print Job | file_id=%s | copies=%d | PageLayout=%s",
		req.FileID.Hex(), req.Copies, req.PageLayout,
	)

	if req.Copies < 1 || req.Copies > 100 {
		return apperrors.ErrInvalidCopies
	}

	exists, err := s.FileRepo.GetFileByID(ctx, req.FileID.Hex())
	if err != nil {
		if errors.Is(err, apperrors.ErrInvalidID) {
			return apperrors.ErrInvalidID
		}
		if errors.Is(err, apperrors.ErrFileNotFound) {
			return apperrors.ErrFileNotFound
		}
		return fmt.Errorf("service: db error while checking file: %w", err)
	}

	if !exists {
		return apperrors.ErrFileNotFound
	}

	card, err := s.FileRepo.GetRFIDCardById(ctx, req.CardID)
	if err != nil {
		return fmt.Errorf("service:failed to get usn by card id: %w", err)
	}

	BalanceInt, err := strconv.Atoi(card.Balance)
	if err != nil {
		return fmt.Errorf("Balance is not in correct formate. please contact support,Error:%v",err)
	}

	newBalance := BalanceInt - req.Amount

	newBalanceStr := strconv.Itoa(newBalance)

	if err := s.FileRepo.UpdateCardBalance(ctx, card.CardID, newBalanceStr); err != nil {
		return fmt.Errorf("Error updating Card Balance,Error:%v", err)
	}

	printJob := domain.PrintJob{
		FileId:              req.FileID,
		USN:                 card.USN,
		FileName:            req.FileName,
		Copies:              req.Copies,
		PrintingSide:        req.PrintingSide,
		PrintingMode:        req.PrintingMode,
		PageRange:           req.PageRange,
		PageLayout:          req.PageLayout,
		OrderStatus:         "Initialized", //we need to some how update this order status after printing
		Amount:              req.Amount,
		TotalNumberOfSheets: req.TotalSheets,
		CreatedAt:           time.Now(),
	}

	err = s.FileRepo.CreatePrintJob(ctx, printJob)
	if err != nil {
		return fmt.Errorf("service: create print job failed: %w", err)
	}

	ok := utils.UpdateRemainingPagesInTray(req.TotalSheets)
	if !ok {
		return errors.New("service:Page update error")
	}

	return nil
}

func (s *FileService) AccessFileService(
	ctx context.Context,
	fileID string,
) (string, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	fileID = strings.TrimSpace(fileID)
	if fileID == "" {
		return "", apperrors.ErrInvalidID
	}

	if _, err := primitive.ObjectIDFromHex(fileID); err != nil {
		return "", apperrors.ErrInvalidID
	}

	s.Logger.Infow(
		"accessing file",
		"file_id", fileID,
	)

	fileKey, err := s.FileRepo.GetFileKeyfromtheFileID(ctx, fileID)
	if err != nil {
		if errors.Is(err, apperrors.ErrFileNotFound) {
			return "", apperrors.ErrFileNotFound
		}

		return "", fmt.Errorf(
			"service: failed to fetch file key for file_id=%s: %w",
			fileID, err,
		)
	}

	if fileKey == "" {
		s.Logger.Errorw(
			"empty file key returned from repository",
			"file_id", fileID,
		)
		return "", errors.New("internal error: empty file key")
	}

	signedURL, err := s.Storage.GenerateSignedURL(ctx, fileKey)
	if err != nil {
		return "", fmt.Errorf(
			"service: failed to generate signed url for key=%s: %w",
			fileKey, err,
		)
	}

	s.Logger.Infow(
		"signed url generated successfully",
		"file_id", fileID,
	)

	return signedURL, nil
}

func (s *FileService) FetchPrintJobsService(ctx context.Context) ([]domain.PrintJob, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	s.Logger.Info("Fetching printJOB details")

	data, err := s.FileRepo.FetchPrintJobs(ctx)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return []domain.PrintJob{}, nil
	}

	return data, nil
}

func (s *FileService) CalculateTotalRevenueService(ctx context.Context) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	s.Logger.Info("Calculating total revenue")

	totalRevenue, err := s.FileRepo.CalculateTotalRevenue(ctx)
	if err != nil {
		return 0, err
	}

	return totalRevenue, nil
}

func (s *FileService) GetRecentPrintJobsService(ctx context.Context) ([]domain.PrintJob, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	s.Logger.Info("Fetching recent print jobs")

	printJobs, err := s.FileRepo.GetRecentPrintJobs(ctx)
	if err != nil {
		return nil, err
	}

	if len(printJobs) == 0 {
		return []domain.PrintJob{}, nil
	}

	return printJobs, nil
}
