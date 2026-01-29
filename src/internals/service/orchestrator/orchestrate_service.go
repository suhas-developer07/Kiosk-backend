package orchestrator

import (
	"context"
	"errors"
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
	"github.com/suhas-developer07/Kiosk-backend/src/pkg/utils"
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

func (s *OrchestrateService) AddFacultyService(ctx context.Context, req model.FacultyPayload) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	//validating the subjects
	for _,sub := range req.Subjects{
		if !subjects.IsValidSubject(string(sub)){
			return fmt.Errorf("Invalid subject :%s",sub)
		}
	}

	if req.Password != "" {
		hashed, err := utils.HashPassword(req.Password)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}
		req.Password = hashed
	}

	s.Logger.Infof("Adding Faculty| email=%s", req.Email)

	faculty := model.Faculty{
		ID:        primitive.NewObjectID(),
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password,
		Subjects:  req.Subjects,
		Stream:    req.Stream,
		Gender:    req.Gender,
		CreatedAt: time.Now(),
	}

	err := s.FacultyRepo.CreateAccount(ctx, faculty)
	switch {
	case errors.Is(err, model.ErrEmailAlreadyExists):
		return model.ErrEmailAlreadyExists

	case err != nil:
		return fmt.Errorf("failed to add faculty: %w", err)
	}
	s.Logger.Infof("Failed to add faculty successfully | email=%s", req.Email)

	return nil
}
