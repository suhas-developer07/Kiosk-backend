package admin

import (
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
		Logger: Logger,
	}
}
