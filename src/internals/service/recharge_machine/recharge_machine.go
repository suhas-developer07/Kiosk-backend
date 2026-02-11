package rechargemachine

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	apperrors "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/errors"
	model "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/recharge_machine"
	repo "github.com/suhas-developer07/Kiosk-backend/src/internals/repository/recharge_machine_repo"
	"github.com/suhas-developer07/Kiosk-backend/src/pkg/utils"
	"go.uber.org/zap"
)

const (
	defaultOperationTimeout = 5 * time.Second
	timezoneLocation        = "Asia/Kolkata"
	initialMachineBalance   = "0"
	dateFormat              = "2006-01-02"
	timeFormat              = "15:04:05"
)

// RechargeMachineService handles business logic for recharge machine operations
type RechargeMachineService struct {
	repo   *repo.RechargeMachineRepo
	logger *zap.SugaredLogger
}

// NewRechargeMachineService creates a new instance of RechargeMachineService
func NewRechargeMachineService(
	rechargeMachineRepo *repo.RechargeMachineRepo,
	logger *zap.SugaredLogger,
) *RechargeMachineService {
	return &RechargeMachineService{
		repo:   rechargeMachineRepo,
		logger: logger,
	}
}

// CreateAccountService creates a new main admin account
func (s *RechargeMachineService) CreateAccountService(ctx context.Context, req model.CreateMainAdminPayload) error {
	ctx, cancel := context.WithTimeout(ctx, defaultOperationTimeout)
	defer cancel()

	// Sanitize email input
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Name = strings.TrimSpace(req.Name)

	// Validate required fields
	if req.Email == "" {
		s.logger.Warnw("Empty email provided for account creation")
		return apperrors.ErrInvalidInput
	}

	if req.Password == "" {
		s.logger.Warnw("Empty password provided for account creation",
			"email", req.Email,
		)
		return apperrors.ErrInvalidPassword
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		s.logger.Errorw("Failed to hash password during account creation",
			"email", req.Email,
			"error", err,
		)
		return fmt.Errorf("password hashing failed: %w", err)
	}
	req.Password = hashedPassword

	s.logger.Infow("Creating admin account",
		"email", req.Email,
		"username", req.Name,
	)

	// Generate unique ID
	adminID := utils.GenerateUUID()

	// Build admin entity
	admin := model.MainAdmin{
		ID:        adminID,
		Username:  req.Name,
		Email:     req.Email,
		Password:  req.Password,
		CreatedAt: time.Now().UTC(),
	}

	// Persist to database
	if err := s.repo.CreateAccount(ctx, admin); err != nil {
		if errors.Is(err, apperrors.ErrEmailAlreadyExists) {
			s.logger.Warnw("Attempted to create account with existing email",
				"email", req.Email,
			)
			return apperrors.ErrEmailAlreadyExists
		}
		s.logger.Errorw("Failed to create admin account in database",
			"email", req.Email,
			"error", err,
		)
		return fmt.Errorf("account creation failed: %w", err)
	}

	s.logger.Infow("Admin account created successfully",
		"admin_id", adminID,
		"email", req.Email,
	)

	return nil
}

// CreateMachineService creates a new machine
func (s *RechargeMachineService) CreateMachineService(ctx context.Context, req model.MachineCreateRequest) error {
	ctx, cancel := context.WithTimeout(ctx, defaultOperationTimeout)
	defer cancel()

	// Sanitize inputs
	req.MachineNo = strings.TrimSpace(req.MachineNo)
	req.MachineName = strings.TrimSpace(req.MachineName)

	// Validate required fields
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

	// Generate unique machine ID
	machineID := utils.GenerateUUID()

	// Build machine entity
	machine := model.Machine{
		MachineID:   machineID,
		MachineNo:   req.MachineNo,
		MachineName: req.MachineName,
		Balance:     initialMachineBalance,
	}

	// Persist to database
	if err := s.repo.CreateMachine(ctx, machine); err != nil {
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

// RechargeMachine handles adding balance to a machine
func (s *RechargeMachineService) RechargeMachine(ctx context.Context, req model.MachineRechargeRequest) error {
	ctx, cancel := context.WithTimeout(ctx, defaultOperationTimeout)
	defer cancel()

	// Validate recharge amount format
	if err := s.validateRechargeAmount(req.RechargeAmount); err != nil {
		s.logger.Warnw("Invalid recharge amount provided",
			"machine_id", req.MachineID,
			"amount", req.RechargeAmount,
			"error", err,
		)
		return err
	}

	rechargeAmt, _ := strconv.Atoi(req.RechargeAmount)

	s.logger.Infow("Processing machine recharge",
		"machine_id", req.MachineID,
		"amount", rechargeAmt,
	)

	// Get current machine balance
	currentBalance, err := s.repo.GetMachineBalance(ctx, req.MachineID)
	if err != nil {
		s.logger.Errorw("Failed to fetch machine balance for recharge",
			"machine_id", req.MachineID,
			"error", err,
		)
		return errors.New("machine not found or balance unavailable")
	}

	// Parse current balance
	currentBalanceInt, err := strconv.Atoi(currentBalance)
	if err != nil {
		s.logger.Errorw("Machine balance is in invalid format",
			"machine_id", req.MachineID,
			"balance", currentBalance,
			"error", err,
		)
		return errors.New("machine balance is corrupted. Please contact support")
	}

	// Calculate new balance
	newBalance := currentBalanceInt + rechargeAmt
	newBalanceStr := strconv.Itoa(newBalance)

	// Update machine balance
	if err := s.repo.UpdateMachineBalance(ctx, req.MachineID, newBalanceStr); err != nil {
		s.logger.Errorw("Failed to update machine balance",
			"machine_id", req.MachineID,
			"new_balance", newBalance,
			"error", err,
		)
		return errors.New("failed to update machine balance. Please try again")
	}

	// Record transaction in history
	if err := s.recordMachineRechargeHistory(ctx, req, currentBalance); err != nil {
		// Rollback balance update on history insertion failure
		s.logger.Errorw("Failed to record recharge history, attempting rollback",
			"machine_id", req.MachineID,
			"error", err,
		)
		_ = s.repo.UpdateMachineBalance(ctx, req.MachineID, currentBalance)
		return errors.New("failed to record recharge history. Transaction rolled back")
	}

	s.logger.Infow("Machine recharge completed successfully",
		"machine_id", req.MachineID,
		"amount", rechargeAmt,
		"new_balance", newBalance,
	)

	return nil
}

// RechargeRFIDService handles RFID-based recharge operations
func (s *RechargeMachineService) RechargeRFIDService(ctx context.Context, req model.RechargeRFIDRequest) error {
	ctx, cancel := context.WithTimeout(ctx, defaultOperationTimeout)
	defer cancel()

	// Validate recharge amount
	if err := s.validateRechargeAmount(req.RechargeAmount); err != nil {
		s.logger.Warnw("Invalid RFID recharge amount provided",
			"machine_id", req.MachineID,
			"user_id", req.UserID,
			"amount", req.RechargeAmount,
			"error", err,
		)
		return err
	}

	rechargeAmt, _ := strconv.Atoi(req.RechargeAmount)

	s.logger.Infow("Processing RFID recharge",
		"machine_id", req.MachineID,
		"user_id", req.UserID,
		"amount", rechargeAmt,
	)

	// Get current machine balance
	currentBalance, err := s.repo.GetMachineBalance(ctx, req.MachineID)
	if err != nil {
		s.logger.Errorw("Failed to fetch machine balance for RFID recharge",
			"machine_id", req.MachineID,
			"error", err,
		)
		return errors.New("machine not found or balance unavailable")
	}

	// Parse current balance
	currentBalanceInt, err := strconv.Atoi(currentBalance)
	if err != nil {
		s.logger.Errorw("Machine balance is in invalid format during RFID recharge",
			"machine_id", req.MachineID,
			"balance", currentBalance,
			"error", err,
		)
		return errors.New("machine balance is corrupted. Please contact support")
	}

	// Verify sufficient balance
	if currentBalanceInt < rechargeAmt {
		s.logger.Warnw("Insufficient machine balance for RFID recharge",
			"machine_id", req.MachineID,
			"available_balance", currentBalanceInt,
			"requested_amount", rechargeAmt,
		)
		return errors.New("insufficient machine balance for this recharge")
	}

	// Calculate new balance (deduct for RFID recharge)
	newBalance := currentBalanceInt - rechargeAmt
	newBalanceStr := strconv.Itoa(newBalance)

	// Update machine balance
	if err := s.repo.UpdateMachineBalance(ctx, req.MachineID, newBalanceStr); err != nil {
		s.logger.Errorw("Failed to update machine balance for RFID recharge",
			"machine_id", req.MachineID,
			"new_balance", newBalance,
			"error", err,
		)
		return errors.New("failed to update machine balance. Please try again")
	}

	// Record transaction in RFID history
	if err := s.recordRFIDRechargeHistory(ctx, req, currentBalance); err != nil {
		// Rollback balance update on history insertion failure
		s.logger.Errorw("Failed to record RFID recharge history, attempting rollback",
			"machine_id", req.MachineID,
			"user_id", req.UserID,
			"error", err,
		)
		_ = s.repo.UpdateMachineBalance(ctx, req.MachineID, currentBalance)
		return errors.New("failed to record recharge history. Transaction rolled back")
	}

	s.logger.Infow("RFID recharge completed successfully",
		"machine_id", req.MachineID,
		"user_id", req.UserID,
		"amount", rechargeAmt,
		"new_balance", newBalance,
	)

	return nil
}

// GetRFIDRechargeHistoryService retrieves RFID recharge history for a machine
func (s *RechargeMachineService) GetRFIDRechargeHistoryService(
	ctx context.Context,
	machineID string,
) ([]model.RechargerRFIDHistory, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultOperationTimeout)
	defer cancel()

	s.logger.Infow("Fetching RFID recharge history",
		"machine_id", machineID,
	)

	history, err := s.repo.GetRFIDRechargeHistory(ctx, machineID)
	if err != nil {
		s.logger.Errorw("Failed to fetch RFID recharge history",
			"machine_id", machineID,
			"error", err,
		)
		return nil, errors.New("unable to fetch recharge history at this time")
	}

	s.logger.Infow("RFID recharge history fetched successfully",
		"machine_id", machineID,
		"record_count", len(history),
	)

	return history, nil
}

// GetMachineBalanceService retrieves the current balance of a machine
func (s *RechargeMachineService) GetMachineBalanceService(
	ctx context.Context,
	machineID string,
) (model.MachineBalanceResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultOperationTimeout)
	defer cancel()

	s.logger.Infow("Fetching machine balance",
		"machine_id", machineID,
	)

	balance, err := s.repo.GetMachineBalance(ctx, machineID)
	if err != nil {
		s.logger.Errorw("Failed to fetch machine balance",
			"machine_id", machineID,
			"error", err,
		)
		return model.MachineBalanceResponse{}, err
	}

	s.logger.Infow("Machine balance fetched successfully",
		"machine_id", machineID,
		"balance", balance,
	)

	return model.MachineBalanceResponse{
		MachineID: machineID,
		Balance:   balance,
	}, nil
}

// GetRechargeMachineHistoryService retrieves machine recharge history
func (s *RechargeMachineService) GetRechargeMachineHistoryService(
	ctx context.Context,
	machineID string,
) ([]model.MachineRechargeHistory, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultOperationTimeout)
	defer cancel()

	s.logger.Infow("Fetching machine recharge history",
		"machine_id", machineID,
	)

	history, err := s.repo.GetRechargeMachineHistory(ctx, machineID)
	if err != nil {
		s.logger.Errorw("Failed to fetch machine recharge history",
			"machine_id", machineID,
			"error", err,
		)
		return nil, errors.New("unable to fetch recharge history at this time")
	}

	s.logger.Infow("Machine recharge history fetched successfully",
		"machine_id", machineID,
		"record_count", len(history),
	)

	return history, nil
}

// FetchConnectedMachinesService retrieves machine details by machine number
func (s *RechargeMachineService) FetchConnectedMachinesService(
	ctx context.Context,
	machineNo string,
) (model.Machine, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultOperationTimeout)
	defer cancel()

	s.logger.Infow("Fetching connected machines",
		"machine_no", machineNo,
	)

	machine, err := s.repo.GetMachineDetailsByMachineNo(ctx, machineNo)
	if err != nil {
		s.logger.Errorw("Failed to fetch connected machines",
			"machine_no", machineNo,
			"error", err,
		)
		return model.Machine{}, fmt.Errorf("unable to fetch connected machines: %w", err)
	}

	s.logger.Infow("Connected machines fetched successfully",
		"machine_no", machineNo,
		"machine_id", machine.MachineID,
	)

	return machine, nil
}

// GetAvailableMachinesService retrieves all available machines
func (s *RechargeMachineService) GetAvailableMachinesService(ctx context.Context) ([]model.Machine, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultOperationTimeout)
	defer cancel()

	s.logger.Info("Fetching all available machines")

	machines, err := s.repo.GetAvailableMachines(ctx)
	if err != nil {
		s.logger.Errorw("Failed to fetch available machines",
			"error", err,
		)
		return nil, fmt.Errorf("unable to fetch available machines: %w", err)
	}

	s.logger.Infow("Available machines fetched successfully",
		"machine_count", len(machines),
	)

	return machines, nil
}

// CreateUserService creates a new warden user
func (s *RechargeMachineService) CreateUserService(
	ctx context.Context,
	req model.UserAccessCreateRequest,
) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultOperationTimeout)
	defer cancel()

	// Sanitize inputs
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.UserName = strings.TrimSpace(req.UserName)

	s.logger.Infow("Creating warden user",
		"email", req.Email,
		"username", req.UserName,
		"machine_id", req.MachineId,
	)

	// Check if email already exists
	existingUser, _ := s.repo.GetUserByEmail(ctx, req.Email)
	if existingUser != nil {
		s.logger.Warnw("Attempted to create user with existing email",
			"email", req.Email,
		)
		return nil, errors.New("this email is already registered. Please use a different email")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		s.logger.Errorw("Failed to hash password during user creation",
			"email", req.Email,
			"error", err,
		)
		return nil, errors.New("unable to create account at this time. Please try again")
	}

	// Generate unique user ID
	userID := utils.GenerateUUID()

	// Fetch machine name
	machineName := s.repo.GetMachineNameByID(ctx, req.MachineId)
	if machineName == "" {
		s.logger.Warnw("Machine name not found for user creation",
			"machine_id", req.MachineId,
		)
	}

	// Build user entity
	user := &model.User{
		UserID:      userID,
		UserName:    req.UserName,
		Password:    hashedPassword,
		Email:       req.Email,
		MachineID:   req.MachineId,
		MachineName: machineName,
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}

	// Persist to database
	if err := s.repo.CreateUser(ctx, user); err != nil {
		s.logger.Errorw("Failed to create user in database",
			"email", req.Email,
			"error", err,
		)
		return nil, errors.New("unable to create account at this time. Please try again later")
	}

	s.logger.Infow("Warden user created successfully",
		"user_id", userID,
		"email", req.Email,
		"machine_id", req.MachineId,
	)

	return user, nil
}

// LoginUserService handles warden user authentication
func (s *RechargeMachineService) LoginUserService(
	ctx context.Context,
	req model.UserAccessLoginRequest,
) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultOperationTimeout)
	defer cancel()

	// Sanitize email
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	s.logger.Infow("User login attempt",
		"email", req.Email,
	)

	// Retrieve user by email
	user, err := s.repo.GetWardenByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, apperrors.ErrFacultyNotFound) {
			s.logger.Warnw("Login attempt for non-existent user",
				"email", req.Email,
			)
			return "", apperrors.ErrFacultyNotFound
		}
		s.logger.Errorw("Database error during user lookup",
			"email", req.Email,
			"error", err,
		)
		return "", fmt.Errorf("user lookup failed: %w", err)
	}

	// Verify password
	if !utils.CheckPassword(req.Password, user.Password) {
		s.logger.Warnw("Invalid password attempt",
			"email", req.Email,
		)
		return "", apperrors.ErrInvalidPassword
	}

	// Generate access token
	accessToken, err := utils.GenerateAccessTokenForWarden(user.UserID)
	if err != nil {
		s.logger.Errorw("Failed to generate access token",
			"user_id", user.UserID,
			"email", req.Email,
			"error", err,
		)
		return "", fmt.Errorf("token generation failed: %w", err)
	}

	s.logger.Infow("User logged in successfully",
		"user_id", user.UserID,
		"email", req.Email,
	)

	return accessToken, nil
}

// validateRechargeAmount validates that the recharge amount is a positive whole number
func (s *RechargeMachineService) validateRechargeAmount(amount string) error {
	// Check for decimal point
	if strings.Contains(amount, ".") {
		return errors.New("recharge amount must be a whole number")
	}

	// Parse and validate amount
	amountInt, err := strconv.Atoi(amount)
	if err != nil {
		return errors.New("recharge amount must be a valid number")
	}

	if amountInt <= 0 {
		return errors.New("recharge amount must be greater than zero")
	}

	return nil
}

// recordMachineRechargeHistory records a machine recharge transaction in history
func (s *RechargeMachineService) recordMachineRechargeHistory(
	ctx context.Context,
	req model.MachineRechargeRequest,
	previousBalance string,
) error {
	// Load timezone
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
	history := model.MachineRechargeHistory{
		MachineID:      req.MachineID,
		RechargeAmount: req.RechargeAmount,
		Date:           now.Format(dateFormat),
		Time:           now.Format(timeFormat),
	}

	// Insert into database
	if err := s.repo.InsertRechargeHistory(ctx, history); err != nil {
		s.logger.Errorw("Failed to insert machine recharge history",
			"machine_id", req.MachineID,
			"error", err,
		)
		return err
	}

	return nil
}

// recordRFIDRechargeHistory records an RFID recharge transaction in history
func (s *RechargeMachineService) recordRFIDRechargeHistory(
	ctx context.Context,
	req model.RechargeRFIDRequest,
	previousBalance string,
) error {
	// Load timezone
	loc, err := time.LoadLocation(timezoneLocation)
	if err != nil {
		s.logger.Errorw("Failed to load timezone for RFID recharge history",
			"timezone", timezoneLocation,
			"error", err,
		)
		return errors.New("timezone configuration error")
	}

	now := time.Now().In(loc)

	// Build history record
	history := model.RechargerRFIDHistory{
		MachineID:      req.MachineID,
		UserID:         req.UserID,
		RechargeAmount: req.RechargeAmount,
		Date:           now.Format(dateFormat),
		Time:           now.Format(timeFormat),
	}

	// Insert into database
	if err := s.repo.InsertRechargeRFIDHistory(ctx, history); err != nil {
		s.logger.Errorw("Failed to insert RFID recharge history",
			"machine_id", req.MachineID,
			"user_id", req.UserID,
			"error", err,
		)
		return err
	}

	return nil
}