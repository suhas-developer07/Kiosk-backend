package admin

import (
	db "github.com/suhas-developer07/Kiosk-backend/src/internals/repository/admin"
)

type AdminService struct {
	AdminRepo *db.AdminRepo
}

func NewFacultyService(repo *db.AdminRepo) *AdminService {
	return &AdminService{
		AdminRepo: repo,
	}
}

