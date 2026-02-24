package paymentsystem

import (
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

func NewSuperAdminHandler(
	superAdminService *service.SuperAdminService,
	logger *zap.SugaredLogger,
) *SuperAdminHandler {
	return &SuperAdminHandler{
		service:  superAdminService,
		logger:   logger,
		validate: validator.New(),
	}
}

func (h *SuperAdminHandler) CreateSuperAdmin(c echo.Context) error {
	var req model.SuperAdminCreateRequest

	if err := c.Bind(&req); err != nil {
		h.logger.Warnw("Failed to bind super admin creation request",
			"error", err,
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "We could not process your request. Please ensure all required fields are filled correctly.",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		validationMsg := utils.FormatValidationError(err)
		h.logger.Warnw("Super admin creation validation failed",
			"email", req.SuperAdminEmail,
			"validation_error", validationMsg,
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  validationMsg,
		})
	}

	if err := h.service.CreateSuperAdmin(c.Request().Context(), req); err != nil {
		h.logger.Errorw("Failed to create super admin",
			"email", req.SuperAdminEmail,
			"error", err,
		)
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "error",
			Error:  "Unable to create super admin. Please try again later.",
		})
	}

	h.logger.Infow("Super admin created successfully",
		"email", req.SuperAdminEmail,
	)

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
		h.logger.Warnw("Super admin login validation failed",
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

/* Super Admin College Management Handlers  */

func (h *SuperAdminHandler) CreateCollege(c echo.Context) error {
	requestIP := c.RealIP()

	var req model.CollegeCreateRequest

	if err := c.Bind(&req); err != nil {
		h.logger.Warnw("Failed to bind college creation request",
			"ip", requestIP,
			"error", err,
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "We could not process your request. Please ensure all required fields are filled correctly.",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		validationMsg := utils.FormatValidationError(err)
		h.logger.Warnw("College creation validation failed",
			"email", req.CollegeEmail,
			"validation_error", validationMsg,
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  validationMsg,
		})
	}

	adminID, ok := c.Get("super_admin_id").(string)
	if !ok || adminID == "" {
		h.logger.Warnw("Missing super admin ID in context",
			"ip", requestIP,
		)
		return c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
			Status: "error",
			Error:  "Unauthorized: super admin identity could not be verified.",
		})
	}

	res, err := h.service.CreateCollege(c.Request().Context(), req, adminID)
	if err != nil {
		h.logger.Errorw("Failed to create college",
			"super_admin_id", adminID,
			"email", req.CollegeEmail,
			"error", err,
		)
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "error",
			Error:  "Unable to create college. Please try again later.",
		})
	}

	h.logger.Infow("College created successfully",
		"super_admin_id", adminID,
		"college_id", res.CollegeID,
	)

	return c.JSON(http.StatusCreated, domain.SuccessResponse{
		Status:  "success",
		Message: "College created successfully.",
		Data:    res,
	})
}

func (h *SuperAdminHandler) GetCollegesBySuperAdmin(c echo.Context) error {
	ctx := c.Request().Context()
	requestIP := c.RealIP()

	adminID, ok := c.Get("super_admin_id").(string)
	if !ok || adminID == "" {
		h.logger.Warnw("Missing super admin ID in context",
			"ip", requestIP,
		)
		return c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
			Status: "error",
			Error:  "Unauthorized: super admin identity could not be verified.",
		})
	}

	colleges, err := h.service.GetCollegesBySuperAdminID(ctx, adminID)
	if err != nil {
		h.logger.Errorw("Failed to fetch colleges for super admin",
			"super_admin_id", adminID,
			"ip", requestIP,
			"error", err,
		)
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "error",
			Error:  "Unable to fetch colleges at this time.",
		})
	}

	h.logger.Infow("Colleges fetched successfully",
		"super_admin_id", adminID,
		"college_count", len(colleges),
		"ip", requestIP,
	)

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "Colleges fetched successfully.",
		Data:    colleges,
	})
}

func (h *SuperAdminHandler) GetCollegeDetails(c echo.Context) error {
	ctx := c.Request().Context()
	requestIP := c.RealIP()

	collegeID := strings.TrimSpace(c.Param("college_id"))
	if collegeID == "" {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "College ID is required to fetch details.",
		})
	}

	college, err := h.service.GetCollegeDetails(ctx, collegeID)
	if err != nil {
		if errors.Is(err, apperrors.ErrCollegeNotFound) {
			h.logger.Warnw("College not found",
				"college_id", collegeID,
				"ip", requestIP,
			)
			return c.JSON(http.StatusNotFound, domain.ErrorResponse{
				Status: "error",
				Error:  "College not found.",
			})
		}
		h.logger.Errorw("Failed to fetch college details",
			"college_id", collegeID,
			"ip", requestIP,
			"error", err,
		)
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "error",
			Error:  "Unable to fetch college details at this time.",
		})
	}

	h.logger.Infow("College details retrieved successfully",
		"college_id", collegeID,
		"ip", requestIP,
	)

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "College details retrieved successfully.",
		Data:    college,
	})
}

func (h *SuperAdminHandler) DeleteCollege(c echo.Context) error {
	ctx := c.Request().Context()
	requestIP := c.RealIP()

	collegeID := strings.TrimSpace(c.Param("college_id"))
	if collegeID == "" {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "College ID is required to delete.",
		})
	}

	if err := h.service.DeleteCollege(ctx, collegeID); err != nil {
		if errors.Is(err, apperrors.ErrCollegeNotFound) {
			h.logger.Warnw("Attempted to delete non-existent college",
				"college_id", collegeID,
				"ip", requestIP,
			)
			return c.JSON(http.StatusNotFound, domain.ErrorResponse{
				Status: "error",
				Error:  "College not found.",
			})
		}
		h.logger.Errorw("Failed to delete college",
			"college_id", collegeID,
			"ip", requestIP,
			"error", err,
		)
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "error",
			Error:  "Unable to delete college at this time.",
		})
	}

	h.logger.Infow("College deleted successfully",
		"college_id", collegeID,
		"ip", requestIP,
	)

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "College deleted successfully.",
	})
}

func (h *SuperAdminHandler) RechargeCollege(c echo.Context) error {
	ctx := c.Request().Context()
	requestIP := c.RealIP()

	collegeID := strings.TrimSpace(c.Param("college_id"))
	if collegeID == "" {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "College ID is required for recharge.",
		})
	}

	superAdminID, ok := c.Get("super_admin_id").(string)
	if !ok || superAdminID == "" {
		h.logger.Warnw("Missing super admin ID in context during recharge",
			"college_id", collegeID,
			"ip", requestIP,
		)
		return c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
			Status: "error",
			Error:  "Unauthorized: super admin identity could not be verified.",
		})
	}

	var body struct {
		RechargeAmount string `json:"recharge_amount" validate:"required"`
	}

	if err := c.Bind(&body); err != nil {
		h.logger.Warnw("Failed to bind college recharge request",
			"college_id", collegeID,
			"ip", requestIP,
			"error", err,
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Invalid recharge request format.",
		})
	}

	if err := h.validate.Struct(body); err != nil {
		validationMsg := utils.FormatValidationError(err)
		h.logger.Warnw("College recharge validation failed",
			"college_id", collegeID,
			"validation_error", validationMsg,
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  validationMsg,
		})
	}

	req := model.CollegeRechargeRequest{
		CollegeID:      collegeID,
		RechargeAmount: body.RechargeAmount,
		SuperAdminId:   superAdminID,
	}

	if err := h.service.RechargeToCollege(ctx, req); err != nil {
		h.logger.Errorw("Failed to recharge college account",
			"college_id", collegeID,
			"super_admin_id", superAdminID,
			"amount", body.RechargeAmount,
			"ip", requestIP,
			"error", err,
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  err.Error(),
		})
	}

	h.logger.Infow("College account recharged successfully",
		"college_id", collegeID,
		"super_admin_id", superAdminID,
		"amount", body.RechargeAmount,
		"ip", requestIP,
	)

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "College account recharged successfully.",
	})
}

func (h *SuperAdminHandler) GetCollegeRechargeHistory(c echo.Context) error {
	ctx := c.Request().Context()
	requestIP := c.RealIP()

	collegeID := strings.TrimSpace(c.Param("college_id"))
	if collegeID == "" {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "College ID is required to fetch recharge history.",
		})
	}

	history, err := h.service.GetRechargeHistory(ctx, collegeID)
	if err != nil {
		h.logger.Errorw("Failed to fetch college recharge history",
			"college_id", collegeID,
			"ip", requestIP,
			"error", err,
		)
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "error",
			Error:  "Unable to fetch recharge history at this time.",
		})
	}

	h.logger.Infow("College recharge history retrieved successfully",
		"college_id", collegeID,
		"record_count", len(history),
		"ip", requestIP,
	)

	return c.JSON(http.StatusOK, model.CollegeRechargeHistoryResponse{
		Status:  "success",
		Message: "Recharge history retrieved successfully.",
		Data:    history,
	})
}

/*  Super Admin Machine Management Handlers  */

func (h *SuperAdminHandler) CreateMachineHandler(c echo.Context) error {
	ctx := c.Request().Context()
	requestIP := c.RealIP()

	superAdminID, ok := c.Get("super_admin_id").(string)
	if !ok || superAdminID == "" {
		h.logger.Warnw("Missing super admin ID in context during machine creation",
			"ip", requestIP,
		)
		return c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
			Status: "error",
			Error:  "Unauthorized: super admin identity could not be verified.",
		})
	}

	var payload model.MachineCreateRequest

	if err := utils.DecodeAndValidateJSON(c.Request().Body, &payload); err != nil {
		h.logger.Warnw("Failed to decode machine creation payload",
			"super_admin_id", superAdminID,
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

	if err := h.service.CreateMachineService(ctx, payload, superAdminID); err != nil {
		return h.handleCreateMachineError(c, err, payload.MachineNo)
	}

	h.logger.Infow("Machine created successfully",
		"machine_no", payload.MachineNo,
		"machine_name", payload.MachineName,
		"super_admin_id", superAdminID,
		"ip", requestIP,
	)

	return c.JSON(http.StatusCreated, domain.SuccessResponse{
		Status:  "success",
		Message: "Machine created successfully.",
	})
}

func (h *SuperAdminHandler) GetMachinesByCollegeID(c echo.Context) error {
	ctx := c.Request().Context()
	requestIP := c.RealIP()

	collegeID := strings.TrimSpace(c.Param("college_id"))
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
			Error:  "Unable to fetch machines at this time.",
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

/*  Error Handlers  */

func (h *SuperAdminHandler) handleLoginError(c echo.Context, err error, email string) error {
	switch {
	case errors.Is(err, apperrors.ErrSuperAdminNotFound):
		h.logger.Warnw("Login attempt for non-existent super admin",
			"email", email,
		)
		return c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
			Status: "error",
			Error:  "Invalid email or password.",
		})

	case errors.Is(err, apperrors.ErrInvalidPassword):
		h.logger.Warnw("Invalid password during super admin login",
			"email", email,
		)
		return c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
			Status: "error",
			Error:  "Invalid email or password.",
		})

	default:
		h.logger.Errorw("Unexpected error during super admin login",
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
		h.logger.Warnw("Attempted to create duplicate machine",
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