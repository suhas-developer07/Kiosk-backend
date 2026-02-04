package admin

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	facultymodel "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/faculties"
	domain "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/response"
	service "github.com/suhas-developer07/Kiosk-backend/src/internals/service/admin"
	"github.com/suhas-developer07/Kiosk-backend/src/pkg/utils"
	"go.uber.org/zap"
)

type AdminHandler struct {
	adminService *service.AdminService
	Logger       *zap.SugaredLogger
	validate     *validator.Validate
}

func NewAdminHandler(as *service.AdminService, Logger *zap.SugaredLogger) *AdminHandler {
	return &AdminHandler{
		adminService: as,
		Logger:       Logger,
		validate:     validator.New(),
	}
}

func (h *AdminHandler) GetFacultiesHandler(c echo.Context) error {
	ctx := c.Request().Context()

	h.Logger.Infow("get faculties request received")

	faculty, err := h.adminService.GetFacultiesService(ctx)
	if err != nil {
		h.Logger.Errorw(
			"failed to fetch faculties",
			"error", err,
		)

		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "error",
			Error:  "failed to fetch faculties",
		})
	}

	if faculty == nil {
		faculty = []facultymodel.Faculty{}
	}

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "faculties fetched successfully",
		Data:    faculty,
	})
}

func (h *AdminHandler) GetFacultiesByStreamHandler(c echo.Context) error {
	ctx := c.Request().Context()

	stream := strings.TrimSpace(strings.ToLower(c.Param("stream")))

	h.Logger.Infow(
		"get faculties by stream request received",
		"stream", stream,
	)

	if stream == "" {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "stream parameter is required",
		})
	}

	faculty, err := h.adminService.GetFacultiesByStreamService(ctx, stream)
	if err != nil {
		//todo
		//integrate specific error reponse feature by adding const errors in errro_faculty file

		h.Logger.Errorw(
			"failed to fetch faculties by stream",
			"stream", stream,
			"error", err,
		)

		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "error",
			Error:  "failed to fetch faculties for the given stream",
		})
	}

	if faculty == nil {
		faculty = []facultymodel.Faculty{}
	}

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "faculties fetched successfully",
		Data:    faculty,
	})
}

func (h *AdminHandler) AddFacultyHandler(c echo.Context) error {
	ctx := c.Request().Context()

	var req facultymodel.FacultyPayload

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Invalid Body",
		})
	}

	if err := h.validate.Struct(&req); err != nil {
		msg := utils.FormatValidationError(err)
		h.Logger.Warnf("Account validation failed | payload=%v | error=%v", req, msg)

		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  err.Error(),
		})
	}

	err := h.adminService.AddFacultyService(ctx, req)

	if err != nil {
		switch {
		case errors.Is(err, facultymodel.ErrEmailAlreadyExists):
			h.Logger.Warnf("Email already exists | email=%s", req.Email)
			return c.JSON(http.StatusConflict, domain.ErrorResponse{
				Status: "error",
				Error:  "Email already exists.",
			})

		default:
			h.Logger.Errorf("Failed to Add the faculty | email=%s | error=%v", req.Email, err)
			return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
				Status: "error",
				Error:  "Internal server error",
			})
		}
	}
	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "Faculty successfully Added ",
	})
}

func (h *AdminHandler) GetTotalFacultiesCount(c echo.Context) error {
	ctx := c.Request().Context()

	h.Logger.Infow("get total faculties count")

	count, err := h.adminService.GetTotalFacultiesCountService(ctx)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "error",
			Error:  "Internal server error" + err.Error(),
		})
	}

	if count == 0 {
		return c.JSON(http.StatusOK, domain.SuccessResponse{
			Status:  "success",
			Message: "Faculty count fetched",
			Data: map[string]int{
				"faculties": 0,
			},
		})
	}

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "Faculty count fetched",
		Data: map[string]uint32{
			"faculties": count,
		},
	})
}

func (h *AdminHandler) GetTotalFilesCountHandler(c echo.Context) error {
	ctx := c.Request().Context()

	h.Logger.Infow("get files couunt")

	count, err := h.adminService.GetTotalFilesCountService(ctx)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "error",
			Error:  "Internal server error" + err.Error(),
		})
	}

	if count == 0 {
		return c.JSON(http.StatusOK, domain.SuccessResponse{
			Status:  "success",
			Message: "files count fetched",
			Data: map[string]int{
				"faculties": 0,
			},
		})
	}

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "files count fetched",
		Data: map[string]int64{
			"faculties": count,
		},
	})
}
