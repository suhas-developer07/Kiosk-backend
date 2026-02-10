package rechargemachine

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	apperrors "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/errors"
	model "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/recharge_machine"
	domain "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/response"
	service "github.com/suhas-developer07/Kiosk-backend/src/internals/service/recharge_machine"
	"github.com/suhas-developer07/Kiosk-backend/src/pkg/utils"
	"go.uber.org/zap"
)

type RechargeMachineHandler struct {
	RechargeMachineService *service.RechargeMachineService
	Logger                 *zap.SugaredLogger
	validate               validator.Validate
}

func NewRechargeMachineHandler(RechargeMachineService *service.RechargeMachineService, Logger *zap.SugaredLogger) *RechargeMachineHandler {
	return &RechargeMachineHandler{
		RechargeMachineService: RechargeMachineService,
		Logger:                 Logger,
		validate:               *validator.New(),
	}
}

func (h *RechargeMachineHandler) CreateMainAdminHandler(c echo.Context) error {
	ctx := c.Request().Context()

	var payload model.CreateMainAdminPayload

	if err := utils.DecodeAndValidateJSON(c.Request().Body, &payload); err != nil {
		h.Logger.Warnf("Invalid print payload | IP=%s | Error=%v", c.RealIP(), err)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  err.Error(),
		})
	}

	if err := h.validate.Struct(&payload); err != nil {
		msg := utils.FormatValidationError(err)
		h.Logger.Warnf("Account validation failed | payload=%v | error=%v", payload, msg)

		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  msg,
		})
	}

	err := h.RechargeMachineService.CreateAccountService(ctx, payload)
	if err != nil {

		switch {

		case errors.Is(err, apperrors.ErrEmailAlreadyExists):
			h.Logger.Warnf("Email already exists | email=%s", payload.Email)
			return c.JSON(http.StatusConflict, domain.ErrorResponse{
				Status: "error",
				Error:  "Email already exists.",
			})
		case errors.Is(err, apperrors.ErrInvalidCredentials):
			h.Logger.Warnf("Invalid credentials | email=%s | error=%v", payload.Email, err)
			return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
				Status: "error",
				Error:  "Invalid credentials.",
			})

		default:
			h.Logger.Errorf("Failed to create account | email=%s | error=%v", payload.Email, err)
			return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
				Status: "error",
				Error:  "Internal server error",
			})
		}
	}

	return c.JSON(http.StatusCreated, domain.SuccessResponse{
		Status:  "success",
		Message: "Account created successfully",
	})

}

func (h *RechargeMachineHandler) CreateMachineHandler(c echo.Context) error {
	ctx := c.Request().Context()

	var payload model.MachineCreateRequest

	if err := utils.DecodeAndValidateJSON(c.Request().Body, &payload); err != nil {
		h.Logger.Warnf("Invalid machine creation payload | IP=%s | Error=%v", c.RealIP(), err)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  err.Error(),
		})
	}

	if err := h.validate.Struct(&payload); err != nil {
		msg := utils.FormatValidationError(err)
		h.Logger.Warnf("Machine creation validation failed | payload=%v | error=%v", payload, msg)

		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  msg,
		})
	}

	err := h.RechargeMachineService.CreateMachineService(ctx, payload)
	if err != nil {

		switch {

		case errors.Is(err, apperrors.ErrMachineAlreadyExists):
			h.Logger.Warnf("Machine already exists | machineNo=%s", payload.MachineNo)
			return c.JSON(http.StatusConflict, domain.ErrorResponse{
				Status: "error",
				Error:  "A machine with this number already exists.",
			})

		default:
			h.Logger.Errorf("Failed to create machine | machineNo=%s | error=%v", payload.MachineNo, err)
			return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
				Status: "error",
				Error:  "Internal server error",
			})
		}
	}

	return c.JSON(http.StatusCreated, domain.SuccessResponse{
		Status:  "success",
		Message: "Machine created successfully",
	})
}

func (h *RechargeMachineHandler) RechargeMachineHadler(c echo.Context) error {
	var req model.MachineRechargeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "We could not process your request. Please make sure all details are filled in correctly.",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Some required details are missing or incorrect: " + strings.ReplaceAll(err.Error(), "\n", ", "),
		})
	}

	if err := h.RechargeMachineService.RechargeMachine(c.Request().Context(), req); err != nil {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Recharge failed: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "The machine has been recharged successfully.",
	})
}

func (h *RechargeMachineHandler) RechargeRFIDHandler(c echo.Context) error {

	machineID := c.Param("machine_id")
	userID := c.Param("user_id")

	if  machineID == "" || userID == "" {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Missing required path parameters: machine_id, user_id are mandatory.",
		})
	}

	var body struct {
		RechargeAmount string `json:"recharge_amount" validate:"required"`
	}
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Invalid request body. Please send a valid recharge_amount.",
		})
	}

	if body.RechargeAmount == "" {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Recharge amount is required.",
		})
	}

	req := model.RechargeRFIDRequest{
		MachineID:      machineID,
		UserID:         userID,
		RechargeAmount: body.RechargeAmount,
	}

	if err := h.validate.Struct(req); err != nil {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Some required details are missing or incorrect: " + strings.ReplaceAll(err.Error(), "\n", ", "),
		})
	}

	if err := h.RechargeMachineService.RechargeRFIDService(c.Request().Context(), req); err != nil {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Recharge failed: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "The student has been recharged successfully.",
	})
}

func (h *RechargeMachineHandler) GetRFIDRechargeHistoryHandler(c echo.Context) error {
	machineId := c.Param("machine_id")
	if strings.TrimSpace(machineId) == "" {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Please provide a valid machine id to view its recharge history.",
		})
	}

	userRechargeHistory, err := h.RechargeMachineService.GetRFIDRechargeHistoryService(c.Request().Context(), machineId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "We could not fetch the recharge history: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "Recharge history retrieved successfully.",
		Data:    userRechargeHistory,
	})
}

func (h *RechargeMachineHandler) GetMachineBalanceHandler(c echo.Context) error {
	machineID := c.Param("machine_id")

	if strings.TrimSpace(machineID) == "" {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Please provide a valid machine ID to view its Balance.",
		})
	}

	balance, err := h.RechargeMachineService.GetMachineBalanceService(c.Request().Context(), machineID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "We could not fetch the balance: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "Balance retrieved successfully.",
		Data:    balance,
	})
}

func (h *RechargeMachineHandler) GetRechargeMachineHistoryHandler(c echo.Context) error {
	machineID := c.Param("machine_id")

	if strings.TrimSpace(machineID) == "" {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Please provide a valid machine ID to view its recharge history.",
		})
	}

	rechargeHistory, err := h.RechargeMachineService.GetRechargeMachineHistoryService(c.Request().Context(), machineID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "We could not fetch the recharge history: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "Recharge history retrieved successfully.",
		Data:    rechargeHistory,
	})
}


//Warden Routes

func (h *RechargeMachineHandler) CreateUserHandler(c echo.Context) error {
	var req model.UserAccessCreateRequest
	fmt.Println("CreateUserHandler HIT")

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "We could not read your request. Please ensure all required fields are filled correctly.",
		})
	}

	fmt.Println("Received user creation request:", req)
	user, err := h.RechargeMachineService.CreateUserService(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Unable to create user: " + err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, domain.SuccessResponse{
		Status:  "success",
		Message: "User account created successfully.",
		Data:    user,
	})
}

func (h *RechargeMachineHandler) LoginUserHandler(c echo.Context) error {
	var req model.UserAccessLoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "We could not read your request. Please check your email and password format.",
		})
	}

	token, err := h.RechargeMachineService.LoginUserService(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Login failed: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "Login successful.",
		Data: map[string]string{
			"token": token,
		},
	})
}


// func (h *RechargeMachineHandler) GetAllUsers(c echo.Context) error {
// 	collegeId := strings.TrimSpace(c.Param("college_id"))
// 	if collegeId == "" {
// 		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
// 			Status: "error",
// 			Error:  "College ID is required to fetch users.",
// 		})
// 	}

// 	users, err := h.RechargeMachineService.GetAllUsers(c.Request().Context(), collegeId)
// 	if err != nil {
// 		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
// 			Status: "error",
// 			Error:  "Unable to fetch users: " + err.Error(),
// 		})
// 	}
// 	return c.JSON(http.StatusOK, domain.SuccessResponse{
// 		Status:  "success",
// 		Message: "User list retrieved successfully.",
// 		Data:    users,
// 	})
// }


// func (h *RechargeMachineHandler) DeleteUser(c echo.Context) error {
// 	userId := strings.TrimSpace(c.Param("user_id"))
// 	if userId == "" {
// 		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
// 			Status: "error",
// 			Error:  "User ID is required to delete an account.",
// 		})
// 	}

// 	err := h.RechargeMachineService.DeleteUserService(c.Request().Context(), userId)
// 	if err != nil {
// 		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
// 			Status: "error",
// 			Error:  "Unable to delete user: " + err.Error(),
// 		})
// 	}

// 	return c.JSON(http.StatusOK, domain.SuccessResponse{
// 		Status:  "success",
// 		Message: "User deleted successfully.",
// 	})
// }
