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

// MainAdminHandler handles HTTP requests for recharge machine operations
type MainAdminHandler struct {
	service  *service.MainAdminService
	logger   *zap.SugaredLogger
	validate *validator.Validate
}

func NewMainAdminHandler(
	mainAdminService *service.MainAdminService,
	logger *zap.SugaredLogger,
) *MainAdminHandler {
	return &MainAdminHandler{
		service:  mainAdminService,
		logger:   logger,
		validate: validator.New(),
	}
}

/* College Auth handlers*/
func (h *MainAdminHandler) CollegeLoginRequestHandler(c echo.Context) error {
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

	res, err := h.service.CollegeLoginService(c.Request().Context(), req)
	if err != nil {
		h.logger.Warnw("College login failed",
			"error", err,
		)
		if errors.Is(err, apperrors.ErrInvalidPassword) {
			return c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
				Status: "error",
				Error:  "Invalid Credentials",
			})
		}
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

/* Recharge Machine Handlers */
func (h *MainAdminHandler) RechargeMachineHandler(c echo.Context) error {
	ctx := c.Request().Context()
	requestIP := c.RealIP()

	var req model.MachineRechargeRequest

	college_id := c.Get("college_id").(string)

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

	if err := h.service.RechargeMachine(ctx, req, college_id); err != nil {
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

func (h *MainAdminHandler) GetMachineBalanceHandler(c echo.Context) error {
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

func (h *MainAdminHandler) GetRechargeMachineHistoryHandler(c echo.Context) error {
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

func (h *MainAdminHandler) FetchConnectedMachinesHandler(c echo.Context) error {
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

func (h *MainAdminHandler) GetMachinesByCollegeID(c echo.Context) error {
	ctx := c.Request().Context()
	requestIP := c.RealIP()

	collegeID := c.Get("college_id").(string)
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

/* RFID Recharge Handlers */
func (h *MainAdminHandler) RechargeRFIDHandler(c echo.Context) error {
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
		CardID         string `json:"card_id" bson:"card_id" validate:"required"`
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

	req := model.RechargeRFIDRequest{
		MachineID:      machineID,
		UserID:         userID,
		CardID:         body.CardID,
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

func (h *MainAdminHandler) GetRFIDRechargeHistoryHandler(c echo.Context) error {
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

func (h *MainAdminHandler) InitializeCard(c echo.Context) error {
	ctx := c.Request().Context()
	requestIP := c.RealIP()

	type reqBody struct {
		CardID         string `json:"card_id" validate:"required"`
		USN            string `json:"usn" bson:"usn" validate:"required"`
		RechargeAmount string `json:"recharge_amount" bson:"recharge_amount" validate:"required"`
	}
	var req reqBody

	machineId := strings.ToLower(c.Param("machine_id"))
	userId := c.Get("user_id").(string)

	if machineId == "" || userId == "" {
		h.logger.Warnw("Missing required parameters for card initialization",
			"machine_id", machineId,
			"user_id", userId,
			"ip", requestIP,
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Missing required parameters: machine_id and user_id are mandatory.",
		})
	}

	if err := c.Bind(&req); err != nil {
		h.logger.Warnw("Failed to bind card initialization request",
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
		h.logger.Warnw("Card initialization validation failed",
			"card_id", req.CardID,
			"usn", req.USN,
			"validation_error", validationMsg,
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  validationMsg,
		})
	}

	data := model.InitializeCardRequest{
		MachineID:      machineId,
		CardID:         req.CardID,
		UserID:         userId,
		USN:            req.USN,
		RechargeAmount: req.RechargeAmount,
	}

	if err := h.service.InitializeCardService(ctx, data); err != nil {
		h.logger.Errorw("Card initialization failed",
			"card_id", req.CardID,
			"usn", req.USN,
			"error", err,
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  err.Error(),
		})
	}

	h.logger.Infow("Card initialized successfully",
		"card_id", req.CardID,
		"usn", req.USN,
		"ip", requestIP,
	)

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "Card has been initialized successfully.",
	})
}

func (h *MainAdminHandler) GetRFIDCardBalance(c echo.Context) error {
	ctx := c.Request().Context()
	cardID := strings.TrimSpace(c.Param("card_id"))

	if cardID == "" {
		h.logger.Warnw("Empty card ID for balance request",
			"ip", c.RealIP(),
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Please provide a valid card ID to view its balance.",
		})
	}

	balance, err := h.service.GetRFIDCardBalanceService(ctx, cardID)
	if err != nil {
		h.logger.Errorw("Failed to fetch RFID card balance",
			"card_id", cardID,
			"error", err,
		)
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "error",
			Error:  "Unable to fetch card balance at this time.",
		})
	}

	h.logger.Infow("RFID card balance retrieved successfully",
		"card_id", cardID,
		"balance", balance,
	)

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "Card balance retrieved successfully.",
		Data:    balance,
	})
}

func (h *MainAdminHandler) GetRFIDCardDetails(c echo.Context) error {
	ctx := c.Request().Context()
	cardID := strings.TrimSpace(c.Param("card_id"))

	if cardID == "" {
		h.logger.Warnw("Empty card ID for details request",
			"ip", c.RealIP(),
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Please provide a valid card ID to view its details.",
		})
	}

	cardDetails, err := h.service.GetRFIDCardDetailsService(ctx, cardID)
	if err != nil {
		h.logger.Errorw("Failed to fetch RFID card details",
			"card_id", cardID,
			"error", err,
		)
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "error",
			Error:  "Unable to fetch card details at this time.",
		})
	}

	h.logger.Infow("RFID card details retrieved successfully",
		"card_id", cardID,
	)

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "Card details retrieved successfully.",
		Data:    cardDetails,
	})
}

func (h *MainAdminHandler) CardDeativationHandler(c echo.Context) error {
	ctx := c.Request().Context()
	requestIP := c.RealIP()
	cardID := strings.TrimSpace(c.Param("card_id"))

	if cardID == "" {
		h.logger.Warnw("Empty card ID for deactivation request",
			"ip", requestIP,
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Please provide a valid card ID to deactivate.",
		})
	}

	var req struct {
		Status string `json:"status" validate:"required,oneof=active deactivated"`
	}
	if err := c.Bind(&req); err != nil {
		h.logger.Warnw("Failed to bind card deactivation request",
			"card_id", cardID,
			"error", err,
		)
	}
	err := h.service.CardDeactivationService(ctx, cardID, req.Status)
	if err != nil {
		h.logger.Errorw("Card deactivation failed",
			"card_id", cardID,
			"error", err,
		)
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "error",
			Error:  "Failed to deactivate card.",
		})
	}

	h.logger.Infow("Card deactivated successfully",
		"card_id", cardID,
	)

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "Card has been deactivated successfully.",
	})
}

func (h *MainAdminHandler) GetAllCardsHandler(c echo.Context) error {
	ctx := c.Request().Context()

	cards,err := h.service.GetAllCardsService(ctx)
	if err != nil {
		h.logger.Errorw("Failed to fetch all cards",
			"error", err,
		)
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "error",
			Error:  "Unable to fetch cards at this time.",
		})
	}

	h.logger.Infow("All cards retrieved successfully",
		"card_count", len(cards),
	)

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "Cards retrieved successfully.",
		Data:    cards,
	})
}

/*Recharge Machine Users Handler */
func (h *MainAdminHandler) CreateRechargeMachineUser(c echo.Context) error {
	ctx := c.Request().Context()
	requestIP := c.RealIP()

	var req model.UserAccessCreateRequest

	college_id := c.Get("college_id").(string)

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
	user, err := h.service.CreateRechargeMachineUserService(ctx, req, college_id)
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

func (h *MainAdminHandler) LoginRechargeMachineUser(c echo.Context) error {
	ctx := c.Request().Context()
	requestIP := c.RealIP()

	var req model.UserAccessLoginRequest

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
			"email", req.Email,
			"validation_error", validationMsg,
		)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  validationMsg,
		})
	}

	token, err := h.service.LoginRechargeMachineUserService(ctx, req)
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
/* Error Handling */
func (h *MainAdminHandler) handleLoginError(c echo.Context, err error, email string) error {
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

func (h *MainAdminHandler) GetMachineUsersByMachineIdHandler(c echo.Context) error {
	machineId := strings.TrimSpace(c.Param("machine_id"))
	if machineId == "" {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Machine ID is required to fetch users.",
		})
	}

	users, err := h.service.GetMachineUsersByMachineIdService(c.Request().Context(), machineId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Unable to fetch users: " + err.Error(),
		})
	}
	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "User list retrieved successfully.",
		Data:    users,
	})
}

func (h *MainAdminHandler) DeleteMachineUserHandler(c echo.Context) error {
	machineId := strings.TrimSpace(c.Param("machine_id"))
	userId := strings.TrimSpace(c.Param("user_id"))

	if machineId == "" || userId == "" {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Machine ID and User ID are required to delete an account.",
		})
	}

	err := h.service.DeleteMachineUserService(c.Request().Context(), machineId, userId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Unable to delete user: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "User deleted successfully.",
	})
}

// func (h *MainAdminHandler) GetAllUsers(c echo.Context) error {
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

// func (h *MainAdminHandler) DeleteUser(c echo.Context) error {
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
