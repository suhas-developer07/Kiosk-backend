package paymentsystem

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	apperrors "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/errors"
	model "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/payment_system"
	repository "github.com/suhas-developer07/Kiosk-backend/src/internals/repository/payment_system"
	"github.com/suhas-developer07/Kiosk-backend/src/pkg/utils"
	"go.uber.org/zap"
)

type SuperAdminService struct {
	repo   *repository.SuperAdminRepository
	logger *zap.SugaredLogger
}

func NewSuperAdminService(
	repo *repository.SuperAdminRepository,
	logger *zap.SugaredLogger,
) *SuperAdminService {
	return &SuperAdminService{
		repo:   repo,
		logger: logger,
	}
}

/*  Super Admin Auth Services  */

func (s *SuperAdminService) CreateSuperAdmin(ctx context.Context, req model.SuperAdminCreateRequest) error {
	ctx, cancel := context.WithTimeout(ctx, defaultOperationTimeout)
	defer cancel()

	email := strings.ToLower(strings.TrimSpace(req.SuperAdminEmail))

	exists, err := s.repo.CheckSuperAdminEmailExists(ctx, email)
	if err != nil {
		s.logger.Errorw("Failed to check super admin email existence",
			"email", email,
			"error", err,
		)
		return errors.New("unable to verify super admin email at the moment. Please try again later")
	}
	if exists {
		s.logger.Warnw("Attempted to register with existing super admin email",
			"email", email,
		)
		return apperrors.ErrEmailAlreadyExists
	}

	hashedPassword, err := utils.HashPassword(req.SuperAdminPassword)
	if err != nil {
		s.logger.Errorw("Failed to hash password during super admin creation",
			"email", email,
			"error", err,
		)
		return errors.New("unable to process password. Please try again")
	}

	superAdminID := utils.GenerateUUID()
	superAdmin := model.SuperAdmin{
		SuperAdminId:       superAdminID,
		SuperAdminName:     req.SuperAdminName,
		SuperAdminEmail:    email,
		SuperAdminPassword: hashedPassword,
		CreatedAt:          time.Now().UTC().Format(time.RFC3339),
	}

	if err := s.repo.CreateSuperAdmin(ctx, superAdmin); err != nil {
		s.logger.Errorw("Failed to persist super admin to database",
			"email", email,
			"error", err,
		)
		return errors.New("failed to create super admin. Please try again")
	}

	s.logger.Infow("Super admin created successfully",
		"super_admin_id", superAdminID,
		"email", email,
	)

	return nil
}

func (s *SuperAdminService) LoginSuperAdminService(ctx context.Context, req model.SuperAdminLoginRequest) (token, email, username string, err error) {
	ctx, cancel := context.WithTimeout(ctx, defaultOperationTimeout)
	defer cancel()

	req.SuperAdminEmail = strings.ToLower(strings.TrimSpace(req.SuperAdminEmail))

	s.logger.Infow("Super admin login attempt",
		"email", req.SuperAdminEmail,
	)

	superAdmin, err := s.repo.GetSuperAdminByEmail(ctx, req.SuperAdminEmail)
	if err != nil {
		if errors.Is(err, apperrors.ErrSuperAdminNotFound) {
			s.logger.Warnw("Login attempt for non-existent super admin",
				"email", req.SuperAdminEmail,
			)
			return "", "", "", apperrors.ErrSuperAdminNotFound
		}
		s.logger.Errorw("Database error during super admin lookup",
			"email", req.SuperAdminEmail,
			"error", err,
		)
		return "", "", "", fmt.Errorf("super admin lookup failed: %w", err)
	}

	if !utils.CheckPassword(req.SuperAdminPassword, superAdmin.SuperAdminPassword) {
		s.logger.Warnw("Invalid password attempt for super admin",
			"email", req.SuperAdminEmail,
		)
		return "", "", "", apperrors.ErrInvalidPassword
	}

	accessToken, err := utils.GenerateAccessTokenForSuperAdmin(superAdmin.SuperAdminId)
	if err != nil {
		s.logger.Errorw("Failed to generate access token for super admin",
			"super_admin_id", superAdmin.SuperAdminId,
			"email", req.SuperAdminEmail,
			"error", err,
		)
		return "", "", "", fmt.Errorf("token generation failed: %w", err)
	}

	s.logger.Infow("Super admin logged in successfully",
		"super_admin_id", superAdmin.SuperAdminId,
		"email", req.SuperAdminEmail,
	)

	return accessToken, superAdmin.SuperAdminEmail, superAdmin.SuperAdminName, nil
}

/*  College Management Services  */

func (s *SuperAdminService) CreateCollege(ctx context.Context, req model.CollegeCreateRequest, superAdminID string) (*model.CollegeResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultOperationTimeout)
	defer cancel()

	email := strings.ToLower(strings.TrimSpace(req.CollegeEmail))

	exists, err := s.repo.CheckCollegeEmailExists(ctx, email)
	if err != nil {
		s.logger.Errorw("Failed to check college email existence",
			"email", email,
			"error", err,
		)
		return nil, errors.New("unable to verify college email at the moment. Please try again later")
	}
	if exists {
		s.logger.Warnw("Attempted to register with existing college email",
			"email", email,
		)
		return nil, errors.New("this email is already registered with another college")
	}

	hashedPassword, err := utils.HashPassword(req.CollegePassword)
	if err != nil {
		s.logger.Errorw("Failed to hash password during college creation",
			"email", email,
			"error", err,
		)
		return nil, errors.New("unable to process password. Please try again")
	}

	collegeID := utils.GenerateUUID()
	college := model.SuperAdminCollege{
		SuperAdminId:    superAdminID,
		CollegeID:       collegeID,
		CollegeName:     req.CollegeName,
		CollegeEmail:    email,
		CollegePhone:    req.CollegePhone,
		CollegePassword: hashedPassword,
		CollegeAddress:  req.CollegeAddress,
		Balance:         "0",
		CreatedAt:       time.Now().UTC().Format(time.RFC3339),
	}

	if err := s.repo.CreateCollege(ctx, college); err != nil {
		s.logger.Errorw("Failed to persist college to database",
			"email", email,
			"super_admin_id", superAdminID,
			"error", err,
		)
		return nil, errors.New("failed to create college. Please try again")
	}

	s.logger.Infow("College created successfully",
		"college_id", collegeID,
		"super_admin_id", superAdminID,
	)

	return &model.CollegeResponse{
		SuperAdminId: superAdminID,
		CollegeID:    collegeID,
		CollegeName:  req.CollegeName,
		Balance:      college.Balance,
	}, nil
}

func (s *SuperAdminService) GetCollegesBySuperAdminID(ctx context.Context, adminID string) ([]model.SuperAdminCollege, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultOperationTimeout)
	defer cancel()

	if adminID == "" {
		return nil, errors.New("admin ID is required")
	}

	colleges, err := s.repo.GetCollegesBySuperAdminID(ctx, adminID)
	if err != nil {
		s.logger.Errorw("Failed to fetch colleges for super admin",
			"super_admin_id", adminID,
			"error", err,
		)
		return nil, errors.New("unable to fetch colleges at this time")
	}

	return colleges, nil
}

func (s *SuperAdminService) GetCollegeDetails(ctx context.Context, collegeID string) (*model.SuperAdminCollege, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultOperationTimeout)
	defer cancel()

	if collegeID == "" {
		return nil, errors.New("college ID is required")
	}

	college, err := s.repo.GetCollegeByID(ctx, collegeID)
	if err != nil {
		s.logger.Errorw("Failed to fetch college details",
			"college_id", collegeID,
			"error", err,
		)
		return nil, err
	}

	return college, nil
}

func (s *SuperAdminService) DeleteCollege(ctx context.Context, collegeID string) error {
	ctx, cancel := context.WithTimeout(ctx, defaultOperationTimeout)
	defer cancel()

	if collegeID == "" {
		return errors.New("college ID is required")
	}

	college, err := s.repo.GetCollegeByID(ctx, collegeID)
	if err != nil {
		s.logger.Errorw("Failed to verify college before deletion",
			"college_id", collegeID,
			"error", err,
		)
		return err
	}
	if college == nil {
		return apperrors.ErrCollegeNotFound
	}

	if err := s.repo.DeleteCollege(ctx, collegeID); err != nil {
		s.logger.Errorw("Failed to delete college",
			"college_id", collegeID,
			"error", err,
		)
		return errors.New("failed to delete college. Please try again")
	}

	s.logger.Infow("College deleted successfully",
		"college_id", collegeID,
	)

	return nil
}

func (s *SuperAdminService) RechargeToCollege(ctx context.Context, req model.CollegeRechargeRequest) error {
	ctx, cancel := context.WithTimeout(ctx, defaultOperationTimeout)
	defer cancel()

	rechargeAmt, err := strconv.Atoi(req.RechargeAmount)
	if err != nil || rechargeAmt <= 0 {
		s.logger.Warnw("Invalid recharge amount for college",
			"college_id", req.CollegeID,
			"recharge_amount", req.RechargeAmount,
		)
		return errors.New("recharge amount must be a positive whole number")
	}

	collegeBalStr, err := s.repo.GetCollegeBalance(ctx, req.CollegeID)
	if err != nil {
		s.logger.Errorw("Failed to retrieve college balance before recharge",
			"college_id", req.CollegeID,
			"error", err,
		)
		return errors.New("unable to retrieve college balance")
	}

	collegeBal, err := strconv.Atoi(collegeBalStr)
	if err != nil {
		s.logger.Errorw("College balance is in invalid format",
			"college_id", req.CollegeID,
			"balance", collegeBalStr,
		)
		return errors.New("college balance data is invalid. Please contact support")
	}

	newBalance := strconv.Itoa(collegeBal + rechargeAmt)

	if err := s.repo.UpdateCollegeBalance(ctx, req.CollegeID, newBalance); err != nil {
		s.logger.Errorw("Failed to update college balance",
			"college_id", req.CollegeID,
			"new_balance", newBalance,
			"error", err,
		)
		return errors.New("failed to update college balance. Please try again")
	}

	if err := s.recordCollegeRechargeHistory(ctx, req); err != nil {
		s.logger.Errorw("Failed to record recharge history, attempting rollback",
			"college_id", req.CollegeID,
			"recharge_amount", req.RechargeAmount,
			"error", err,
		)
		// Rollback balance update on history insertion failure.
		_ = s.repo.UpdateCollegeBalance(ctx, req.CollegeID, collegeBalStr)
		return errors.New("failed to record recharge history. Transaction rolled back")
	}

	s.logger.Infow("College recharged successfully",
		"college_id", req.CollegeID,
		"super_admin_id", req.SuperAdminId,
		"amount", rechargeAmt,
		"new_balance", newBalance,
	)

	return nil
}

func (s *SuperAdminService) GetRechargeHistory(ctx context.Context, collegeID string) ([]model.CollegeRechargeHistory, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultOperationTimeout)
	defer cancel()

	if collegeID == "" {
		return nil, errors.New("college ID is required")
	}

	recharges, err := s.repo.GetRechargeHistoryByCollegeID(ctx, collegeID)
	if err != nil {
		s.logger.Errorw("Failed to retrieve college recharge history",
			"college_id", collegeID,
			"error", err,
		)
		return nil, errors.New("unable to retrieve recharge history at this time")
	}

	return recharges, nil
}

func (s *SuperAdminService) GetTotalCollegesCount(ctx context.Context) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultOperationTimeout)
	defer cancel()

	count, err := s.repo.GetTotalCollegesCount(ctx)
	if err != nil {
		s.logger.Errorw("Failed to retrieve total colleges count",
			"error", err,
		)
		return 0, errors.New("unable to retrieve colleges count at this time")
	}

	return count, nil
}

func (s *SuperAdminService) GetTotalRechargeVolume(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultOperationTimeout)
	defer cancel()

	volume, err := s.repo.GetTotalRechargeVolume(ctx)
	if err != nil {
		s.logger.Errorw("Failed to retrieve total recharge volume",
			"error", err,
		)
		return "", errors.New("unable to retrieve recharge volume at this time")
	}

	return volume, nil
}


func(s *SuperAdminService) GetTotalRechargeVolumeByCollege(ctx context.Context, collegeID string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultOperationTimeout)
	defer cancel()

	if collegeID == "" {
		return "", errors.New("college ID is required")
	}

	volume, err := s.repo.GetTotalRechargeVolumeByCollege(ctx, collegeID)
	if err != nil {
		s.logger.Errorw("Failed to retrieve total recharge volume for college",
			"college_id", collegeID,
			"error", err,
		)
		return "", errors.New("unable to retrieve recharge volume for college at this time")
	}

	return volume, nil
}

func(s *SuperAdminService) GetOverallCollegeRechargeHistory(ctx context.Context) ([]model.CollegeRechargeHistory, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultOperationTimeout)
	defer cancel()

	history, err := s.repo.GetOverallCollegeRechargeHistory(ctx)
	if err != nil {
		s.logger.Errorw("Failed to retrieve overall college recharge history",
			"error", err,
		)
		return nil, errors.New("unable to retrieve overall recharge history at this time")
	}

	return history, nil
}
/*  Machine Management Services  */

func (s *SuperAdminService) CreateMachineService(ctx context.Context, req model.MachineCreateRequest, superAdminID string) error {
	ctx, cancel := context.WithTimeout(ctx, defaultOperationTimeout)
	defer cancel()

	req.MachineNo = strings.TrimSpace(req.MachineNo)
	req.MachineName = strings.TrimSpace(req.MachineName)

	if req.MachineNo == "" || req.MachineName == "" {
		s.logger.Warnw("Missing required fields for machine creation",
			"machine_no", req.MachineNo,
			"machine_name", req.MachineName,
		)
		return apperrors.ErrInvalidInput
	}

	s.logger.Infow("Creating new machine",
		"machine_no", req.MachineNo,
		"machine_name", req.MachineName,
		"super_admin_id", superAdminID,
	)

	machineID := utils.GenerateUUID()
	machine := model.Machine{
		MachineID:    machineID,
		MachineNo:    req.MachineNo,
		MachineName:  req.MachineName,
		CollegeId:    req.CollegeId,
		SuperAdminId: superAdminID,
		Balance:      initialMachineBalance,
	}

	if err := s.repo.CreateMachine(ctx, machine); err != nil {
		if errors.Is(err, apperrors.ErrMachineAlreadyExists) {
			s.logger.Warnw("Attempted to create machine with existing number",
				"machine_no", req.MachineNo,
				"college_id", req.CollegeId,
			)
			return apperrors.ErrMachineAlreadyExists
		}
		s.logger.Errorw("Failed to create machine in database",
			"machine_no", req.MachineNo,
			"error", err,
		)
		return fmt.Errorf("machine creation failed: %w", err)
	}

	s.logger.Infow("Machine created successfully",
		"machine_id", machineID,
		"machine_no", req.MachineNo,
		"super_admin_id", superAdminID,
	)

	return nil
}

func (s *SuperAdminService) GetMachinesByCollegeID(ctx context.Context, collegeID string) ([]model.Machine, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultOperationTimeout)
	defer cancel()

	if collegeID == "" {
		s.logger.Warn("College ID is required to fetch machines")
		return nil, errors.New("college ID is required")
	}

	machines, err := s.repo.GetMachinesByCollegeID(ctx, collegeID)
	if err != nil {
		if errors.Is(err, apperrors.ErrCollegeNotFound) {
			s.logger.Warnw("College not found while fetching machines",
				"college_id", collegeID,
			)
			return nil, apperrors.ErrCollegeNotFound
		}
		s.logger.Errorw("Failed to fetch machines for college",
			"college_id", collegeID,
			"error", err,
		)
		return nil, err
	}

	return machines, nil
}

func (s *SuperAdminService) GetTotalMachinesCountByCollege(ctx context.Context, collegeID string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultOperationTimeout)
	defer cancel()

	if collegeID == "" {
		s.logger.Warn("College ID is required to fetch machines count")
		return 0, errors.New("college ID is required")
	}

	count, err := s.repo.GetTotalMachinesCountByCollege(ctx, collegeID)
	if err != nil {
		if errors.Is(err, apperrors.ErrCollegeNotFound) {
			s.logger.Warnw("College not found while fetching machines count",
				"college_id", collegeID,
			)
			return 0, apperrors.ErrCollegeNotFound
		}
		s.logger.Errorw("Failed to fetch machines count for college",
			"college_id", collegeID,
			"error", err,
		)
		return 0, err
	}

	return count, nil
}

func (s *SuperAdminService) GetTotalMachinesCount(ctx context.Context) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultOperationTimeout)
	defer cancel()

	count, err := s.repo.GetTotalMachinesCount(ctx)
	if err != nil {
		s.logger.Errorw("Failed to retrieve total machines count",
			"error", err,
		)
		return 0, errors.New("unable to retrieve machines count at this time")
	}

	return count, nil
}

/*  Private Helper Methods  */

func (s *SuperAdminService) recordCollegeRechargeHistory(ctx context.Context, req model.CollegeRechargeRequest) error {
	loc, err := time.LoadLocation(timezoneLocation)
	if err != nil {
		s.logger.Errorw("Failed to load timezone for recharge history",
			"timezone", timezoneLocation,
			"error", err,
		)
		return errors.New("timezone configuration error")
	}

	now := time.Now().In(loc)

	history := model.CollegeRechargeHistory{
		RechargeID:     utils.GenerateUUID(),
		CollegeID:      req.CollegeID,
		SuperAdminId:   req.SuperAdminId,
		RechargeAmount: req.RechargeAmount,
		Date:           now.Format(dateFormat),
		Time:           now.Format(timeFormat),
	}

	if err := s.repo.InsertCollegeRechargeHistory(ctx, history); err != nil {
		s.logger.Errorw("Failed to insert college recharge history",
			"college_id", req.CollegeID,
			"error", err,
		)
		return err
	}

	return nil
}
