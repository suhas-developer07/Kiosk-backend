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
	FacultyService *service.FacultyService
	Logger         *zap.SugaredLogger
	validate       *validator.Validate
}

func NewFacultyHandler(fs *service.FacultyService, Logger *zap.SugaredLogger) *FacultyHandler {
	return &FacultyHandler{
		FacultyService: fs,
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

	err := h.FacultyService.CreateAccountService(ctx, payload)
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

func (h *FacultyHandler) Signin(c echo.Context) error {
	ctx := c.Request().Context()

	var payload domain.SigninPayload

	if err := utils.DecodeAndValidateJSON(c.Request().Body, &payload); err != nil {
		h.Logger.Warnf("Invalid signin payload | Error=%v", err)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  err.Error(),
		})
	}

	if err := h.validate.Struct(&payload); err != nil {
		msg := utils.FormatValidationError(err)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  msg,
		})
	}

	access, refresh, err := h.FacultyService.SigninService(ctx, payload)
	if err != nil {

		switch {

		case errors.Is(err, domain.ErrUserNotFound):
			return c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
				Status: "error",
				Error:  "Invalid email or password",
			})

		case errors.Is(err, domain.ErrInvalidPassword):
			return c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
				Status: "error",
				Error:  "Invalid email or password",
			})

		default:
			h.Logger.Errorf("Signin failed | error=%v", err)
			return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
				Status: "error",
				Error:  "Internal server error",
			})
		}
	}

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "Signin successful",
		Data: map[string]string{
			"access_token":  access,
			"refresh_token": refresh,
		},
	})
}

func (h *FacultyHandler) UpdateProfile(c echo.Context) error {
	ctx := c.Request().Context()

	FacultyID := c.Get("faculty_id").(string)

	var payload domain.UpdateProfilePayload

	if err := utils.DecodeAndValidateJSON(c.Request().Body, &payload); err != nil {
		h.Logger.Warnf("Invalid profile payload | error=%v", err)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  err.Error(),
		})
	}

	if err := h.validate.Struct(&payload); err != nil {
		msg := utils.FormatValidationError(err)
		h.Logger.Warnf("Profile validation failed | error=%v", msg, err)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  msg,
		})
	}

	err := h.FacultyService.UpdateProfileService(ctx, FacultyID, payload)
	if err != nil {

		switch {

		case errors.Is(err, domain.ErrInvalidID):
			return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
				Status: "error",
				Error:  "Invalid user ID",
			})

		case errors.Is(err, domain.ErrUserNotFound):
			return c.JSON(http.StatusNotFound, domain.ErrorResponse{
				Status: "error",
				Error:  "User not found",
			})

		default:
			h.Logger.Errorf("Profile update failed | error=%v", err)
			return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
				Status: "error",
				Error:  "Internal server error",
			})
		}
	}

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "Profile updated successfully",
	})
}

//This handler used when he is configuring the file details to upload 
func (h *FacultyHandler) GetSubjectsByFacultyIDHandler(c echo.Context) error {
	ctx := c.Request().Context()

	facultyID := c.Get("faculty_id").(string)

	h.Logger.Infow(
		"get subjects by faculty id request received",
		"faculty_id", facultyID,
	)

	if facultyID == "" {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "faculty_id is required",
		})
	}

	subjectsList, err := h.FacultyService.GetSubjectsByFacultyID(ctx, facultyID)
	if err != nil {

		switch {
		case errors.Is(err, domain.ErrInvalidID):
			return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
				Status: "error",
				Error:  "invalid faculty id",
			})

		case errors.Is(err, domain.ErrFacultyNotFound):
			return c.JSON(http.StatusNotFound, domain.ErrorResponse{
				Status: "error",
				Error:  "faculty not found",
			})

		default:
			h.Logger.Errorw(
				"failed to fetch subjects for faculty",
				"faculty_id", facultyID,
				"error", err,
			)

			return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
				Status: "error",
				Error:  "failed to fetch subjects",
			})
		}
	}

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "subjects fetched successfully",
		Data: map[string]interface{}{
			"subjects": subjectsList,
		},
	})
}

func (h *FacultyHandler) GetAvailableSubjectsHandler(c echo.Context) error {
	ctx := c.Request().Context()

	h.Logger.Infow("get available subjects request received")

	subjectsList, err := h.FacultyService.GetAvailableSubjects(ctx)
	if err != nil {
		h.Logger.Errorw("failed to fetch available subjects", "error", err)

		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "error",
			Error:  "failed to fetch subjects",
		})
	}

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "available subjects fetched successfully",
		Data: map[string]interface{}{
			"subjects": subjectsList,
		},
	})
}

