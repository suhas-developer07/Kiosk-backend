package cmd

import (
	"github.com/labstack/echo/v4"
	handler_Faculty "github.com/suhas-developer07/Kiosk-backend/src/internals/handlers/faculty_handler"
	handler_File "github.com/suhas-developer07/Kiosk-backend/src/internals/handlers/file_handler"
	handler_orchestrator "github.com/suhas-developer07/Kiosk-backend/src/internals/handlers/orchestrator"
	handler_Admin  "github.com/suhas-developer07/Kiosk-backend/src/internals/handlers/admin"
)

func SetupRouter(
	e *echo.Echo,
	fileHandler *handler_File.FileHandler,
	facultyHandler *handler_Faculty.FacultyHandler,
	orchestratorHandler *handler_orchestrator.OrchestrateHandler,
	adminHandler   *handler_Admin.AdminHandler,
	auth echo.MiddlewareFunc,
) {
	//public routes
	files := e.Group("/files")
	files.GET("/:grade/:subject", fileHandler.GetFilesByGradeAndSubjectHandler)
	files.POST("/printjob", fileHandler.PrintUploadHandler)
	files.GET("/accessfile/:file_id", fileHandler.AccessFileHandler)
	files.DELETE("/delete/:file_id",fileHandler.DeleteFileHandler)
	files.GET("/printjobs",fileHandler.FetchPrintJobsHandler)

	faculty := e.Group("/faculty")
	
	faculty.POST("/signin", facultyHandler.Signin)
	faculty.GET("/subjects", facultyHandler.GetAvailableSubjectsHandler) //todo:need to change this place
	//todo:need to change this all are admin routes
	faculty.GET("/faculties",adminHandler.GetFacultiesHandler)
	faculty.GET("/faculties/:stream",adminHandler.GetFacultiesByStreamHandler)
	faculty.PUT("/delete/:file_id",orchestratorHandler.FileDeleteRequestHandler)//todo:its need to be an private route

	admin := e.Group("/admin")

	admin.POST("/createfaculty", adminHandler.AddFacultyHandler)
	admin.GET("/getfaculties",adminHandler.GetFacultiesHandler)
	admin.GET("/facultiescount",adminHandler.GetTotalFacultiesCount)
	admin.GET("/getaculties/:stream",adminHandler.GetFacultiesByStreamHandler)
	admin.POST("/files/:file_id",adminHandler.FileDeleteDecisionHandler)
	admin.GET("/pendingdeletrequest",adminHandler.PendingDeleteRequestHandler)
	admin.GET("/recentactivity",adminHandler.RecentlyUploadedFilesHandler)


	fileAuth := files.Group("")
	fileAuth.Use(auth)
	facultyAuth := faculty.Group("")
	facultyAuth.Use(auth)

	//private routes
	fileAuth.POST("/upload", orchestratorHandler.UploadFileHandler)
	fileAuth.GET("/recentfiles", orchestratorHandler.GetRecentUploadedFilesHandler)

//	facultyAuth.PUT("/profileupdate", facultyHandler.UpdateProfile)
	facultyAuth.GET("/ownedsubjects", facultyHandler.GetSubjectsByFacultyIDHandler)
	facultyAuth.GET("/me",facultyHandler.GetFacultyByIdHandler)
}
