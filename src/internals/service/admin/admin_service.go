package admin

import (
	"context"
	"strings"
	"time"

	domain "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/faculties"
	db "github.com/suhas-developer07/Kiosk-backend/src/internals/repository/admin"
	facultydb "github.com/suhas-developer07/Kiosk-backend/src/internals/repository/faculty_repo"
)

type AdminService struct {
	AdminRepo *db.AdminRepo
	FacultyRepo *facultydb.FacultyRepo
}

func NewAdminService(AdminRepo *db.AdminRepo,FacultyRepo *facultydb.FacultyRepo) *AdminService {
	return &AdminService{
		AdminRepo: AdminRepo,
		FacultyRepo: FacultyRepo,
	}
}

func (s *AdminService) GetFacultiesService(
	ctx context.Context,
) ([]domain.Faculty, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	faculties, err := s.FacultyRepo.GetFaculties(ctx)
	if err != nil {
		return nil, err
	}

	if faculties == nil {
		return []domain.Faculty{}, nil
	}

	return faculties, nil
}

func (s *AdminService) GetFacultiesByStreamService(
	ctx context.Context,
	stream string,
) ([]domain.Faculty, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	stream = strings.TrimSpace(strings.ToLower(stream))
	if stream == "" {
		return []domain.Faculty{}, nil
	}

	faculties, err := s.FacultyRepo.GetFacultiesByStream(ctx, stream)
	if err != nil {
		return nil, err
	}

	if faculties == nil {
		return []domain.Faculty{}, nil
	}

	return faculties, nil
}
