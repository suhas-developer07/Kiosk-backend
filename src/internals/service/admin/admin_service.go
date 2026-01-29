package admin

import (
	"context"
	"time"

	domain "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/faculties"
	facultydb "github.com/suhas-developer07/Kiosk-backend/src/internals/repository/faculty_repo"
	db "github.com/suhas-developer07/Kiosk-backend/src/internals/repository/admin"
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

func(s *AdminService) GetFacultiesService(ctx context.Context)([]domain.Faculty,error){
	ctx,cancel := context.WithTimeout(ctx,5*time.Second)
	defer cancel()

	Faculties,err := s.FacultyRepo.GetFaculties(ctx)

	if err != nil {
		return nil,err
	}

	return Faculties,nil
}


func(s *AdminService) GetFacultiesByStreamService(ctx context.Context,stream string)([]domain.Faculty,error){
	ctx,cancel := context.WithTimeout(ctx,5*time.Second)
	defer cancel()

	Faculties,err := s.FacultyRepo.GetFacultiesByStream(ctx,stream)

	if err != nil {
		return nil,err
	}

	return Faculties,nil
}

