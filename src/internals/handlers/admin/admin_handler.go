package admin

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/suhas-developer07/Kiosk-backend/src/internals/domain/faculties"
	domain "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/response"
	service "github.com/suhas-developer07/Kiosk-backend/src/internals/service/admin"
	"go.uber.org/zap"
)

type AdminHandler struct {
	adminService *service.AdminService
	Logger       *zap.SugaredLogger
}

func NewAdminHandler(as *service.AdminService, Logger *zap.SugaredLogger) *AdminHandler {
	return &AdminHandler{
		adminService: as,
		Logger:       Logger,
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
		faculty = []faculties.Faculty{}
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
		faculty = []faculties.Faculty{}
	}

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "faculties fetched successfully",
		Data:    faculty,
	})
}
