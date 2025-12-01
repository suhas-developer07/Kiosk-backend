package service

import (
	"context"
	"time"

	 domain "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/Files"
	 repo "github.com/suhas-developer07/Kiosk-backend/src/internals/repository/Files_repo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FileService struct {
	FilesRepo *repo.FilesRepo
}

func NewFileService(r *repo.FilesRepo) *FileService {
	return &FileService{
		FilesRepo: r,
	}
}

func (s *FileService) UploadFileService(ctx context.Context, req domain.FileUploadRequest) error {
	//TODO: call the function to upload file to storage (eg: S3) and get the file URL
	//validate the file type and other deatails

	File := domain.File{
		FileName: req.FileName,
		FileURL:  "https://example.com/" + req.FileName, // Placeholder URL
		Description: req.Description,
		Subject: req.Subject,
		FacultyID: req.FacultyID,
		GroupAllowed: req.GroupAllowed,
		Type: req.Type,
		Date: primitive.NewDateTimeFromTime(time.Now().UTC()),

	}

	err := s.FilesRepo.InsertFile(ctx,File);
	if err != nil {
		return err	
	}

	return nil
}

// func (s *FileService) InsertFile(ctx context.Context, req models.File) (*mongo.InsertOneResult, error) {

// }

// func (s *FileService) InsertPrintJob(ctx context.Context, job models.PrintJob) (*mongo.InsertOneResult, error) {

// }