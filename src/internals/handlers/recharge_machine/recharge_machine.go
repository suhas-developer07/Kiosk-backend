package rechargemachine

import (
	"errors"
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

// RechargeMachineHandler handles HTTP requests for recharge machine operations
type RechargeMachineHandler struct {
	service  *service.RechargeMachineService
	logger   *zap.SugaredLogger
	validate *validator.Validate
}

func NewRechargeMachineHandler(
	rechargeMachineService *service.RechargeMachineService,
	logger *zap.SugaredLogger,
) *RechargeMachineHandler {
	return &RechargeMachineHandler{
		service:  rechargeMachineService,
		logger:   logger,
		validate: validator.New(),
	}
}

// CreateMainAdminHandler handles the creation of main admin accounts
func (h *RechargeMachineHandler) CreateMainAdminHandler(c echo.Context) error {
	ctx := c.Request().Context()
	requestIP := c.RealIP()

	var payload model.CreateMainAdminPayload

	if err := utils.DecodeAndValidateJSON(c.Request().Body, &payload); err != nil {
		h.logger.Warnw("Failed to decode admin creation payload",
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
		h.logger.Warnw("Admin creation validation failed",
			"email", payload.Email,
			"validation_error", validationMsg,
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  validationMsg,
		})
	}

	if err := h.service.CreateAccountService(ctx, payload); err != nil {
		return h.handleCreateAccountError(c, err, payload.Email)
	}

	h.logger.Infow("Admin account created successfully",
		"email", payload.Email,
		"ip", requestIP,
	)

	return c.JSON(http.StatusCreated, domain.SuccessResponse{
		Status:  "success",
		Message: "Account created successfully",
	})
}

// CreateMachineHandler handles the creation of new machines
func (h *RechargeMachineHandler) CreateMachineHandler(c echo.Context) error {
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

// RechargeMachineHandler handles machine balance recharge requests
func (h *RechargeMachineHandler) RechargeMachineHandler(c echo.Context) error {
	ctx := c.Request().Context()
	requestIP := c.RealIP()

	var req model.MachineRechargeRequest

	if err := c.Bind(&req); err != nil {
		h.logger.Warnw("Failed to bind machine recharge request",
			"ip", requestIP,
			"error", err,
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Invalid request format. Please ensure all fields are correctly formatted.",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		validationMsg := utils.FormatValidationError(err)
		h.logger.Warnw("Machine recharge validation failed",
			"machine_id", req.MachineID,
			"validation_error", validationMsg,
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  validationMsg,
		})
	}

	if err := h.service.RechargeMachine(ctx, req); err != nil {
		h.logger.Errorw("Machine recharge failed",
			"machine_id", req.MachineID,
			"amount", req.RechargeAmount,
			"error", err,
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  err.Error(),
		})
	}

	h.logger.Infow("Machine recharged successfully",
		"machine_id", req.MachineID,
		"amount", req.RechargeAmount,
		"ip", requestIP,
	)

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "The machine has been recharged successfully.",
	})
}

// RechargeRFIDHandler handles RFID-based recharge requests
func (h *RechargeMachineHandler) RechargeRFIDHandler(c echo.Context) error {
	ctx := c.Request().Context()
	requestIP := c.RealIP()

	machineID := strings.TrimSpace(c.Param("machine_id"))

	userID := c.Get("user_id").(string)

	if machineID == "" || userID == "" {
		h.logger.Warnw("Missing required path parameters for RFID recharge",
			"machine_id", machineID,
			"user_id", userID,
			"ip", requestIP,
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Missing required parameters: machine_id and user_id are mandatory.",
		})
	}

	var body struct {
		RechargeAmount string `json:"recharge_amount" validate:"required"`
	}
	if err := c.Bind(&body); err != nil {
		h.logger.Warnw("Failed to bind RFID recharge request body",
			"machine_id", machineID,
			"user_id", userID,
			"error", err,
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Invalid request body. Please provide a valid recharge_amount.",
		})
	}

	// Validate recharge amount
	if strings.TrimSpace(body.RechargeAmount) == "" {
		h.logger.Warnw("Empty recharge amount in RFID recharge request",
			"machine_id", machineID,
			"user_id", userID,
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Recharge amount is required.",
		})
	}

	// Build request object
	req := model.RechargeRFIDRequest{
		MachineID:      machineID,
		UserID:         userID,
		RechargeAmount: body.RechargeAmount,
	}

	if err := h.validate.Struct(req); err != nil {
		validationMsg := utils.FormatValidationError(err)
		h.logger.Warnw("RFID recharge validation failed",
			"machine_id", machineID,
			"user_id", userID,
			"validation_error", validationMsg,
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  validationMsg,
		})
	}

	if err := h.service.RechargeRFIDService(ctx, req); err != nil {
		h.logger.Errorw("RFID recharge failed",
			"machine_id", machineID,
			"user_id", userID,
			"amount", body.RechargeAmount,
			"error", err,
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  err.Error(),
		})
	}

	h.logger.Infow("RFID recharge completed successfully",
		"machine_id", machineID,
		"user_id", userID,
		"amount", body.RechargeAmount,
		"ip", requestIP,
	)

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "The student has been recharged successfully.",
	})
}

// GetRFIDRechargeHistoryHandler retrieves RFID recharge history for a machine
func (h *RechargeMachineHandler) GetRFIDRechargeHistoryHandler(c echo.Context) error {
	ctx := c.Request().Context()
	machineID := strings.TrimSpace(c.Param("machine_id"))

	if machineID == "" {
		h.logger.Warnw("Empty machine ID for RFID recharge history request",
			"ip", c.RealIP(),
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Please provide a valid machine ID to view recharge history.",
		})
	}

	history, err := h.service.GetRFIDRechargeHistoryService(ctx, machineID)
	if err != nil {
		h.logger.Errorw("Failed to fetch RFID recharge history",
			"machine_id", machineID,
			"error", err,
		)
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "error",
			Error:  "Unable to fetch recharge history at this time.",
		})
	}

	h.logger.Infow("RFID recharge history retrieved successfully",
		"machine_id", machineID,
		"record_count", len(history),
	)

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "Recharge history retrieved successfully.",
		Data:    history,
	})
}

// GetMachineBalanceHandler retrieves the current balance of a machine
func (h *RechargeMachineHandler) GetMachineBalanceHandler(c echo.Context) error {
	ctx := c.Request().Context()
	machineID := strings.TrimSpace(c.Param("machine_id"))

	if machineID == "" {
		h.logger.Warnw("Empty machine ID for balance request",
			"ip", c.RealIP(),
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Please provide a valid machine ID to view its balance.",
		})
	}

	balance, err := h.service.GetMachineBalanceService(ctx, machineID)
	if err != nil {
		h.logger.Errorw("Failed to fetch machine balance",
			"machine_id", machineID,
			"error", err,
		)
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "error",
			Error:  "Unable to fetch machine balance at this time.",
		})
	}

	h.logger.Infow("Machine balance retrieved successfully",
		"machine_id", machineID,
		"balance", balance.Balance,
	)

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "Balance retrieved successfully.",
		Data:    balance,
	})
}

// GetRechargeMachineHistoryHandler retrieves machine recharge history
func (h *RechargeMachineHandler) GetRechargeMachineHistoryHandler(c echo.Context) error {
	ctx := c.Request().Context()
	machineID := strings.TrimSpace(c.Param("machine_id"))

	if machineID == "" {
		h.logger.Warnw("Empty machine ID for recharge history request",
			"ip", c.RealIP(),
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Please provide a valid machine ID to view its recharge history.",
		})
	}

	history, err := h.service.GetRechargeMachineHistoryService(ctx, machineID)
	if err != nil {
		h.logger.Errorw("Failed to fetch machine recharge history",
			"machine_id", machineID,
			"error", err,
		)
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "error",
			Error:  "Unable to fetch recharge history at this time.",
		})
	}

	h.logger.Infow("Machine recharge history retrieved successfully",
		"machine_id", machineID,
		"record_count", len(history),
	)

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "Recharge history retrieved successfully.",
		Data:    history,
	})
}

// FetchConnectedMachinesHandler retrieves machine details by machine number
func (h *RechargeMachineHandler) FetchConnectedMachinesHandler(c echo.Context) error {
	ctx := c.Request().Context()
	machineNo := strings.TrimSpace(c.Param("machine_no"))

	if machineNo == "" {
		h.logger.Warnw("Empty machine number for connected machines request",
			"ip", c.RealIP(),
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Please provide a valid machine number to view connected machines.",
		})
	}

	machine, err := h.service.FetchConnectedMachinesService(ctx, machineNo)
	if err != nil {
		h.logger.Errorw("Failed to fetch connected machines",
			"machine_no", machineNo,
			"error", err,
		)
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "error",
			Error:  "Unable to fetch connected machines at this time.",
		})
	}

	h.logger.Infow("Connected machines retrieved successfully",
		"machine_no", machineNo,
	)

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "Connected machines retrieved successfully.",
		Data:    machine,
	})
}

// GetAvailableMachinesHandler retrieves all available machines
func (h *RechargeMachineHandler) GetAvailableMachinesHandler(c echo.Context) error {
	ctx := c.Request().Context()

	machines, err := h.service.GetAvailableMachinesService(ctx)
	if err != nil {
		h.logger.Errorw("Failed to fetch available machines",
			"error", err,
		)
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "error",
			Error:  "Unable to fetch available machines at this time.",
		})
	}

	h.logger.Infow("Available machines retrieved successfully",
		"machine_count", len(machines),
	)

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "Machines retrieved successfully.",
		Data:    machines,
	})
}

// CreateUserHandler handles warden user creation
func (h *RechargeMachineHandler) CreateUserHandler(c echo.Context) error {
	ctx := c.Request().Context()
	requestIP := c.RealIP()

	var req model.UserAccessCreateRequest

	// Bind request payload
	if err := c.Bind(&req); err != nil {
		h.logger.Warnw("Failed to bind user creation request",
			"ip", requestIP,
			"error", err,
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Invalid request format. Please ensure all required fields are provided.",
		})
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		validationMsg := utils.FormatValidationError(err)
		h.logger.Warnw("User creation validation failed",
			"email", req.Email,
			"validation_error", validationMsg,
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  validationMsg,
		})
	}

	// Call service layer
	user, err := h.service.CreateUserService(ctx, req)
	if err != nil {
		h.logger.Errorw("User creation failed",
			"email", req.Email,
			"machine_id", req.MachineId,
			"error", err,
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  err.Error(),
		})
	}

	h.logger.Infow("User created successfully",
		"user_id", user.UserID,
		"email", req.Email,
		"machine_id", req.MachineId,
		"ip", requestIP,
	)

	return c.JSON(http.StatusCreated, domain.SuccessResponse{
		Status:  "success",
		Message: "User account created successfully.",
		Data:    user,
	})
}

// LoginUserHandler handles warden user login
func (h *RechargeMachineHandler) LoginUserHandler(c echo.Context) error {
	ctx := c.Request().Context()
	requestIP := c.RealIP()

	var req model.UserAccessLoginRequest

	// Bind request payload
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

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		validationMsg := utils.FormatValidationError(err)
		h.logger.Warnw("Login validation failed",
			"email", req.Email,
			"validation_error", validationMsg,
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  validationMsg,
		})
	}

	// Call service layer
	token, err := h.service.LoginUserService(ctx, req)
	if err != nil {
		return h.handleLoginError(c, err, req.Email)
	}

	h.logger.Infow("User logged in successfully",
		"email", req.Email,
		"ip", requestIP,
	)

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "Login successful.",
		Data: map[string]string{
			"token": token,
		},
	})
}

// handleCreateAccountError handles errors from account creation service
func (h *RechargeMachineHandler) handleCreateAccountError(c echo.Context, err error, email string) error {
	switch {
	case errors.Is(err, apperrors.ErrEmailAlreadyExists):
		h.logger.Warnw("Email already exists during account creation",
			"email", email,
		)
		return c.JSON(http.StatusConflict, domain.ErrorResponse{
			Status: "error",
			Error:  "An account with this email already exists.",
		})

	case errors.Is(err, apperrors.ErrInvalidCredentials):
		h.logger.Warnw("Invalid credentials provided during account creation",
			"email", email,
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Invalid credentials provided.",
		})

	default:
		h.logger.Errorw("Unexpected error during account creation",
			"email", email,
			"error", err,
		)
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "error",
			Error:  "Unable to create account at this time. Please try again later.",
		})
	}
}

// handleCreateMachineError handles errors from machine creation service
func (h *RechargeMachineHandler) handleCreateMachineError(c echo.Context, err error, machineNo string) error {
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

// handleLoginError handles errors from login service
func (h *RechargeMachineHandler) handleLoginError(c echo.Context, err error, email string) error {
	switch {
	case errors.Is(err, apperrors.ErrFacultyNotFound):
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
