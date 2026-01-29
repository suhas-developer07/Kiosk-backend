package admin

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
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

	Faculties, err := h.adminService.GetFacultiesService(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "error",
			Error:  err.Error(),
		})
	}

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "Faculties fetched successfully",
		Data:    Faculties,
	})
}

func (h *AdminHandler) GetFacultiesByStreamHandler(c echo.Context) error {
	ctx := c.Request().Context()

	stream := strings.TrimSpace(strings.ToLower(c.Param("stream")))

	Faculties, err := h.adminService.GetFacultiesByStreamService(ctx, stream)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "error",
			Error:  err.Error(),
		})
	}

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "Faculties fetched successfully",
		Data:    Faculties,
	})

}
