package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	domain "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/files"
	"github.com/suhas-developer07/Kiosk-backend/src/internals/domain/subjects"
	facultydb "github.com/suhas-developer07/Kiosk-backend/src/internals/repository/faculty_repo"
	filedb "github.com/suhas-developer07/Kiosk-backend/src/internals/repository/files_repo"
	"github.com/suhas-developer07/Kiosk-backend/src/pkg/filestore"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type OrchestrateService struct {
	FileRepo    *filedb.FilesRepo
	FacultyRepo *facultydb.FacultyRepo
	Storage     filestore.FileStorage
	Logger      *zap.SugaredLogger
}

func NewOrchestrateService(FileRepo *filedb.FilesRepo, FacultyRepo *facultydb.FacultyRepo, storage filestore.FileStorage, Logger *zap.SugaredLogger) *OrchestrateService {
	return &OrchestrateService{
		FileRepo:    FileRepo,
		FacultyRepo: FacultyRepo,
		Storage:     storage,
		Logger:      Logger,
	}
}

func (s *OrchestrateService) UploadFileService(
	ctx context.Context,
	filename string,
	file io.Reader,
	req domain.FileUploadRequest,
	facultyID string,
) (string, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

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

	checkSubject, err := s.FacultyRepo.HasSubject(ctx, objID, subjects.Subject(req.Subject))

	if !checkSubject {
		return "", fmt.Errorf("You can't have an access to upload this subject file ")
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

	faculty, err := s.FacultyRepo.GetFacultyByID(ctx, facultyID) //Todo:need to be test

	if err != nil {
		return "", fmt.Errorf("service:internal error,not able to get faculty by his id")
	}

	req.Category = strings.TrimSpace(strings.ToLower(req.Category))

	fileData := domain.File{
		Title:        req.Title,
		FileKey:      fileKey,
		Grade:        strings.ToUpper(strings.TrimSpace(req.Grade)),
		Subject:      strings.ToLower(strings.TrimSpace(req.Subject)),
		Category:     req.Category,
		Description:  req.Description,
		FacultyID:    objID,
		FacultyName:  faculty.Username,
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

func (s *OrchestrateService) InitiateUpload(
	ctx context.Context,
	req domain.FileUploadRequest,
	facultyID string,
) (map[string]string, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	facultyID = strings.TrimSpace(facultyID)
	objID, err := primitive.ObjectIDFromHex(facultyID)
	if err != nil {
		return nil, domain.ErrInvalidID
	}

	if !subjects.IsValidSubject(req.Subject) {
		return nil, fmt.Errorf("invalid subject")
	}

	allowed, err := s.FacultyRepo.HasSubject(ctx, objID, subjects.Subject(req.Subject))
	if err != nil || !allowed {
		return nil, fmt.Errorf("no access to upload this subject file")
	}

	fileKey := fmt.Sprintf(
		"grade-%s/%s/%s-%d",
		strings.ToUpper(req.Grade),
		strings.ToLower(req.Subject),
		uuid.NewString(),
		time.Now().Unix(),
	)

	uploadURL, err := s.Storage.GeneratePresignedPutURL(ctx, fileKey)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"upload_url": uploadURL,
		"file_key":   fileKey,
	}, nil
}

func (s *OrchestrateService) CompleteUpload(
	ctx context.Context,
	fileKey string,
	req domain.FileUploadRequest,
	facultyID string,
) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(facultyID)
	if err != nil {
		return domain.ErrInvalidID
	}

	faculty, err := s.FacultyRepo.GetFacultyByID(ctx, facultyID)
	if err != nil {
		return err
	}

	req.Category = strings.TrimSpace(strings.ToLower(req.Category))

	fileData := domain.File{
		Title:        req.Title,
		FileKey:      fileKey,
		Grade:        strings.ToUpper(req.Grade),
		Subject:      strings.ToLower(req.Subject),
		Category:     req.Category,
		Description:  req.Description,
		FacultyID:    objID,
		FacultyName:  faculty.Username,
		GroupAllowed: req.GroupAllowed,
		FileType:     req.FileType,
		UploadedAt:   primitive.NewDateTimeFromTime(time.Now()),
	}

	return s.FileRepo.WithTransaction(ctx, func(sc mongo.SessionContext) error {
		return s.FileRepo.SaveFileRecord(sc, fileData)
	})
}


func (s *OrchestrateService) GetRecentUploadedFilesByFacultyIDService(ctx context.Context, FacultyID string) ([]domain.File, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if FacultyID == "" {
		return nil, fmt.Errorf("Faculty id is empty")
	}

	files, err := s.FileRepo.GetRecentUploadedFilesByFacultyID(ctx, FacultyID)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return []domain.File{}, nil
	}

	return files, nil
}

func (s *OrchestrateService) FileDeleteRequestService(ctx context.Context, fileID string, reason string) error {

	if fileID == "" {
		return errors.New("fileId cannot be empty")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return s.FileRepo.DeleteFileRequest(ctx, fileID, reason)
}

func (s *OrchestrateService) GetPendingDeleteRequestsService(ctx context.Context, facultyId string) ([]domain.File, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	if facultyId == "" {
		return nil, errors.New("facultyId cannot be empty")
	}

	return s.FileRepo.GetPendingDeleteRequests(ctx, facultyId)
}