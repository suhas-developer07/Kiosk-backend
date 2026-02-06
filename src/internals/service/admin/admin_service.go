package admin

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	apperrors "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/errors"
	facultymodel "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/faculties"
	filemodel "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/files"
	model "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/admin"
	"github.com/suhas-developer07/Kiosk-backend/src/internals/domain/subjects"
	db "github.com/suhas-developer07/Kiosk-backend/src/internals/repository/admin"
	facultydb "github.com/suhas-developer07/Kiosk-backend/src/internals/repository/faculty_repo"
	filesdb "github.com/suhas-developer07/Kiosk-backend/src/internals/repository/files_repo"
	"github.com/suhas-developer07/Kiosk-backend/src/pkg/filestore"
	"github.com/suhas-developer07/Kiosk-backend/src/pkg/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type AdminService struct {
	AdminRepo   *db.AdminRepo
	FacultyRepo *facultydb.FacultyRepo
	FileRepo    *filesdb.FilesRepo
	Storage     filestore.FileStorage
	Logger      *zap.SugaredLogger
}

func NewAdminService(AdminRepo *db.AdminRepo, FacultyRepo *facultydb.FacultyRepo, FilesRepo *filesdb.FilesRepo, Storage filestore.FileStorage, sugar *zap.SugaredLogger) *AdminService {
	return &AdminService{
		AdminRepo:   AdminRepo,
		FacultyRepo: FacultyRepo,
		FileRepo:    FilesRepo,
		Storage:     Storage,
		Logger:      sugar,
	}
}

func (s *AdminService) GetFacultiesService(ctx context.Context,) ([]facultymodel.Faculty, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	faculties, err := s.FacultyRepo.GetFaculties(ctx)
	if err != nil {
		return nil, err
	}

	if faculties == nil {
		return []facultymodel.Faculty{}, nil
	}

	return faculties, nil
}

func (s *AdminService) GetFacultiesByStreamService(
	ctx context.Context,
	stream string,
) ([]facultymodel.Faculty, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	stream = strings.TrimSpace(strings.ToLower(stream))
	if stream == "" {
		return []facultymodel.Faculty{}, nil
	}

	faculties, err := s.FacultyRepo.GetFacultiesByStream(ctx, stream)
	if err != nil {
		return nil, err
	}

	if faculties == nil {
		return []facultymodel.Faculty{}, nil
	}

	return faculties, nil
}

func (s *AdminService) AddFacultyService(ctx context.Context, req facultymodel.FacultyPayload) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	//validating the subjects
	for _, sub := range req.Subjects {
		if !subjects.IsValidSubject(string(sub)) {
			return fmt.Errorf("Invalid subject :%s", sub)
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

	faculty := facultymodel.Faculty{
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
	case errors.Is(err, facultymodel.ErrEmailAlreadyExists):
		return facultymodel.ErrEmailAlreadyExists

	case err != nil:
		return fmt.Errorf("failed to add faculty: %w", err)
	}
	s.Logger.Infof("Failed to add faculty successfully | email=%s", req.Email)

	return nil
}

func (s *AdminService) GetTotalFacultiesCountService(ctx context.Context) (uint32, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	faculties, err := s.FacultyRepo.GetFaculties(ctx)

	if err != nil {
		return 0, err
	}
	var count uint32 = 0

	for range faculties {
		count++
	}

	if count == 0 {
		return 0, nil
	}
	return count, nil
}

func (s *AdminService) GetTotalFilesCountService(ctx context.Context) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	count, err := s.FileRepo.GetTotalFilesCount(ctx)

	if err != nil {
		return 0, err
	}

	if count == 0 {
		return 0, nil
	}
	return count, nil
}

func (s *AdminService) FileDeleteDecisionService(ctx context.Context, fileID string, action string) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	switch action {

	case "accept":
		fileKey, err := s.FileRepo.GetFileKeyfromtheFileID(ctx, fileID)
		if err != nil {
			if errors.Is(err, apperrors.ErrFileNotFound) {
				return apperrors.ErrFileNotFound
			}
			s.Logger.Errorw("failed to fetch file key", "file_id", fileID, "error", err)
			return apperrors.ErrInternal
		}

		if strings.TrimSpace(fileKey) == "" {
			s.Logger.Errorw("empty file key returned from repository", "file_id", fileID)
			return apperrors.ErrInternal
		}

		if err := s.Storage.Delete(ctx, fileKey); err != nil {
			s.Logger.Errorw("failed to delete file from storage", "file_key", fileKey, "error", err)
			return apperrors.ErrInternal
		}
		return s.FileRepo.DeleteFilePermanently(ctx, fileID)

	case "reject":
		return s.FileRepo.RejectDeleteRequest(ctx, fileID)

	default:
		return apperrors.ErrInvalidInput
	}
}

func (s *AdminService) PendingDeleteRequestCountService(ctx context.Context) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	count, err := s.FileRepo.PendingDeleteRequestCount(ctx)

	if err != nil {
		return 0, err
	}

	if count == 0 {
		return 0, nil
	}
	return count, nil
}

func (s *AdminService) RecentlyUploadedFileService(ctx context.Context) ([]filemodel.File, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return s.FileRepo.RecentlyUploadedFiles(ctx)
}

func (s *AdminService) GetPendingDeleteRequestFilesService(ctx context.Context) ([]filemodel.File, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return s.FileRepo.GetPendingDeleteRequestFiles(ctx)
}


func (s *AdminService) GetTotalFilesService(ctx context.Context) ([]filemodel.File, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return s.FileRepo.TotalFiles(ctx)
}

func (s *AdminService) DeleteFileService(ctx context.Context, fileID string) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	fileID = strings.TrimSpace(fileID)
	if fileID == "" {
		return apperrors.ErrInvalidID
	}

	if _, err := primitive.ObjectIDFromHex(fileID); err != nil {
		return apperrors.ErrInvalidID
	}

	s.Logger.Infow(
		"deleting file",
		"file_id", fileID,
	)

	fileKey, err := s.FileRepo.GetFileKeyfromtheFileID(ctx, fileID)
	if err != nil {
		if errors.Is(err, apperrors.ErrFileNotFound) {
			return apperrors.ErrFileNotFound
		}

		return fmt.Errorf("service: failed to fetch file key for file_id=%s: %w", fileID, err)
	}

	if fileKey == "" {
		s.Logger.Errorw(
			"empty file key returned from repository",
			"file_id", fileID,
		)
		return errors.New("internal error: empty file key")
	}

	err = s.Storage.Delete(ctx, fileKey)
	if err != nil {
		return fmt.Errorf("service: failed delete file for this key=%s: %w", fileKey, err)
	}

	err = s.FileRepo.DeleteFileRecord(ctx, fileID)

	if err != nil {
		return fmt.Errorf("Service:Error Deleting file from the db,Err:%v", err)
	}

	s.Logger.Infow("file deleted successfully", "file_id", fileID)

	return nil
}

func (s *AdminService) SigninService(ctx context.Context, req model.SigninPayload) (string,string,error) {

    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    s.Logger.Infof("Signin attempt | email=%s", req.Email)

    admin, err := s.AdminRepo.GetAdminByEmail(ctx, req.Email)
    if err != nil {
        if errors.Is(err, apperrors.ErrAdminNotFound) {
            return "","", apperrors.ErrAdminNotFound
        }
        return "","", fmt.Errorf("service: db lookup failed: %w", err)
    }

    if !utils.CheckPassword(req.Password, admin.Password) {
        return "","",apperrors.ErrInvalidPassword
    }

    accessToken, err := utils.GenerateAccessTokenForAdmin(admin.ID.Hex())
    if err != nil {
        return "", "",fmt.Errorf("service: failed generating access token: %w", err)
    }

    s.Logger.Infof("Signin successful | email=%s", req.Email)
    return accessToken, admin.Username,nil
}


func (s *AdminService) CreateAccountService(ctx context.Context,req model.AccoutCreationPayload) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := utils.ValidateAccountPayload(req); err != nil {
		return err
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.Password != "" {
		hashed, err := utils.HashPassword(req.Password)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}
		req.Password = hashed
	}

	s.Logger.Infof("Creating faculty account | email=%s", req.Email)

	admin := model.Admin{
		ID:        primitive.NewObjectID(),
		Username:  req.Name,
		Email:     req.Email,
		Password:  req.Password,
		CreatedAt: time.Now(),
	}
	err := s.AdminRepo.CreateAccount(ctx, admin)

	switch {
	case errors.Is(err, apperrors.ErrEmailAlreadyExists):
		return apperrors.ErrEmailAlreadyExists

	case err != nil:
		return fmt.Errorf("failed to create account: %w", err)
	}
	s.Logger.Infof("Account created successfully | email=%s", req.Email)

	return nil
}

func (s *AdminService) GetAvailableSubjects(ctx context.Context,) ([]subjects.Subject, error) {

	s.Logger.Infow("fetching available subjects")

	return subjects.AllSubjects(), nil
}
