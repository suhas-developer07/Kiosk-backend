package orchestrator

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	domain "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/files"
	"github.com/suhas-developer07/Kiosk-backend/src/internals/domain/subjects"
	facultydb "github.com/suhas-developer07/Kiosk-backend/src/internals/repository/faculty_repo"
	filedb "github.com/suhas-developer07/Kiosk-backend/src/internals/repository/files_repo"
	"github.com/suhas-developer07/Kiosk-backend/src/pkg/filestore"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type UploadService struct {
	FileRepo    *filedb.FilesRepo
	FacultyRepo *facultydb.FacultyRepo
	Storage     filestore.FileStorage
	Logger      *zap.SugaredLogger
}

func NewUploadService(FileRepo *filedb.FilesRepo, FacultyRepo *facultydb.FacultyRepo, storage filestore.FileStorage, Logger *zap.SugaredLogger) *UploadService {
	return &UploadService{
		FileRepo:    FileRepo,
		FacultyRepo: FacultyRepo,
		Storage:     storage,
		Logger:      Logger,
	}
}

func (s *UploadService) UploadFileService(
	ctx context.Context,
	filename string,
	file io.Reader,
	req domain.FileUploadRequest,
	facultyID string,
) (string, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	fmt.Print("debugging service layer")

	facultyID = strings.TrimSpace(facultyID)
	if facultyID == "" {
		return "", domain.ErrInvalidID
	}

	objID, err := primitive.ObjectIDFromHex(facultyID)
	if err != nil {
		return "", domain.ErrInvalidID
	}

	if !subjects.IsValidSubject(req.Subject) {
		return "", fmt.Errorf("service:subject is not valid")
	}

	checkSubject,err := s.FacultyRepo.HasSubject(ctx,objID,subjects.Subject(req.Subject))

	if !checkSubject {
		return "",fmt.Errorf("You can't have an access to upload this subject files ")
	}

	fileKey, etag, err := s.Storage.Save(
		ctx,
		file,
		filename,
		req.Grade,
		req.Subject,
	)
	if err != nil {
		return "", err
	}

	fileData := domain.File{
		Title:        req.Title,
		FileKey:      fileKey,
		Grade:        strings.ToUpper(strings.TrimSpace(req.Grade)),
		Subject:      strings.ToLower(strings.TrimSpace(req.Subject)),
		Description:  req.Description,
		FacultyID:    objID,
		GroupAllowed: req.GroupAllowed,
		ETag:         strings.Trim(etag, `"`),
		FileType:     req.FileType,
		UploadedAt:   primitive.NewDateTimeFromTime(time.Now()),
	}

	err = s.FileRepo.WithTransaction(ctx, func(sc mongo.SessionContext) error {
		return s.FileRepo.SaveFileRecord(sc, fileData)
	})

	if err != nil {
		_ = s.Storage.Delete(ctx, fileKey)
		return "", err
	}

	return fileKey, nil
}
