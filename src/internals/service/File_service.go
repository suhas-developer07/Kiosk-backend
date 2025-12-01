package service

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	domain "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/Files"
	db "github.com/suhas-developer07/Kiosk-backend/src/internals/repository/Files_repo"
	"github.com/suhas-developer07/Kiosk-backend/src/pkg/filestore"
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

	return files, nil
}
