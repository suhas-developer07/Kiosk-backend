package Facultyservice

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	domain "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/faculties"
	db "github.com/suhas-developer07/Kiosk-backend/src/internals/repository/faculty_repo"
	"github.com/suhas-developer07/Kiosk-backend/src/pkg/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type FacultyService struct {
	FacultyRepo *db.FacultyRepo
	Logger      *zap.SugaredLogger
}

func NewFacultyService(repo *db.FacultyRepo, Logger *zap.SugaredLogger) *FacultyService {
	return &FacultyService{
		FacultyRepo: repo,
		Logger:      Logger,
	}
}
func (s *FacultyService) CreateAccountService(
	ctx context.Context,
	req domain.AccoutCreationPayload,
) error {

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

	faculty := domain.Faculty{
		ID:        primitive.NewObjectID(),
		Username:  req.Name,
		Email:     req.Email,
		Password:  req.Password,
		CreatedAt: time.Now(),
	}
	err := s.FacultyRepo.CreateAccount(ctx, faculty)

	switch {
	case errors.Is(err, domain.ErrEmailAlreadyExists):
		return domain.ErrEmailAlreadyExists

	case err != nil:
		return fmt.Errorf("failed to create account: %w", err)
	}
	s.Logger.Infof("Account created successfully | email=%s", req.Email)

	return nil
}
