package Fileservice

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	domain "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/files"
	"github.com/suhas-developer07/Kiosk-backend/src/internals/domain/subjects"
	db "github.com/suhas-developer07/Kiosk-backend/src/internals/repository/files_repo"

	"github.com/suhas-developer07/Kiosk-backend/src/pkg/filestore"
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

func (s *FileService) GetFileByGradeAndSubjectService(
	ctx context.Context,
	grade string,
	subject string,
) ([]domain.File, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	grade = strings.TrimSpace(strings.ToUpper(grade))
	subject = strings.TrimSpace(strings.ToLower(subject))

	if grade != "1PUC" && grade != "2PUC" {
		return nil, domain.ErrInvalidGrade
	}

	if subject == "" {
		return nil, domain.ErrInvalidSubject
	}

	if !subjects.IsValidSubject(subject) {
		return nil, fmt.Errorf("service:subject is not valid type")
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

func (s *FileService) CreatePrintJobService(
	ctx context.Context,
	req domain.PrintJobPayload,
) (string, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	s.Logger.Infof(
		"Creating print Job | file_id=%s | copies=%d | PageLayout=%s",
		req.FileID.Hex(), req.Copies, req.PageLayout,
	)

	if req.Copies < 1 || req.Copies > 100 {
		return "", domain.ErrInvalidCopies
	}

	exists, err := s.FileRepo.GetFileByID(ctx, req.FileID.Hex())
	if err != nil {
		if errors.Is(err, domain.ErrInvalidID) {
			return "", domain.ErrInvalidID
		}
		if errors.Is(err, domain.ErrFileNotFound) {
			return "", domain.ErrFileNotFound
		}
		return "", fmt.Errorf("service: db error while checking file: %w", err)
	}

	if !exists {
		return "", domain.ErrFileNotFound
	}

	printJob := domain.PrintJob{
		FileID:              req.FileID,
		Copies:              req.Copies,
		PrintingSide:        req.PrintingSide,
		PrintingMode:        req.PrintingMode,
		PageRange:           req.PageRange,
		PageLayout:          req.PageLayout,
		OrderStatus:         "Initialized",
		Price:               req.Price,
		TotalSheetsRequired: req.TotalSheets,
		CreatedAt:           time.Now(),
	}

	err = s.FileRepo.CreatePrintJob(ctx, printJob)
	if err != nil {
		return "", fmt.Errorf("service: create print job failed: %w", err)
	}

	return "", nil
}

func (s *FileService) AccessFileService(
	ctx context.Context,
	fileID string,
) (string, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	fileID = strings.TrimSpace(fileID)
	if fileID == "" {
		return "", domain.ErrInvalidID
	}

	if _, err := primitive.ObjectIDFromHex(fileID); err != nil {
		return "", domain.ErrInvalidID
	}

	s.Logger.Infow(
		"accessing file",
		"file_id", fileID,
	)

	fileKey, err := s.FileRepo.GetFileKeyfromtheFileID(ctx, fileID)
	if err != nil {
		if errors.Is(err, domain.ErrFileNotFound) {
			return "", domain.ErrFileNotFound
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
