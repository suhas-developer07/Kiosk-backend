package cmd

import (
	"github.com/labstack/echo/v4"
	handler_Faculty "github.com/suhas-developer07/Kiosk-backend/src/internals/handlers/faculty_handler"
	handler_File "github.com/suhas-developer07/Kiosk-backend/src/internals/handlers/file_handler"
)

func SetupRouter(e *echo.Echo, fileHandler *handler_File.FileHandler,facultyHandler *handler_Faculty.FacultyHandler) {

	files := e.Group("/files")
	files.POST("/upload", fileHandler.UploadFileHandler)
	files.GET("/:grade/:subject", fileHandler.GetFilesByGradeAndSubjectHandler)
	files.POST("/printjob", fileHandler.PrintUploadHandler)

	faculty := e.Group("/faculty")
	faculty.POST("/signup",facultyHandler.CreateAccount)
}
