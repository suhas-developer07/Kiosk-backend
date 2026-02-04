package admin

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	facultymodel "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/faculties"
	"github.com/suhas-developer07/Kiosk-backend/src/internals/domain/subjects"
	db "github.com/suhas-developer07/Kiosk-backend/src/internals/repository/admin"
	facultydb "github.com/suhas-developer07/Kiosk-backend/src/internals/repository/faculty_repo"
	filesdb "github.com/suhas-developer07/Kiosk-backend/src/internals/repository/files_repo"
	"github.com/suhas-developer07/Kiosk-backend/src/pkg/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type AdminService struct {
	AdminRepo   *db.AdminRepo
	FacultyRepo *facultydb.FacultyRepo
	FilesRepo   *filesdb.FilesRepo
	Logger      *zap.SugaredLogger
}

func NewAdminService(AdminRepo *db.AdminRepo, FacultyRepo *facultydb.FacultyRepo, FilesRepo *filesdb.FilesRepo, sugar *zap.SugaredLogger) *AdminService {
	return &AdminService{
		AdminRepo:   AdminRepo,
		FacultyRepo: FacultyRepo,
		FilesRepo: FilesRepo,
		Logger:      sugar,
	}
}

func (s *AdminService) GetFacultiesService(
	ctx context.Context,
) ([]facultymodel.Faculty, error) {

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

	count, err := s.FilesRepo.GetTotalFilesCount(ctx)

	if err != nil {
		return 0, err
	}

	if count == 0 {
		return 0, nil
	}
	return count, nil
}
