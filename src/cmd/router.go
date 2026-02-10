package cmd

import (
	"github.com/labstack/echo/v4"
	handler_Admin "github.com/suhas-developer07/Kiosk-backend/src/internals/handlers/admin"
	handler_Faculty "github.com/suhas-developer07/Kiosk-backend/src/internals/handlers/faculty_handler"
	handler_File "github.com/suhas-developer07/Kiosk-backend/src/internals/handlers/file_handler"
	handler_orchestrator "github.com/suhas-developer07/Kiosk-backend/src/internals/handlers/orchestrator"
	handler_RechargeMachine "github.com/suhas-developer07/Kiosk-backend/src/internals/handlers/recharge_machine"
)

func SetupRouter(
	e *echo.Echo,
	fileHandler *handler_File.FileHandler,
	facultyHandler *handler_Faculty.FacultyHandler,
	orchestratorHandler *handler_orchestrator.OrchestrateHandler,
	adminHandler *handler_Admin.AdminHandler,
	machineHandler *handler_RechargeMachine.RechargeMachineHandler,
	faculty_auth echo.MiddlewareFunc,
	admin_auth echo.MiddlewareFunc,
	warden_auth echo.MiddlewareFunc,
) {
	//public routes
	files := e.Group("/files")
	files.GET("/:grade/:subject", fileHandler.GetFilesByGradeAndSubjectHandler)
	files.POST("/printjob", fileHandler.PrintUploadHandler)
	files.GET("/accessfile/:file_id", fileHandler.AccessFileHandler)
	files.GET("/printjobs", fileHandler.FetchPrintJobsHandler)

	faculty := e.Group("/faculty")

	faculty.POST("/signin", facultyHandler.Signin)
	//todo:need to change this all are admin routes
	faculty.GET("/faculties", adminHandler.GetFacultiesHandler)
	faculty.GET("/faculties/:stream", adminHandler.GetFacultiesByStreamHandler)
	faculty.PUT("/delete/:file_id", orchestratorHandler.FileDeleteRequestHandler) //todo:its need to be an private route

	admin := e.Group("/admin")

	adminAuth := admin.Group("")
	adminAuth.Use(admin_auth)

	adminAuth.POST("/create-faculty", adminHandler.AddFacultyHandler)
	adminAuth.GET("/get-faculties", adminHandler.GetFacultiesHandler)
	adminAuth.GET("/total-faculties-count", adminHandler.GetTotalFacultiesCount)
	adminAuth.GET("/getaculties/:stream", adminHandler.GetFacultiesByStreamHandler)
	adminAuth.POST("/files/:file_id", adminHandler.FileDeleteDecisionHandler)
	adminAuth.GET("/pending-delete-request-count", adminHandler.PendingDeleteRequestCountHandler)
	adminAuth.GET("/recent-activity", adminHandler.RecentlyUploadedFilesHandler)
	adminAuth.GET("/total-files", adminHandler.GetTotalFilesHandler)
	adminAuth.GET("/total-files-count", adminHandler.GetTotalFilesCountHandler)
	adminAuth.GET("/pending-delete-request-files", adminHandler.GetPendingDeleteRequestFilesHandler)
	adminAuth.DELETE("/files/delete/:file_id", adminHandler.DeleteFileHandler)
	admin.GET("/subjects", adminHandler.GetAvailableSubjectsHandler)
	admin.POST("/signup", adminHandler.CreateAccount)
	admin.POST("/signin", adminHandler.Signin)

	fileAuth := files.Group("")
	fileAuth.Use(faculty_auth)
	facultyAuth := faculty.Group("")
	facultyAuth.Use(faculty_auth)

	//private routes
	fileAuth.POST("/upload", orchestratorHandler.UploadFileHandler)
	fileAuth.POST("/upload/initiate", orchestratorHandler.InitiateUploadHandler)
	fileAuth.POST("/upload/complete", orchestratorHandler.CompleteUploadHandler)

	fileAuth.GET("/recentfiles", orchestratorHandler.GetRecentUploadedFilesHandler)

	//	facultyAuth.PUT("/profileupdate", facultyHandler.UpdateProfile)
	facultyAuth.GET("/ownedsubjects", facultyHandler.GetSubjectsByFacultyIDHandler)
	facultyAuth.GET("/me", facultyHandler.GetFacultyByIdHandler)

	//Recharge Machine Routes
	machine := e.Group("/machine")
	machine.POST("/super-admin/signup", machineHandler.CreateMainAdminHandler)
	machine.POST("/create-machine", machineHandler.CreateMachineHandler)
	machine.POST("/recharge", machineHandler.RechargeMachineHadler)
	machine.GET("/fetch-machine-balance/:machine_id", machineHandler.GetMachineBalanceHandler)
	machine.GET("/fetch-machine-recharge-history/:machine_id", machineHandler.GetRechargeMachineHistoryHandler)

	// machineAuth := machine.Group("")
	// machineAuth.Use(warden_auth)

	machine.POST("/recharge/:machine_id/:user_id", machineHandler.RechargeRFIDHandler)
	machine.GET("/recharge-history/:machine_id", machineHandler.GetRFIDRechargeHistoryHandler)
	machine.GET("/FetchConnectedMachines/:machine_no", machineHandler.FetchConnectedMachinesHandler)

	//warden routes
	warden := e.Group("/warden")
	warden.POST("/create-user", machineHandler.CreateUserHandler)
	warden.POST("/login-user", machineHandler.LoginUserHandler)

}
