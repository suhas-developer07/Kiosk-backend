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

type RechargeMachineService struct {
	RechargeMachineRepo *repo.RechargeMachineRepo
	Logger              *zap.SugaredLogger
}

func NewRechargeMachineService(RechargeMachineRepo *repo.RechargeMachineRepo, logger *zap.SugaredLogger) *RechargeMachineService {
	return &RechargeMachineService{
		RechargeMachineRepo: RechargeMachineRepo,
		Logger:              logger,
	}
}

func (s *RechargeMachineService) CreateAccountService(ctx context.Context, req model.CreateMainAdminPayload) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.Password != "" {
		hashed, err := utils.HashPassword(req.Password)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}
		req.Password = hashed
	}

	s.Logger.Infof("Creating faculty account | email=%s", req.Email)

	id := utils.GenerateUUID()

	admin := model.MainAdmin{
		ID:        id,
		Username:  req.Name,
		Email:     req.Email,
		Password:  req.Password,
		CreatedAt: time.Now(),
	}
	err := s.RechargeMachineRepo.CreateAccount(ctx, admin)

	switch {
	case errors.Is(err, apperrors.ErrEmailAlreadyExists):
		return apperrors.ErrEmailAlreadyExists

	case err != nil:
		return fmt.Errorf("failed to create account: %w", err)
	}
	s.Logger.Infof("Account created successfully | email=%s", req.Email)

	return nil
}

func (s *RechargeMachineService) CreateMachineService(ctx context.Context, req model.MachineCreateRequest) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req.MachineNo = strings.TrimSpace(req.MachineNo)
	req.MachineName = strings.TrimSpace(req.MachineName)

	if req.MachineNo == "" || req.MachineName == "" {
		return fmt.Errorf("machineNo and machine Name required")
	}

	machineId := utils.GenerateUUID()

	machine := model.Machine{
		MachineID:   machineId,
		MachineNo:   req.MachineNo,
		MachineName: req.MachineName,
		Balance:     "0",
	}
	err := s.RechargeMachineRepo.CreateMachine(ctx, machine)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrMachineAlreadyExists):
			return apperrors.ErrMachineAlreadyExists

		default:
			return fmt.Errorf("failed to create machine: %w", err)
		}
	}
	return nil
}

func (s *RechargeMachineService) RechargeMachine(ctx context.Context, req model.MachineRechargeRequest) error {

	if strings.Contains(req.RechargeAmount, ".") {
		return errors.New("recharge amount must be a whole number")
	}

	rechargeAmt, err := strconv.Atoi(req.RechargeAmount)
	if err != nil || rechargeAmt <= 0 {
		return errors.New("please enter a valid recharge amount greater than zero")
	}

	machineBalStr, err := s.RechargeMachineRepo.GetMachineBalance(ctx, req.MachineID)
	if err != nil {
		return errors.New("machine account not found")
	}

	machineBal, err := strconv.Atoi(machineBalStr)
	if err != nil {
		return errors.New("machine balance is not in a valid format. Please contact support")
	}

	newMachineBal := strconv.Itoa(machineBal + rechargeAmt)

	if err := s.RechargeMachineRepo.UpdateMachineBalance(ctx, req.MachineID, newMachineBal); err != nil {
		return errors.New("failed to update machine balance. Please try again")
	}

	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		return errors.New("failed to load timezone")
	}

	now := time.Now().In(loc)
	history := model.MachineRechargeHistory{

		MachineID:      req.MachineID,
		RechargeAmount: req.RechargeAmount,
		Date:           now.Format("2006-01-02"),
		Time:           now.Format("15:04:05"),
	}
	if err := s.RechargeMachineRepo.InsertRechargeHistory(ctx, history); err != nil {
		_ = s.RechargeMachineRepo.UpdateMachineBalance(ctx, req.MachineID, machineBalStr)
		return errors.New("failed to record recharge history. Please try again")
	}

	return nil
}

func (s *RechargeMachineService) RechargeRFIDService(ctx context.Context, req model.RechargeRFIDRequest) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if strings.Contains(req.RechargeAmount, ".") {
		return errors.New("recharge amount must be a whole number")
	}

	rechargeAmt, err := strconv.Atoi(req.RechargeAmount)
	if err != nil || rechargeAmt <= 0 {
		return errors.New("please enter a valid recharge amount greater than zero")
	}

	machineBalStr, err := s.RechargeMachineRepo.GetMachineBalance(ctx, req.MachineID)
	if err != nil {
		return errors.New("machine account not found")
	}

	machineBal, err := strconv.Atoi(machineBalStr)
	if err != nil {
		return errors.New("machine balance is not in a valid format. Please contact support")
	}

	if machineBal < rechargeAmt {
		return errors.New("insufficient machine balance for this recharge")
	}

	newMachineBal := strconv.Itoa(machineBal - rechargeAmt)
	if err := s.RechargeMachineRepo.UpdateMachineBalance(ctx, req.MachineID, newMachineBal); err != nil {
		return errors.New("failed to update machine balance. Please try again")
	}

	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		return errors.New("failed to load timezone")
	}

	now := time.Now().In(loc)

	history := model.RechargerRFIDHistory{
		MachineID:      req.MachineID,
		UserID:         req.UserID,
		RechargeAmount: req.RechargeAmount,
		Date:           now.Format("2006-01-02"),
		Time:           now.Format("15:04:05"),
	}

	if err := s.RechargeMachineRepo.InsertRechargeRFIDHistory(ctx, history); err != nil {
		_ = s.RechargeMachineRepo.UpdateMachineBalance(ctx, req.MachineID, machineBalStr)
		return errors.New("failed to record recharge history. Please try again")
	}
	return nil
}

func (s *RechargeMachineService) GetRFIDRechargeHistoryService(ctx context.Context, MachineID string) ([]model.RechargerRFIDHistory, error) {
	history, err := s.RechargeMachineRepo.GetRFIDRechargeHistory(ctx, MachineID)
	if err != nil {
		return nil, errors.New("unable to fetch recharge history at the moment")
	}
	return history, nil
}

func (s *RechargeMachineService) GetMachineBalanceService(ctx context.Context, machineID string) (model.MachineBalanceResponse, error) {
	balance, err := s.RechargeMachineRepo.GetMachineBalance(ctx, machineID)
	if err != nil {
		return model.MachineBalanceResponse{}, err
	}
	return model.MachineBalanceResponse{
		MachineID: machineID,
		Balance:   balance,
	}, nil
}

func (s *RechargeMachineService) GetRechargeMachineHistoryService(ctx context.Context, machineID string) ([]model.MachineRechargeHistory, error) {
	history, err := s.RechargeMachineRepo.GetRechargeMachineHistory(ctx, machineID)
	if err != nil {
		return nil, errors.New("unable to fetch recharge history at the moment")
	}
	return history, nil
}


//wardens routes

func (s *RechargeMachineService) CreateUserService(ctx context.Context, req model.UserAccessCreateRequest) (*model.User, error) {

	existing, _ := s.RechargeMachineRepo.GetUserByEmail(ctx, req.Email)
	if existing != nil {
		return nil, errors.New("this email is already registered. Use a different email to continue")
	}

	hashedPwd, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("we couldn't create your account at this moment. Please try again")
	}
	UserID := utils.GenerateUUID()
	MachineName := s.RechargeMachineRepo.GetMachineNameByID(ctx, req.MachineId)

	user := &model.User{
		UserID:    UserID,
		UserName:  req.UserName,
		Password:  hashedPwd,
		Email:     req.Email,
		MachineID: req.MachineId,
		MachineName: MachineName,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	if err := s.RechargeMachineRepo.CreateUser(ctx, user); err != nil {
		return nil, errors.New("unable to create account right now. Please try again later")
	}

	return user, nil
}

func (s *RechargeMachineService) LoginUserService(ctx context.Context, req model.UserAccessLoginRequest) (string,error) {

    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    s.Logger.Infof("Signin attempt | email=%s", req.Email)

    User, err := s.RechargeMachineRepo.GetWardenByEmail(ctx, req.Email)
    if err != nil {
        if errors.Is(err, apperrors.ErrFacultyNotFound) {
            return "", apperrors.ErrFacultyNotFound
        }
        return "", fmt.Errorf("service: db lookup failed: %w", err)
    }

    if !utils.CheckPassword(req.Password, User.Password) {
        return "", apperrors.ErrInvalidPassword
    }

    accessToken, err := utils.GenerateAccessTokenForWarden(User.UserID)
    if err != nil {
        return "", fmt.Errorf("service: failed generating access token: %w", err)
    }

    s.Logger.Infof("Signin successful | email=%s", req.Email)
    return accessToken,nil
}
