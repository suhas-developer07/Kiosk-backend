package cmd

import (
	"github.com/labstack/echo/v4"
	service_File "github.com/suhas-developer07/Kiosk-backend/src/internals/service"
	handler_File "github.com/suhas-developer07/Kiosk-backend/src/internals/handlers/File_Handler"
)



func SetupRouter(e *echo.Echo,FileService *service_File.FileService) {
	
	FileHandler := handler_File.NewFileHandler(FileService)

	Files := e.Group("/files")
	Files.POST("/upload",FileHandler.UploadFileHandler)

	//health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status": "ok",
		})
	})
}