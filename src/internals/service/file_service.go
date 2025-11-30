package service

import (
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

// func (s *FileService) InsertFile(ctx context.Context, req models.File) (*mongo.InsertOneResult, error) {

// }

// func (s *FileService) InsertPrintJob(ctx context.Context, job models.PrintJob) (*mongo.InsertOneResult, error) {

// }