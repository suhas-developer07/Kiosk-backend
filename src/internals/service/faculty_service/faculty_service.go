package Facultyservice

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	domain "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/faculties"
	"github.com/suhas-developer07/Kiosk-backend/src/internals/domain/subjects"
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

func (s *FacultyService) SigninService(ctx context.Context, req domain.SigninPayload) (string, string, error) {

    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    s.Logger.Infof("Signin attempt | email=%s", req.Email)

    faculty, err := s.FacultyRepo.GetFacultyByEmail(ctx, req.Email)
    if err != nil {
        if errors.Is(err, domain.ErrUserNotFound) {
            return "", "", domain.ErrUserNotFound
        }
        return "", "", fmt.Errorf("service: db lookup failed: %w", err)
    }

    if !utils.CheckPassword(req.Password, faculty.Password) {
        return "", "", domain.ErrInvalidPassword
    }

    accessToken, err := utils.GenerateAccessToken(faculty.ID.Hex())
    if err != nil {
        return "", "", fmt.Errorf("service: failed generating access token: %w", err)
    }

    refreshToken, err := utils.GenerateRefreshToken(faculty.ID.Hex())
    if err != nil {
        return "", "", fmt.Errorf("service: failed generating refresh token: %w", err)
    }

    s.Logger.Infof("Signin successful | email=%s", req.Email)
    return accessToken, refreshToken, nil
}

func (s *FacultyService) UpdateProfileService(
	ctx context.Context,
	FacultyId string,
	req domain.UpdateProfilePayload,
) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	s.Logger.Infof("Updating faculty profile | faculty_id=%s", FacultyId)

	objectID, err := primitive.ObjectIDFromHex(FacultyId)
	if err != nil {
		return domain.ErrInvalidID
	}

	//TODO : I need to check faculty id in id database.
	profile := domain.FacultyProfile{
		FacultyID:     FacultyId,
		Subjects:      req.Subjects,
		Gender:        req.Gender,
		Qualification: req.Qualification,
		Experience:    req.Experience,
		PhoneNumber:   req.PhoneNumber,
	}

	err = s.FacultyRepo.UpdateProfile(ctx, objectID, profile)
	if err != nil {

		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			return domain.ErrUserNotFound

		default:
			return fmt.Errorf("service: update profile failed: %w", err)
		}
	}

	s.Logger.Infof("Profile updated successfully | faculty_id=%s", FacultyId)
	return nil
}

func (s *FacultyService) GetSubjectsByFacultyID(
	ctx context.Context,
	facultyID string,
) ([]subjects.Subject, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	facultyID = strings.TrimSpace(facultyID)
	if facultyID == "" {
		return nil, domain.ErrInvalidID
	}

	objID, err := primitive.ObjectIDFromHex(facultyID)
	if err != nil {
		return nil, domain.ErrInvalidID
	}

	s.Logger.Infow(
		"fetching subjects for faculty",
		"faculty_id", objID.Hex(),
	)

	profile, err := s.FacultyRepo.GetFacultyProfileByID(ctx, objID)
	if err != nil {

		if errors.Is(err, domain.ErrFacultyNotFound) {
			return nil, domain.ErrFacultyNotFound
		}

		return nil, fmt.Errorf(
			"service: failed to fetch faculty profile: %w",
			err,
		)
	}

	if len(profile.Subjects) == 0 {
		return []subjects.Subject{}, nil
	}

	validSubjects := make([]subjects.Subject, 0, len(profile.Subjects))
	for _, sub := range profile.Subjects {
		if subjects.IsValidSubject(string(sub)) {
			validSubjects = append(validSubjects, sub)
		} else {
			s.Logger.Warnw(
				"invalid subject found in faculty profile",
				"faculty_id", facultyID,
				"subject", sub,
			)
		}
	}

	return validSubjects, nil
}

func (s *FacultyService) GetAvailableSubjects(
	ctx context.Context,
) ([]subjects.Subject, error) {

	s.Logger.Infow("fetching available subjects")

	return subjects.AllSubjects(), nil
}

