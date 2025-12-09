package facultyhandler

import (
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	domain "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/faculties"
	service "github.com/suhas-developer07/Kiosk-backend/src/internals/service/faculty_service"
	"github.com/suhas-developer07/Kiosk-backend/src/pkg/utils"
	"go.uber.org/zap"
)

type FacultyHandler struct {
	Facultyservice *service.FacultyService
	Logger         *zap.SugaredLogger
	validate       *validator.Validate
}

func NewFacultyHandler(fs *service.FacultyService, Logger *zap.SugaredLogger) *FacultyHandler {
	return &FacultyHandler{
		Facultyservice: fs,
		Logger:         Logger,
		validate:       validator.New(),
	}
}

func (h *FacultyHandler) CreateAccount(c echo.Context) error {
	ctx := c.Request().Context()

	var payload domain.AccoutCreationPayload

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

	err := h.Facultyservice.CreateAccountService(ctx,payload)
	if err != nil {

		switch {

		case errors.Is(err, domain.ErrEmailAlreadyExists):
			h.Logger.Warnf("Email already exists | email=%s", payload.Email)
			return c.JSON(http.StatusConflict, domain.ErrorResponse{
				Status: "error",
				Error:  "Email already exists.",
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
