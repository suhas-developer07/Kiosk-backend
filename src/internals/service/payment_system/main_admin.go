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

const (
	defaultOperationTimeout = 5 * time.Second
	timezoneLocation        = "Asia/Kolkata"
	initialMachineBalance   = "0"
	dateFormat              = "2006-01-02"
	timeFormat              = "15:04:05"
)

type MainAdminService struct {
	repo   *repository.MainAdminRepo
	logger *zap.SugaredLogger
}

func NewMainAdminService(
	MainAdminRepo *repository.MainAdminRepo,
	logger *zap.SugaredLogger,
) *MainAdminService {
	return &MainAdminService{
		repo:   MainAdminRepo,
		logger: logger,
	}
}

/* College Auth Service */
func (s *MainAdminService) CollegeLoginService(ctx context.Context, req model.CollegeLoginRequest) (*model.CollegeTokenResponse, error) {

	collegeID, name, hashedPassword, superadminID, err := s.repo.GetCollegeForLogin(ctx, req.CollegeEmail)
	if err != nil {
		return nil, errors.New("college account not found")
	}

	if !utils.CheckPassword(req.CollegePassword, hashedPassword) {
		return nil, apperrors.ErrInvalidPassword
	}

	balance, err := s.repo.GetCollegeBalance(ctx, collegeID)
	if err != nil {
		return nil, errors.New("unable to retrieve account balance at the moment")
	}

	token, err := utils.GenerateAccessTokenForCollegeLogin(req.CollegeEmail, name, collegeID, superadminID)
	if err != nil {
		return nil, errors.New("failed to log in, please try again")
	}

	return &model.CollegeTokenResponse{
		Token:       token,
		Balance:     balance,
		CollegeName: name,
	}, nil
}

/* Recharge Machine Services */
func (s *MainAdminService) GetMachinesByCollegeID(ctx context.Context, collegeID string) ([]model.Machine, error) {
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

func (s *MainAdminService) RechargeMachine(ctx context.Context, req model.MachineRechargeRequest, college_id string) error {
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

	rechargeAmt, err := strconv.Atoi(req.RechargeAmount)

	if err != nil || rechargeAmt <= 0 {
		s.logger.Warnw("Recharge amount must be a positive whole number", "recharge_amount", req.RechargeAmount)
		return errors.New("recharge amount must be a positive whole number")
	}

	collegeBalStr, err := s.repo.GetCollegeBalance(ctx, college_id)
	if err != nil {
		return errors.New("college account not found")
	}

	collegeBal, err := strconv.Atoi(collegeBalStr)
	if err != nil {
		return errors.New("college balance is not in a valid format. Please contact support")
	}

	if collegeBal < rechargeAmt {
		return errors.New("your college balance is too low for this recharge amount")
	}

	s.logger.Infow("Processing machine recharge",
		"machine_id", req.MachineID,
		"amount", rechargeAmt,
	)

	currentBalance, err := s.repo.GetMachineBalance(ctx, req.MachineID)
	if err != nil {
		s.logger.Errorw("Failed to fetch machine balance for recharge",
			"machine_id", req.MachineID,
			"error", err,
		)
		return errors.New("machine not found or balance unavailable")
	}

	currentBalanceInt, err := strconv.Atoi(currentBalance)
	if err != nil {
		s.logger.Errorw("Machine balance is in invalid format",
			"machine_id", req.MachineID,
			"balance", currentBalance,
			"error", err,
		)
		return errors.New("machine balance is corrupted. Please contact support")
	}

	newCollegeBalance := collegeBal - rechargeAmt
	newCollegeBalanceStr := strconv.Itoa(newCollegeBalance)
	newBalance := currentBalanceInt + rechargeAmt
	newBalanceStr := strconv.Itoa(newBalance)

	if err := s.repo.UpdateCollegeBalance(ctx, college_id, newCollegeBalanceStr); err != nil {
		s.logger.Errorw("Failed to update college balance",
			"college_id", college_id,
			"new_balance", newCollegeBalance,
			"error", err,
		)
		return errors.New("failed to update college balance. Please try again")
	}

	if err := s.repo.UpdateMachineBalance(ctx, req.MachineID, newBalanceStr); err != nil {
		s.logger.Errorw("Failed to update machine balance",
			"machine_id", req.MachineID,
			"new_balance", newBalance,
			"error", err,
		)
		_ = s.repo.UpdateCollegeBalance(ctx, college_id, collegeBalStr)
		return errors.New("failed to update machine balance. Please try again")
	}

	if err := s.recordMachineRechargeHistory(ctx, req, college_id); err != nil {
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

func (s *MainAdminService) GetMachineBalanceService(
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

func (s *MainAdminService) GetRechargeMachineHistoryService(
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

func (s *MainAdminService) FetchConnectedMachinesService(
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

func (s *MainAdminService) GetAvailableMachinesService(ctx context.Context) ([]model.Machine, error) {
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

/* RFID Recharge Services */
func (s *MainAdminService) RechargeRFIDService(ctx context.Context, req model.RechargeRFIDRequest) error {
	ctx, cancel := context.WithTimeout(ctx, defaultOperationTimeout)
	defer cancel()

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

	currentBalance, err := s.repo.GetMachineBalance(ctx, req.MachineID)
	if err != nil {
		s.logger.Errorw("Failed to fetch machine balance for RFID recharge",
			"machine_id", req.MachineID,
			"error", err,
		)
		return errors.New("machine not found or balance unavailable")
	}

	currentBalanceInt, err := strconv.Atoi(currentBalance)
	if err != nil {
		s.logger.Errorw("Machine balance is in invalid format during RFID recharge",
			"machine_id", req.MachineID,
			"balance", currentBalance,
			"error", err,
		)
		return errors.New("machine balance is corrupted. Please contact support")
	}

	if currentBalanceInt < rechargeAmt {
		s.logger.Warnw("Insufficient machine balance for RFID recharge",
			"machine_id", req.MachineID,
			"available_balance", currentBalanceInt,
			"requested_amount", rechargeAmt,
		)
		return errors.New("insufficient machine balance for this recharge")
	}

	CardBalance, err := s.repo.GetRFIDCardBalance(ctx, req.CardID)
	if err != nil {
		s.logger.Errorw("Failed to fetch RFID card balance for recharge",
			"card_id", req.CardID,
			"error", err,
		)
		return errors.New("RFID card not found or balance unavailable")
	}

	err = s.repo.RechargeRFIDCard(ctx, req.CardID, req.RechargeAmount)
	if err != nil {
		s.logger.Errorw("Failed to recharge RFID card",
			"card_id", req.CardID,
			"error", err,
		)
		return errors.New("failed to recharge RFID card. Please try again")
	}

	newBalance := currentBalanceInt - rechargeAmt
	newBalanceStr := strconv.Itoa(newBalance)

	if err := s.repo.UpdateMachineBalance(ctx, req.MachineID, newBalanceStr); err != nil {
		s.logger.Errorw("Failed to update machine balance for RFID recharge",
			"machine_id", req.MachineID,
			"new_balance", newBalance,
			"error", err,
		)
		_ = s.repo.UpdateRFIDCardBalance(ctx, req.CardID, CardBalance) // Rollback RFID card recharge
		return errors.New("failed to update machine balance. Please try again")
	}

	if err := s.recordRFIDRechargeHistory(ctx, req); err != nil {
		// Rollback balance update on history insertion failure
		s.logger.Errorw("Failed to record RFID recharge history, attempting rollback",
			"machine_id", req.MachineID,
			"user_id", req.UserID,
			"error", err,
		)
		_ = s.repo.UpdateMachineBalance(ctx, req.MachineID, currentBalance)
		_ = s.repo.UpdateRFIDCardBalance(ctx, req.CardID, CardBalance) // Rollback RFID card recharge
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

func (s *MainAdminService) GetRFIDRechargeHistoryService(
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

func (s *MainAdminService) InitializeCardService(ctx context.Context, req model.InitializeCardRequest) error {
	ctx, cancel := context.WithTimeout(ctx, defaultOperationTimeout)
	defer cancel()

	s.logger.Infow("Initializing RFID card",
		"card_id", req.CardID,
		"usn", req.USN,
	)

	if err := s.validateRechargeAmount(req.RechargeAmount); err != nil {
		s.logger.Warnw("Invalid initial recharge amount for card initialization",
			"card_id", req.CardID,
			"usn", req.USN,
			"amount", req.RechargeAmount,
			"error", err,
		)
		return err
	}
	rechargeAmt, _ := strconv.Atoi(req.RechargeAmount)

	s.logger.Infow("Processing RFID recharge",
		"amount", rechargeAmt,
	)
	CollegeID, err := s.repo.GetCollegeIdByMachineID(ctx, req.MachineID)
	if err != nil {
		s.logger.Errorw("Failed to get college ID by machine ID",
			"machine_id", req.MachineID,
			"error", err,
		)
		return errors.New("failed to initialize card. Please try again")
	}

	currentBalance, err := s.repo.GetMachineBalance(ctx, req.MachineID)
	if err != nil {
		s.logger.Errorw("Failed to fetch machine balance for card initialization",
			"machine_id", req.MachineID,
			"error", err,
		)
		return errors.New("machine not found or balance unavailable")
	}

	currentBalanceInt, err := strconv.Atoi(currentBalance)
	if err != nil {
		s.logger.Errorw("Machine balance is in invalid format during card initialization",
			"machine_id", req.MachineID,
			"balance", currentBalance,
			"error", err,
		)
		return errors.New("machine balance is corrupted. Please contact support")
	}

	if currentBalanceInt < rechargeAmt {
		s.logger.Warnw("Insufficient machine balance for card initialization",
			"machine_id", req.MachineID,
			"available_balance", currentBalanceInt,
			"requested_amount", rechargeAmt,
		)
		return errors.New("insufficient machine balance for this recharge")
	}

	data := model.RFIDCard{
		CardID:    req.CardID,
		USN:       req.USN,
		Balance:   req.RechargeAmount,
		CollegeID: CollegeID,
		Status:    "active",
	}
	err = s.repo.InitializeRFIDCard(ctx, data)
	if err != nil {
		s.logger.Errorw("Failed to initialize RFID card",
			"card_id", req.CardID,
			"usn", req.USN,
			"error", err,
		)
		return errors.New("failed to initialize card. Please try again")
	}

	newBalance := currentBalanceInt - rechargeAmt
	newBalanceStr := strconv.Itoa(newBalance)

	if err := s.repo.UpdateMachineBalance(ctx, req.MachineID, newBalanceStr); err != nil {
		s.logger.Errorw("Failed to update machine balance for RFID recharge",
			"machine_id", req.MachineID,
			"new_balance", newBalance,
			"error", err,
		)
		_ = s.repo.DeleteRFIDCard(ctx, req.CardID) // Rollback card initialization
		return errors.New("failed to update machine balance. Please try again")
	}

	rfidRechargeReq := model.RechargeRFIDRequest{
		UserID:         req.UserID,
		CardID:         req.CardID,
		MachineID:      req.MachineID,
		RechargeAmount: req.RechargeAmount,
	}

	if err := s.recordRFIDRechargeHistory(ctx, rfidRechargeReq); err != nil {
		// Rollback balance update on history insertion failure
		s.logger.Errorw("Failed to record RFID recharge history, attempting rollback",
			"machine_id", req.MachineID,
			"card_id", req.CardID,
			"usn", req.USN,
			"error", err,
		)
		_ = s.repo.UpdateMachineBalance(ctx, req.MachineID, currentBalance)
		_ = s.repo.DeleteRFIDCard(ctx, req.CardID) // Rollback card initialization
		return errors.New("failed to record recharge history. Transaction rolled back")
	}

	s.logger.Infow("RFID card initialized successfully",
		"card_id", req.CardID,
		"usn", req.USN,
	)

	return nil
}

func (s *MainAdminService) GetRFIDCardBalanceService(ctx context.Context, cardID string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultOperationTimeout)
	defer cancel()

	s.logger.Infow("Fetching RFID card balance",
		"card_id", cardID,
	)

	balance, err := s.repo.GetRFIDCardBalance(ctx, cardID)
	if err != nil {
		s.logger.Errorw("Failed to fetch RFID card balance",
			"card_id", cardID,
			"error", err,
		)
		return "", errors.New("unable to fetch RFID card balance at this time")
	}

	s.logger.Infow("RFID card balance fetched successfully",
		"card_id", cardID,
		"balance", balance,
	)

	return balance, nil
}

func (s *MainAdminService) GetRFIDCardDetailsService(ctx context.Context, cardID string) (model.RFIDCard, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultOperationTimeout)
	defer cancel()

	s.logger.Infow("Fetching RFID card details",
		"card_id", cardID,
	)

	card, err := s.repo.GetRFIDCardDetails(ctx, cardID)
	if err != nil {
		s.logger.Errorw("Failed to fetch RFID card details",
			"card_id", cardID,
			"error", err,
		)
		return model.RFIDCard{}, errors.New("unable to fetch RFID card details at this time")
	}

	s.logger.Infow("RFID card details fetched successfully",
		"card_id", cardID,
	)

	return card, nil
}
func (s *MainAdminService) CardDeactivationService(ctx context.Context, cardID string, status string) error {
	ctx, cancel := context.WithTimeout(ctx, defaultOperationTimeout)
	defer cancel()

	s.logger.Infow("Deactivating RFID card",
		"card_id", cardID,
		"status", status,
	)

	if err := s.repo.DeactivateRFIDCard(ctx, cardID, status); err != nil {
		s.logger.Errorw("Failed to deactivate RFID card",
			"card_id", cardID,
			"status", status,
			"error", err,
		)
		return errors.New("failed to deactivate card. Please try again")
	}

	s.logger.Infow("RFID card deactivated successfully",
		"card_id", cardID,
		"status", status,
	)

	return nil
}
/* Recharge Machine User Services */
func (s *MainAdminService) CreateRechargeMachineUserService(
	ctx context.Context,
	req model.UserAccessCreateRequest,
	college_id string,
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

	existingUser, _ := s.repo.GetRechargeMachineUserByEmail(ctx, req.Email)
	if existingUser != nil {
		s.logger.Warnw("Attempted to create user with existing email",
			"email", req.Email,
		)
		return nil, errors.New("this email is already registered. Please use a different email")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		s.logger.Errorw("Failed to hash password during user creation",
			"email", req.Email,
			"error", err,
		)
		return nil, errors.New("unable to create account at this time. Please try again")
	}

	userID := utils.GenerateUUID()

	machineName := s.repo.GetMachineNameByID(ctx, req.MachineId)
	if machineName == "" {
		s.logger.Warnw("Machine name not found for user creation",
			"machine_id", req.MachineId,
		)
	}

	user := &model.User{
		UserID:      userID,
		UserName:    req.UserName,
		Password:    hashedPassword,
		Email:       req.Email,
		MachineID:   req.MachineId,
		MachineName: machineName,
		College_id:  college_id,
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}

	if err := s.repo.CreateRechargeMachineUser(ctx, user); err != nil {
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

func (s *MainAdminService) LoginRechargeMachineUserService(
	ctx context.Context,
	req model.UserAccessLoginRequest,
) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultOperationTimeout)
	defer cancel()

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	s.logger.Infow("User login attempt",
		"email", req.Email,
	)

	user, err := s.repo.GetRechargeMachineUserByEmail(ctx, req.Email)
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

	if !utils.CheckPassword(req.Password, user.Password) {
		s.logger.Warnw("Invalid password attempt",
			"email", req.Email,
		)
		return "", apperrors.ErrInvalidPassword
	}

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

func (s *MainAdminService) validateRechargeAmount(amount string) error {

	if strings.Contains(amount, ".") {
		return errors.New("recharge amount must be a whole number")
	}

	amountInt, err := strconv.Atoi(amount)
	if err != nil {
		return errors.New("recharge amount must be a valid number")
	}

	if amountInt <= 0 {
		return errors.New("recharge amount must be greater than zero")
	}

	return nil
}

/* Helper methods for recording transaction history */
func (s *MainAdminService) recordMachineRechargeHistory(
	ctx context.Context,
	req model.MachineRechargeRequest,
	college_id string,
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

	Machine, err := s.repo.GetMachineDetails(ctx, req.MachineID)

	if err != nil {
		s.logger.Errorw("Failed to get machine details of the id",
			"machine_id", req.MachineID,
			"error", err,
		)
		return err
	}

	history := model.MachineRechargeHistory{
		SuperAdminID:   Machine.SuperAdminId,
		CollegeID:      college_id,
		MachineID:      req.MachineID,
		MachineName:    Machine.MachineName,
		RechargeAmount: req.RechargeAmount,
		Date:           now.Format(dateFormat),
		Time:           now.Format(timeFormat),
	}

	if err := s.repo.InsertRechargeHistory(ctx, history); err != nil {
		s.logger.Errorw("Failed to insert machine recharge history",
			"machine_id", req.MachineID,
			"error", err,
		)
		return err
	}

	return nil
}

func (s *MainAdminService) recordRFIDRechargeHistory(
	ctx context.Context,
	req model.RechargeRFIDRequest,
) error {
	loc, err := time.LoadLocation(timezoneLocation)
	if err != nil {
		s.logger.Errorw("Failed to load timezone for RFID recharge history",
			"timezone", timezoneLocation,
			"error", err,
		)
		return errors.New("timezone configuration error")
	}

	now := time.Now().In(loc)

	Machine, err := s.repo.GetMachineDetails(ctx, req.MachineID)
	if err != nil {
		s.logger.Errorw("Failed to get machine detailes of this id",
			"machine_id", req.MachineID,
			"err", err)
		return err
	}

	UserName, err := s.repo.GetRechargeMachineUserById(ctx, req.UserID)
	if err != nil {
		s.logger.Errorw("failed to get username for this user id",
			"user_id", req.UserID)
		return err
	}

	USN, err := s.repo.GetUSNByCardID(ctx, req.CardID)
	if err != nil {
		s.logger.Errorw("failed to get USN for this card id",
			"card_id", req.CardID)
		return err
	}

	history := model.RechargerRFIDHistory{
		SuperAdminID:   Machine.SuperAdminId,
		CollegeID:      Machine.CollegeId,
		MachineID:      req.MachineID,
		MachineName:    Machine.MachineName,
		UserID:         req.UserID,
		UserName:       UserName,
		CardID:         req.CardID,
		USN:            USN,
		RechargeAmount: req.RechargeAmount,
		Date:           now.Format(dateFormat),
		Time:           now.Format(timeFormat),
	}

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

func (s *MainAdminService) GetMachineUsersByMachineIdService(ctx context.Context, machineId string)([]model.User,error){
	ctx, cancel := context.WithTimeout(ctx, defaultOperationTimeout)
	defer cancel()

	s.logger.Infow("Fetching machine users by machine id",
		"machine_id", machineId,
	)

	users, err := s.repo.GetMachineUsersByMachineId(ctx, machineId)
	if err != nil {
		s.logger.Errorw("Failed to fetch machine users by machine id",
			"machine_id", machineId,
			"error", err,
		)
		return nil, errors.New("unable to fetch machine users at this time")
	}

	s.logger.Infow("Machine users fetched successfully",
		"machine_id", machineId,
		"user_count", len(users),
	)

	return users, nil
}

func (s *MainAdminService) DeleteMachineUserService(ctx context.Context, machineId string,userId string) error{
	ctx, cancel := context.WithTimeout(ctx, defaultOperationTimeout)
	defer cancel()

	s.logger.Infow("Deleting machine user",
		"machine_id", machineId,
		"user_id", userId,
	)

	err := s.repo.DeleteMachineUser(ctx, machineId,userId)
	if err != nil {
		s.logger.Errorw("Failed to delete machine user",
			"machine_id", machineId,
			"user_id", userId,
			"error", err,
		)
		return errors.New("unable to delete machine user at this time")
	}

	s.logger.Infow("Machine user deleted successfully",
		"machine_id", machineId,
		"user_id", userId,
	)

	return nil
}