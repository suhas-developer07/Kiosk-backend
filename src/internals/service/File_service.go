package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	domain "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/Files"
	db "github.com/suhas-developer07/Kiosk-backend/src/internals/repository/Files_repo"
	"github.com/suhas-developer07/Kiosk-backend/src/pkg/filestore"
	"github.com/suhas-developer07/Kiosk-backend/src/pkg/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type FileService struct {
	FileRepo *db.FilesRepo
	Storage  filestore.FileStorage
	Logger   *zap.SugaredLogger
}

func NewFileService(repo *db.FilesRepo, storage filestore.FileStorage,Logger *zap.SugaredLogger) *FileService {
	return &FileService{
		FileRepo: repo,
		Storage:  storage,
		Logger:Logger,
	}
}

func (s *FileService) UploadFileService(
	ctx context.Context,
	filename string,
	file io.Reader,
	req domain.FileUploadRequest,
) (string, error) {

	//TODO:Need to check file type if its not .PDF convert that into .PDF formate

	fileURL, err := s.Storage.Save(file, filename)
	if err != nil {
		return "", err
	}

	fileData := domain.File{
		Title:        req.Title,
		FileURL:      fileURL,
		Description:  req.Description,
		Subject:      req.Subject,
		FacultyID:    req.FacultyID,
		GroupAllowed: req.GroupAllowed,
		FileType:     req.FileType,
		UploadedAt:   primitive.NewDateTimeFromTime(time.Now()),
	}

	err = s.FileRepo.WithTransaction(ctx, func(sc mongo.SessionContext) error {
		return s.FileRepo.SaveFileRecord(sc, fileData)
	})

	if err != nil {
		_ = s.Storage.Delete(fileURL)
		return "", err
	}

	return fileURL, nil
}

func (s *FileService) GetFileByGradeAndSubjectService(
	ctx context.Context,
	grade string,
	subject string,
) ([]domain.File, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	grade = strings.TrimSpace(strings.ToUpper(grade))
	subject = strings.TrimSpace(strings.Title(subject))

	if grade != "1PUC" && grade != "2PUC" {
		return nil, domain.ErrInvalidGrade
	}

	//Todo: need to validate a subject in enum list
	if subject == "" {
		return nil, domain.ErrInvalidSubject
	}

	s.Logger.Infof("fetching files: grade=%s subject=%s", grade, subject)

	files, err := s.FileRepo.GetFileByGradeAndSubject(ctx, grade, subject)
	if err != nil {
		return nil, fmt.Errorf("service: get files: %w", err)
	}

	if len(files) == 0 {
		return []domain.File{}, nil
	}

	//TODO: Need to check if the file exist in s3

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

    //TODO1 : Generate an Token for the JOB Store that into an DB
    //TODO 2 : calculate the price for the JOB  --> DONE  
	
	TotalSheetsRequired,Price := utils.CalculatePrintJob(req.PageRange,req.PageLayout,req.PrintingSide,req.PrintingMode,req.Copies)

	if TotalSheetsRequired < 0 || Price <0{
		return "",fmt.Errorf("Error while calculating the cost")
	}

    printJob := domain.PrintJob{
        FileID:       req.FileID,
        Copies:       req.Copies,
        PrintingSide: req.PrintingSide,
        PrintingMode: req.PrintingMode,
        PageRange:    req.PageRange,
        PageLayout:   req.PageLayout,
        OrderStatus:  "Initialized",
		Price: Price,
		TotalSheetsRequired: TotalSheetsRequired,
        CreatedAt:    time.Now(),
    }

    err = s.FileRepo.CreatePrintJob(ctx, printJob)
    if err != nil {
        return "", fmt.Errorf("service: create print job failed: %w", err)
    }

    return "", nil
}
