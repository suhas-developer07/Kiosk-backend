package cmd

import (
	"github.com/labstack/echo/v4"
	handler_File "github.com/suhas-developer07/Kiosk-backend/src/internals/handlers/File_Handler"
)


func SetupRouter(e *echo.Echo, fileHandler *handler_File.FileHandler) {

	files := e.Group("/files")
	files.POST("/upload", fileHandler.UploadFileHandler)
	files.GET("/:grade/:subject", fileHandler.GetFilesByGradeAndSubjectHandler)
}
