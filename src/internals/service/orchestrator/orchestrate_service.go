package orchestrator

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	model "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/faculties"
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

	fileData := domain.File{
		Title:        req.Title,
		FileKey:      fileKey,
		Grade:        strings.ToUpper(strings.TrimSpace(req.Grade)),
		Subject:      strings.ToLower(strings.TrimSpace(req.Subject)),
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

func (s *OrchestrateService) AddFacultyService(ctx context.Context, req model.UpdateFaculty) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	//todo: need to validate the req

	err := s.FacultyRepo.AddFaculty(ctx, req)

	if err!=nil{
		return err
	}

	return nil
}
