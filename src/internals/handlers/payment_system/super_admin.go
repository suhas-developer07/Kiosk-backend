package paymentsystem

import (
	"context"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	model "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/payment_system"
	domain "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/response"
	service "github.com/suhas-developer07/Kiosk-backend/src/internals/service/payment_system"
	"go.uber.org/zap"
)

type SuperAdminHandler struct {
	service *service.SuperAdminService
	logger  *zap.SugaredLogger
}

func NewSuperAdminHandler(service *service.SuperAdminService, logger *zap.SugaredLogger) *SuperAdminHandler {
	return &SuperAdminHandler{
		service: service,
		logger:  logger,
	}
}

/*
super admin auth handlers still not implemented
*/

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

func (h *SuperAdminHandler) GetCollegesByAdmin(c echo.Context) error {
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
	collegeID := strings.TrimSpace(c.Param("college_id"))
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

	if err := h.service.RechargeCollege(c.Request().Context(), model.CollegeRechargeRequest{
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

func (h *SuperAdminHandler) GetRechargeHistory(c echo.Context) error {
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
