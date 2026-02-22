package cmd

import (
	"github.com/labstack/echo/v4"
	handler_Admin "github.com/suhas-developer07/Kiosk-backend/src/internals/handlers/admin"
	handler_Faculty "github.com/suhas-developer07/Kiosk-backend/src/internals/handlers/faculty_handler"
	handler_File "github.com/suhas-developer07/Kiosk-backend/src/internals/handlers/file_handler"
	handler_orchestrator "github.com/suhas-developer07/Kiosk-backend/src/internals/handlers/orchestrator"
	payment_handler "github.com/suhas-developer07/Kiosk-backend/src/internals/handlers/payment_system"
)

func SetupRouter(
	e *echo.Echo,
	fileHandler *handler_File.FileHandler,
	facultyHandler *handler_Faculty.FacultyHandler,
	orchestratorHandler *handler_orchestrator.OrchestrateHandler,
	adminHandler *handler_Admin.AdminHandler,
	mainAdminHandler *payment_handler.MainAdminHandler,
	superAdminHandler *payment_handler.SuperAdminHandler,
	faculty_auth echo.MiddlewareFunc,
	admin_auth echo.MiddlewareFunc,
	machine_user_auth echo.MiddlewareFunc,
	main_admin_auth echo.MiddlewareFunc,
) {
	/* Files Handler routes */
	files := e.Group("/files")
	files.GET("/:grade/:subject", fileHandler.GetFilesByGradeAndSubjectHandler)
	files.POST("/printjob", fileHandler.PrintUploadHandler)
	files.GET("/accessfile/:file_id", fileHandler.AccessFileHandler)
	files.GET("/printjobs", fileHandler.FetchPrintJobsHandler)

	/*Faculty Handler routes*/
	faculty := e.Group("/faculty")
	faculty.POST("/signin", facultyHandler.Signin)
	faculty.GET("/faculties", adminHandler.GetFacultiesHandler) //todo:need to change this all are admin routes
	faculty.GET("/faculties/:stream", adminHandler.GetFacultiesByStreamHandler)
	faculty.PUT("/delete/:file_id", orchestratorHandler.FileDeleteRequestHandler) //todo:its need to be an private route

	/* Private routes with faculty auth middleware */
	fileAuth := files.Group("")
	fileAuth.Use(faculty_auth)
	facultyAuth := faculty.Group("")
	facultyAuth.Use(faculty_auth)

	fileAuth.POST("/upload", orchestratorHandler.UploadFileHandler)
	fileAuth.POST("/upload/initiate", orchestratorHandler.InitiateUploadHandler)
	fileAuth.POST("/upload/complete", orchestratorHandler.CompleteUploadHandler)
	fileAuth.GET("/recentfiles", orchestratorHandler.GetRecentUploadedFilesHandler)
	facultyAuth.GET("/ownedsubjects", facultyHandler.GetSubjectsByFacultyIDHandler)
	facultyAuth.GET("/me", facultyHandler.GetFacultyByIdHandler)

	/* Admin Handler routes */
	admin := e.Group("/admin")

	admin.POST("/signup", adminHandler.CreateAccount)
	admin.POST("/signin", adminHandler.Signin)
	admin.GET("/subjects", adminHandler.GetAvailableSubjectsHandler)

	/* Admin private routes with admin auth middleware */
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

	
	/*
	 Paymemt system routes
	*/

	mainAdmin := e.Group("/main-admin")
	mainAdminAuth := mainAdmin.Group("")
	mainAdminAuth.Use(main_admin_auth)

	mainAdmin.POST("/signup", mainAdminHandler.CreateMainAdminHandler)
	mainAdmin.POST("/signin", mainAdminHandler.LoginMainAdminHandler)
	mainAdminAuth.POST("/create-machine", mainAdminHandler.CreateMachineHandler)
	mainAdminAuth.POST("/recharge-machine", mainAdminHandler.RechargeMachineHandler)
	mainAdminAuth.GET("/fetch-machine-balance/:machine_id", mainAdminHandler.GetMachineBalanceHandler)
	mainAdminAuth.GET("/fetch-machine-recharge-history/:machine_id", mainAdminHandler.GetRechargeMachineHistoryHandler)
	mainAdminAuth.GET("/fetch-available-machines", mainAdminHandler.GetAvailableMachinesHandler)
	mainAdminAuth.GET("/recharge-history/:machine_id", mainAdminHandler.GetRFIDRechargeHistoryHandler)

	/*machine user routes*/
	machineUser := e.Group("/machine-user")
	machineUser.POST("/create-user", mainAdminHandler.CreateUserHandler)
	machineUser.POST("/login-user", mainAdminHandler.LoginUserHandler)

	machineUserAuth := machineUser.Group("")
	machineUserAuth.Use(machine_user_auth)
	machineUserAuth.POST("/recharge-rfid/:machine_id", mainAdminHandler.RechargeRFIDHandler) //need to inform frontend team
	machineUserAuth.GET("/fetchConnectedMachines/:machine_no", mainAdminHandler.FetchConnectedMachinesHandler)
	machineUserAuth.GET("/fetchMachineBalance/:machine_id", mainAdminHandler.GetMachineBalanceHandler)
	machineUserAuth.GET("/recharge-history/:machine_id", mainAdminHandler.GetRFIDRechargeHistoryHandler)
}
