package service

import (
	"context"

	models "github.com/suhas-developer07/Kiosk-backend/src/internals/Models"
	repo "github.com/suhas-developer07/Kiosk-backend/src/internals/repository/files_repo"
)

type FileService struct {
	FilesRepo *repo.FilesRepo
}

func NewFileService(r *repo.FilesRepo) *FileService {
	return &FileService{
		FilesRepo: r,
	}
}

func (s *FileService) UploadFileService(ctx context.Context, req models.FileUploadRequest) error {
	//TODO: call the function to upload file to storage (eg: S3) and get the file URL
	return nil
}

// func (s *FileService) InsertFile(ctx context.Context, req models.File) (*mongo.InsertOneResult, error) {

// }

// func (s *FileService) InsertPrintJob(ctx context.Context, job models.PrintJob) (*mongo.InsertOneResult, error) {

// }