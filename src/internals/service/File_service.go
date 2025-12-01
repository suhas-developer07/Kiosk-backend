package service

import (
	"context"
	"io"
	"time"

	domain "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/Files"
	db "github.com/suhas-developer07/Kiosk-backend/src/internals/repository/Files_repo"
	"github.com/suhas-developer07/Kiosk-backend/src/pkg/filestore"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FileService struct {
	FileRepo   *db.FilesRepo
	Storage    filestore.FileStorage
}

func NewFileService(repo *db.FilesRepo, storage filestore.FileStorage) *FileService {
	return &FileService{
		FileRepo: repo,
		Storage:  storage,
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
		FileName:     req.FileName,
		FileURL:      fileURL,
		Description:  req.Description,
		Subject:      req.Subject,
		FacultyID:    req.FacultyID,
		GroupAllowed: req.GroupAllowed,
		Type:         req.Type,
		Date:         primitive.NewDateTimeFromTime(time.Now()),
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
