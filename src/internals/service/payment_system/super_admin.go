package paymentsystem

import (
	"context"
	"errors"
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

func (s *SuperAdminService) RechargeCollege(ctx context.Context, req model.CollegeRechargeRequest) error {
	amtInt, err := strconv.Atoi(req.RechargeAmount)
	if err != nil || amtInt <= 0 {
		return errors.New("recharge amount must be greater than zero")
	}

	superBalanceStr, err := s.Repo.GetSuperAdminBalance(ctx, req.SuperAdminId)
	if err != nil {
		return errors.New("unable to retrieve super admin balance")
	}
	superBalance, err := strconv.Atoi(superBalanceStr)
	if err != nil {
		return errors.New("super admin balance data is invalid")
	}
	if superBalance < amtInt {
		return errors.New("insufficient super admin balance for recharge")
	}

	currStr, err := s.Repo.GetCollegeBalance(ctx, req.CollegeID)
	if err != nil {
		return errors.New("unable to retrieve current college balance")
	}
	currInt, err := strconv.Atoi(currStr)
	if err != nil {
		return errors.New("college balance data is invalid")
	}
	newBalance := strconv.Itoa(currInt + amtInt)

	recharge := model.CollegeRecharge{
		RechargeID:     utils.GenerateUUID(),
		CollegeID:      req.CollegeID,
		SuperAdminId:   req.SuperAdminId,
		RechargeAmount: req.RechargeAmount,
		RechargedAt:    time.Now().Format(time.RFC3339),
	}
	if err := s.Repo.RechargeCollege(ctx, recharge, req.SuperAdminId); err != nil {
		return errors.New("failed to process recharge")
	}

	if err := s.Repo.UpdateCollegeBalance(ctx, req.CollegeID, newBalance); err != nil {
		return errors.New("failed to update college balance")
	}

	newSuperBalance := strconv.Itoa(superBalance - amtInt)
	if err := s.Repo.UpdateSuperAdminBalance(ctx, req.SuperAdminId, newSuperBalance); err != nil {
		return errors.New("failed to update super admin balance")
	}

	return nil
}

func (s *SuperAdminService) GetRechargeHistory(ctx context.Context, collegeID string) ([]model.CollegeRecharge, error) {
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
