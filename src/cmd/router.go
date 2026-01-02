package cmd

import (
	"github.com/labstack/echo/v4"
	handler_Faculty "github.com/suhas-developer07/Kiosk-backend/src/internals/handlers/faculty_handler"
	handler_File "github.com/suhas-developer07/Kiosk-backend/src/internals/handlers/file_handler"
)

func SetupRouter(
	e *echo.Echo, 
	fileHandler *handler_File.FileHandler,
	facultyHandler *handler_Faculty.FacultyHandler,
	auth echo.MiddlewareFunc,
	) {

	files := e.Group("/files")
	files.GET("/:grade/:subject", fileHandler.GetFilesByGradeAndSubjectHandler)
	files.POST("/printjob", fileHandler.PrintUploadHandler)
	files.GET("/accessfile/:file_id",fileHandler.AccessFileHandler)

	faculty := e.Group("/faculty")
	faculty.POST("/signup",facultyHandler.CreateAccount)
	faculty.POST("/signin",facultyHandler.Signin)
	faculty.GET("/subjects",facultyHandler.GetAvailableSubjectsHandler)

	fileAuth := files.Group("")
	fileAuth.Use(auth)

	fileAuth.POST("/upload",fileHandler.UploadFileHandler)

	facultyAuth := faculty.Group("")   
	facultyAuth.Use(auth)              

	facultyAuth.PUT("/profileupdate", facultyHandler.UpdateProfile)
	facultyAuth.GET("/ownedsubjects",facultyHandler.GetSubjectsByFacultyIDHandler)
}
