package paymentsystem

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	apperrors "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/errors"
	model "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/payment_system"
	domain "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/response"
	service "github.com/suhas-developer07/Kiosk-backend/src/internals/service/payment_system"
	"github.com/suhas-developer07/Kiosk-backend/src/pkg/utils"
	"go.uber.org/zap"
)

type SuperAdminHandler struct {
	service  *service.SuperAdminService
	logger   *zap.SugaredLogger
	validate *validator.Validate
}

func NewSuperAdminHandler(service *service.SuperAdminService, logger *zap.SugaredLogger) *SuperAdminHandler {
	return &SuperAdminHandler{
		service:  service,
		logger:   logger,
		validate: validator.New(),
	}
}

/*super admin auth handlers */

func (h *SuperAdminHandler) CreateSuperAdmin(c echo.Context) error {
	var req model.SuperAdminCreateRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Warnw("Failed to bind super admin creation request",
			"error", err)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "We could not process your request. Please ensure all required fields are filled correctly.",
		})
	}

	err := h.service.CreateSuperAdmin(c.Request().Context(), req)
	if err != nil {
		h.logger.Errorw("Failed to create super admin",
			"error", err)
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "error",
			Error:  "Unable to create super admin: " + err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, domain.SuccessResponse{
		Status:  "success",
		Message: "Super admin created successfully.",
	})
}

func (h *SuperAdminHandler) LoginSuperAdminHandler(c echo.Context) error {
	ctx := c.Request().Context()
	requestIP := c.RealIP()

	var req model.SuperAdminLoginRequest

	if err := c.Bind(&req); err != nil {
		h.logger.Warnw("Failed to bind login request",
			"ip", requestIP,
			"error", err,
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Invalid request format. Please check your email and password.",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		validationMsg := utils.FormatValidationError(err)
		h.logger.Warnw("Login validation failed",
			"email", req.SuperAdminEmail,
			"validation_error", validationMsg,
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  validationMsg,
		})
	}

	token, email, username, err := h.service.LoginSuperAdminService(ctx, req)
	if err != nil {
		return h.handleLoginError(c, err, req.SuperAdminEmail)
	}

	h.logger.Infow("Super admin logged in successfully",
		"email", req.SuperAdminEmail,
		"ip", requestIP,
	)

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "Login successful.",
		Data: map[string]string{
			"token":    token,
			"email":    email,
			"username": username,
		},
	})
}

/* super admin college management handlers */
func (h *SuperAdminHandler) CreateCollege(c echo.Context) error {
	var req model.CollegeCreateRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Warnw("Failed to bind college creation request",
			"error", err,
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "We could not process your request. Please ensure all required fields are filled correctly.",
		})
	}

	adminID := strings.TrimSpace(c.Param("super_admin_id"))
	if adminID == "" {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Super Admin ID is required to create a college.",
		})
	}

	res, err := h.service.CreateCollege(c.Request().Context(), req, adminID)
	if err != nil {
		h.logger.Errorw("Failed to create college",
			"super_admin_id", adminID,
			"error", err,
		)
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "error",
			Error:  "Unable to create college: " + err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, domain.SuccessResponse{
		Status:  "success",
		Message: "College created successfully.",
		Data:    res,
	})
}

func (h *SuperAdminHandler) CollegeLogin(c echo.Context) error {
	var req model.CollegeLoginRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Warnw("Failed to bind college login request",
			"error", err,
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Invalid login request. Please check your credentials format.",
		})
	}

	res, err := h.service.CollegeLogin(c.Request().Context(), req)
	if err != nil {
		h.logger.Warnw("College login failed",
			"error", err,
		)
		return c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
			Status: "error",
			Error:  "Login failed: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "Login successful.",
		Data:    res,
	})
}

func (h *SuperAdminHandler) GetCollegesBySuperAdmin(c echo.Context) error {
	//TODO:NEED TO RECIEVE SUPER ADMIN ID FROM JWT CLAIMS INSTEAD OF PARAMS
	adminID := strings.TrimSpace(c.Param("super_admin_id"))
	if adminID == "" {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Super Admin ID is required to fetch colleges.",
		})
	}

	colleges, err := h.service.GetCollegesBySuperAdminID(c.Request().Context(), adminID)
	if err != nil {
		h.logger.Errorw("Failed to fetch colleges for super admin",
			"super_admin_id", adminID,
			"error", err,
		)
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "error",
			Error:  "Unable to fetch colleges: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "Colleges fetched successfully.",
		Data:    colleges,
	})
}

func (h *SuperAdminHandler) GetCollegeDetails(c echo.Context) error {
	collegeID := strings.TrimSpace(c.Param("college_id")) //TODO:NEED TO RECIEVE COLLEGE ID FROM JWT CLAIMS INSTEAD OF PARAMS
	if collegeID == "" {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "College ID is required to fetch details.",
		})
	}

	college, err := h.service.GetCollegeDetails(c.Request().Context(), collegeID)
	if err != nil {
		h.logger.Errorw("Failed to fetch college details",
			"college_id", collegeID,
			"error", err,
		)
		return c.JSON(http.StatusNotFound, domain.ErrorResponse{
			Status: "error",
			Error:  "College not found: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "College details retrieved successfully.",
		Data:    college,
	})
}

func (h *SuperAdminHandler) DeleteCollege(c echo.Context) error {
	collegeID := strings.TrimSpace(c.Param("college_id"))
	if collegeID == "" {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "College ID is required to delete.",
		})
	}

	if err := h.service.DeleteCollege(c.Request().Context(), collegeID); err != nil {
		h.logger.Errorw("Failed to delete college",
			"college_id", collegeID,
			"error", err,
		)
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "error",
			Error:  "Unable to delete college: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "College deleted successfully.",
	})
}

func (h *SuperAdminHandler) RechargeCollege(c echo.Context) error {
	type Request struct {
		RechargeAmount string `json:"recharge_amount"`
	}

	var req Request
	if err := c.Bind(&req); err != nil {
		h.logger.Warnw("Failed to bind college recharge request",
			"error", err,
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Invalid recharge request format.",
		})
	}

	collegeID := strings.TrimSpace(c.Param("college_id"))
	if collegeID == "" {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "College ID is required for recharge.",
		})
	}

	superadminID := strings.TrimSpace(c.Param("super_admin_id"))
	if superadminID == "" {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Super Admin ID is required for recharge.",
		})
	}

	if err := h.service.RechargeToCollege(c.Request().Context(), model.CollegeRechargeRequest{
		CollegeID:      collegeID,
		RechargeAmount: req.RechargeAmount,
		SuperAdminId:   superadminID,
	}); err != nil {
		h.logger.Errorw("Failed to recharge college account",
			"college_id", collegeID,
			"super_admin_id", superadminID,
			"amount", req.RechargeAmount,
			"error", err,
		)
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "error",
			Error:  "Unable to process recharge: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "College account recharged successfully.",
	})
}

func (h *SuperAdminHandler) GetCollegeRechargeHistory(c echo.Context) error {
	collegeID := strings.TrimSpace(c.Param("college_id"))
	if collegeID == "" {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "College ID is required to fetch recharge history.",
		})
	}

	ctx := context.Background()
	history, err := h.service.GetRechargeHistory(ctx, collegeID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "error",
			Error:  "Unable to fetch recharge history: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, model.CollegeRechargeHistoryResponse{
		Status:  "success",
		Message: "Recharge history retrieved successfully.",
		Data:    history,
	})
}

func (h *SuperAdminHandler) GetSuperAdminBalance(c echo.Context) error {
	ctx := c.Request().Context()
	superAdminID := c.Param("super_admin_id")
	if superAdminID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "super admin ID is required",
		})
	}

	balance, err := h.service.GetSuperAdminBalance(ctx, superAdminID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "Balance fetched successfully",
		Data: map[string]string{
			"balance": balance,
		},
	})
}

/*super admin machine management handlers */

func (h *SuperAdminHandler) CreateMachineHandler(c echo.Context) error {
	ctx := c.Request().Context()
	requestIP := c.RealIP()

	var payload model.MachineCreateRequest

	if err := utils.DecodeAndValidateJSON(c.Request().Body, &payload); err != nil {
		h.logger.Warnw("Failed to decode machine creation payload",
			"ip", requestIP,
			"error", err,
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Invalid request payload. Please check your input.",
		})
	}

	if err := h.validate.Struct(&payload); err != nil {
		validationMsg := utils.FormatValidationError(err)
		h.logger.Warnw("Machine creation validation failed",
			"machine_no", payload.MachineNo,
			"validation_error", validationMsg,
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  validationMsg,
		})
	}

	if err := h.service.CreateMachineService(ctx, payload); err != nil {
		return h.handleCreateMachineError(c, err, payload.MachineNo)
	}

	h.logger.Infow("Machine created successfully",
		"machine_no", payload.MachineNo,
		"machine_name", payload.MachineName,
		"ip", requestIP,
	)

	return c.JSON(http.StatusCreated, domain.SuccessResponse{
		Status:  "success",
		Message: "Machine created successfully",
	})
}

func (h *SuperAdminHandler) GetMachinesByCollegeID(c echo.Context) error {
	ctx := c.Request().Context()
	requestIP := c.RealIP()

	collegeID := c.Param("college_id")
	if collegeID == "" {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "College ID is required to fetch machines.",
		})
	}

	machines, err := h.service.GetMachinesByCollegeID(ctx, collegeID)
	if err != nil {
		if errors.Is(err, apperrors.ErrCollegeNotFound) {
			h.logger.Warnw("College not found when fetching machines",
				"college_id", collegeID,
				"ip", requestIP,
			)
			return c.JSON(http.StatusNotFound, domain.ErrorResponse{
				Status: "error",
				Error:  "College not found.",
			})
		}

		h.logger.Errorw("Failed to fetch machines for college",
			"college_id", collegeID,
			"ip", requestIP,
			"error", err,
		)
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "error",
			Error:  "Unable to fetch machines: " + err.Error(),
		})
	}

	if len(machines) == 0 {
		h.logger.Infow("No machines found for college",
			"college_id", collegeID,
			"ip", requestIP,
		)
		return c.JSON(http.StatusOK, domain.SuccessResponse{
			Status:  "success",
			Message: "No machines found for this college.",
			Data:    []model.Machine{},
		})
	}

	h.logger.Infow("Machines fetched successfully for college",
		"college_id", collegeID,
		"machine_count", len(machines),
		"ip", requestIP,
	)
	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "Machines fetched successfully.",
		Data:    machines,
	})
}

/*
super admin Error handling
*/
func (h *SuperAdminHandler) handleLoginError(c echo.Context, err error, email string) error {
	switch {
	case errors.Is(err, apperrors.ErrSuperAdminNotFound):
		h.logger.Warnw("User not found during login",
			"email", email,
		)
		return c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
			Status: "error",
			Error:  "Invalid email or password.",
		})

	case errors.Is(err, apperrors.ErrInvalidPassword):
		h.logger.Warnw("Invalid password during login",
			"email", email,
		)
		return c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
			Status: "error",
			Error:  "Invalid email or password.",
		})

	default:
		h.logger.Errorw("Unexpected error during login",
			"email", email,
			"error", err,
		)
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "error",
			Error:  "Unable to process login at this time. Please try again later.",
		})
	}
}

func (h *SuperAdminHandler) handleCreateMachineError(c echo.Context, err error, machineNo string) error {
	switch {
	case errors.Is(err, apperrors.ErrMachineAlreadyExists):
		h.logger.Warnw("Machine already exists during creation",
			"machine_no", machineNo,
		)
		return c.JSON(http.StatusConflict, domain.ErrorResponse{
			Status: "error",
			Error:  "A machine with this number already exists.",
		})

	default:
		h.logger.Errorw("Unexpected error during machine creation",
			"machine_no", machineNo,
			"error", err,
		)
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "error",
			Error:  "Unable to create machine at this time. Please try again later.",
		})
	}
}
