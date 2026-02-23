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
	Repo   *repository.SuperAdminRepository
	logger *zap.SugaredLogger
}

func NewSuperAdminService(repo *repository.SuperAdminRepository, logger *zap.SugaredLogger) *SuperAdminService {
	return &SuperAdminService{
		Repo:   repo,
		logger: logger,
	}
}

/*
super admin auth services still not implemented
*/

func (s *SuperAdminService) CreateSuperAdmin(ctx context.Context, req model.SuperAdminCreateRequest) (error) {
	
	email := strings.ToLower(strings.TrimSpace(req.SuperAdminEmail))
	exists, err := s.Repo.CheckSuperAdminEmailExists(ctx, email)
	if err != nil {
		return  errors.New("unable to verify super admin email at the moment, please try again later")
	}
	if exists {
		return  errors.New("this email is already registered with another super admin")
	}

	hashedPassword, err := utils.HashPassword(req.SuperAdminPassword)
	if err != nil {
		return  errors.New("unable to process password, please try again")
	}

	superadminID := utils.GenerateUUID()
	superadmin := model.SuperAdmin{
		SuperAdminId:    superadminID,
		SuperAdminName:  req.SuperAdminName,
		SuperAdminEmail: email,
		SuperAdminPassword: hashedPassword,
		CreatedAt:       time.Now().Format(time.RFC3339),
	}

	if err := s.Repo.CreateSuperAdmin(ctx, superadmin); err != nil {
		return  errors.New("failed to create super admin, please try again")
	}

	return nil
}

func (s *SuperAdminService) LoginSuperAdminService(ctx context.Context,req model.SuperAdminLoginRequest) (string, string, string, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultOperationTimeout)
	defer cancel()

	req.SuperAdminEmail = strings.TrimSpace(strings.ToLower(req.SuperAdminEmail))

	s.logger.Infow("Super admin login attempt",
		"email", req.SuperAdminEmail,
	)

	super_admin, err := s.Repo.GetSuperAdminByEmail(ctx, req.SuperAdminEmail)
	if err != nil {
		if errors.Is(err, apperrors.ErrSuperAdminNotFound) {
			s.logger.Warnw("Login attempt for non-existent user",
				"email", req.SuperAdminEmail,
			)
			return "", "", "", apperrors.ErrSuperAdminNotFound
		}
		s.logger.Errorw("Database error during user lookup",
			"email", req.SuperAdminEmail,
			"error", err,
		)
		return "", "", "", fmt.Errorf("Super admin lookup failed: %w", err)
	}

	if !utils.CheckPassword(req.SuperAdminPassword, super_admin.SuperAdminPassword) {
		s.logger.Warnw("Invalid password attempt",
			"email", req.SuperAdminEmail,
		)
		return "", "", "", apperrors.ErrInvalidPassword
	}

	accessToken, err := utils.GenerateAccessTokenForSuperAdmin(super_admin.SuperAdminId)
	if err != nil {
		s.logger.Errorw("Failed to generate access token",
			"super_admin_id", super_admin.SuperAdminId,
			"email", req.SuperAdminEmail,
			"error", err,
		)
		return "", "", "", fmt.Errorf("token generation failed: %w", err)
	}

	s.logger.Infow("Super admin logged in successfully",
		"super_admin_id", super_admin.SuperAdminId,
		"email", req.SuperAdminEmail,
	)

	return accessToken, super_admin.SuperAdminEmail, super_admin.SuperAdminName, nil
}


func (s *SuperAdminService) CreateCollege(ctx context.Context, req model.CollegeCreateRequest, SuperadminID string) (*model.CollegeResponse, error) {

	email := strings.ToLower(strings.TrimSpace(req.CollegeEmail))
	exists, err :=  s.Repo.CheckCollegeEmailExists(ctx, email)
	if err != nil {
		return nil, errors.New("unable to verify college email at the moment, please try again later")
	}
	if exists {
		return nil, errors.New("this email is already registered with another college")
	}

	hashedPassword, err := utils.HashPassword(req.CollegePassword)
	if err != nil {
		return nil, errors.New("unable to process password, please try again")
	}

	collegeID := utils.GenerateUUID()
	college := model.SuperAdminCollege{
		SuperAdminId:    SuperadminID,
		CollegeID:       collegeID,
		CollegeName:     req.CollegeName,
		CollegeEmail:    email,
		CollegePhone:    req.CollegePhone,
		CollegePassword: hashedPassword,
		CollegeAddress:  req.CollegeAddress,
		Balance:         "0",
		CreatedAt:       time.Now().Format(time.RFC3339),
	}

	if err := s.Repo.CreateCollege(ctx, college); err != nil {
		return nil, errors.New("failed to create college, please try again")
	}

	return &model.CollegeResponse{
		SuperAdminId: SuperadminID,
		CollegeID:    collegeID,
		CollegeName:  req.CollegeName,
		Balance:      college.Balance,
	}, nil
}

func (s *SuperAdminService) CollegeLogin(ctx context.Context, req model.CollegeLoginRequest) (*model.CollegeTokenResponse, error) {

	collegeID, name, hashedPassword, superadminID, err := s.Repo.GetCollegeForLogin(ctx, req.CollegeEmail)
	if err != nil {
		return nil, errors.New("college account not found")
	}

	if !utils.CheckPassword(req.CollegePassword, hashedPassword) {
		return nil, apperrors.ErrInvalidPassword
	}

	balance, err := s.Repo.GetCollegeBalance(ctx, collegeID)
	if err != nil {
		return nil, errors.New("unable to retrieve account balance at the moment")
	}

	token, err := utils.GenerateAccessTokenForCollegeLogin(req.CollegeEmail, name, collegeID, superadminID)
	if err != nil {
		return nil, errors.New("failed to log in, please try again")
	}

	return &model.CollegeTokenResponse{
		Token:   token,
		Balance: balance,
	}, nil
}

func (s *SuperAdminService) GetCollegesBySuperAdminID(ctx context.Context, adminID string) ([]model.SuperAdminCollege, error) {
	if adminID == "" {
		return nil, errors.New("admin ID is required")
	}
	colleges, err := s.Repo.GetCollegesBySuperAdminID(ctx, adminID)
	if err != nil {
		return nil, errors.New("unable to fetch colleges at this time")
	}

	return colleges, nil
}

func (s *SuperAdminService) GetCollegeDetails(ctx context.Context, collegeID string) (*model.SuperAdminCollege, error) {
	return s.Repo.GetCollegeByID(ctx, collegeID)
}

func (s *SuperAdminService) DeleteCollege(ctx context.Context, collegeID string) error {
	if collegeID == "" {
		return errors.New("college ID is required")
	}

	college, err := s.Repo.GetCollegeByID(ctx, collegeID)
	if err != nil {
		return errors.New("unable to verify college details")
	}
	if college == nil {
		return errors.New("college not found")
	}

	if err := s.Repo.DeleteCollege(ctx, collegeID); err != nil {
		return errors.New("failed to delete college, please try again")
	}

	return nil
}

func (s *SuperAdminService) RechargeToCollege(ctx context.Context,req model.CollegeRechargeRequest) error {
	ctx,cancel := context.WithTimeout(ctx, defaultOperationTimeout)
	defer cancel()

	//if err := s.valid
	rechargeAmt,err := strconv.Atoi(req.RechargeAmount)
	if err != nil || rechargeAmt <= 0 {
		s.logger.Warnw("Recharge amount must be an Positive whole number","recharge_amount",req.RechargeAmount)
		return errors.New("recharge amount must be greater than zero")
	}

	collegeBalStr, err := s.Repo.GetCollegeBalance(ctx, req.CollegeID)
	if err != nil {
		return errors.New("unable to retrieve college balance")
	}
	collegeBal, err := strconv.Atoi(collegeBalStr)
	if err != nil {
		return errors.New("college balance data is invalid")
	}
	
	newCollegeBalance := strconv.Itoa(collegeBal + rechargeAmt)
	
	if err := s.Repo.UpdateCollegeBalance(ctx, req.CollegeID, newCollegeBalance); err != nil {
		return errors.New("failed to update college balance. Please try again")
	}

	if err := s.recordCollegeRechargeHistory(ctx, req); err != nil {
		s.logger.Errorw("Failed to record college recharge history",
			"college_id", req.CollegeID,
			"recharge_amount", req.RechargeAmount,
			"error", err,
		)
		//Rollback balance update on history insertion failure
		_ = s.Repo.UpdateCollegeBalance(ctx, req.CollegeID, collegeBalStr)
		return errors.New("failed to record recharge history. Please try again")
	}

	return nil
}

func (s *SuperAdminService) GetRechargeHistory(ctx context.Context, collegeID string) ([]model.CollegeRechargeHistory, error) {
	if collegeID == "" {
		return nil, errors.New("college ID is required")
	}

	recharges, err := s.Repo.GetRechargeHistoryByCollegeID(ctx, collegeID)
	if err != nil {
		return nil, errors.New("unable to retrieve recharge history")
	}

	return recharges, nil
}

func (s *SuperAdminService) GetSuperAdminBalance(ctx context.Context, superAdminID string) (string, error) {
	if superAdminID == "" {
		return "0", errors.New("super admin ID is required")
	}

	balance, err := s.Repo.GetSuperAdminBalance(ctx, superAdminID)
	if err != nil {
		return "0", err
	}

	return balance, nil
}


/* Machine management services */

func (s *SuperAdminService) CreateMachineService(ctx context.Context, req model.MachineCreateRequest,superadminId string) error {
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
	)

	machineID := utils.GenerateUUID()

	machine := model.Machine{
		MachineID:   machineID,
		MachineNo:   req.MachineNo,
		MachineName: req.MachineName,
		CollegeId: req.CollegeId,
		SuperAdminId: superadminId,
		Balance:     initialMachineBalance,
	}

	if err := s.Repo.CreateMachine(ctx, machine); err != nil {
		if errors.Is(err, apperrors.ErrMachineAlreadyExists) {
			s.logger.Warnw("Attempted to create machine with existing number",
				"machine_no", req.MachineNo,
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

	machines, err := s.Repo.GetMachinesByCollegeID(ctx, collegeID)
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

func (s *SuperAdminService) recordCollegeRechargeHistory(ctx context.Context,req model.CollegeRechargeRequest) error {

	loc, err := time.LoadLocation(timezoneLocation)
	if err != nil {
		s.logger.Errorw("Failed to load timezone for recharge history",
			"timezone", timezoneLocation,
			"error", err,
		)
		return errors.New("timezone configuration error")
	}

	now := time.Now().In(loc)

	// Build history record
	history := model.CollegeRechargeHistory{
		RechargeID: utils.GenerateUUID(),
		CollegeID:      req.CollegeID,
		SuperAdminId: req.SuperAdminId,
		RechargeAmount: req.RechargeAmount,
		Date:           now.Format(dateFormat),
		Time:           now.Format(timeFormat),
	}

	// Insert into database
	if err := s.Repo.InsertCollegeRechargeHistory(ctx, history); err != nil {
		s.logger.Errorw("Failed to insert college recharge history",
			"college_id", req.CollegeID,
			"error", err,
		)
		return err
	}

	return nil
}
