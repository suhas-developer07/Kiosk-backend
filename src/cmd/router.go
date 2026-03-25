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
	college_auth echo.MiddlewareFunc,
	super_admin_auth echo.MiddlewareFunc,
) {
	/* Files Handler routes */
	files := e.Group("/files")
	files.GET("/:grade/:subject", fileHandler.GetFilesByGradeAndSubjectHandler)
	files.GET("/accessfile/:file_id", fileHandler.AccessFileHandler)
	files.POST("/printjobs", fileHandler.CreatePrintJobsHandler)
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
	facultyAuth.GET("/pending-delete-requests", orchestratorHandler.GetPendingDeleteRequests)
	/* Admin Handler routes */
	admin := e.Group("/admin")
	adminAuth := admin.Group("")
	adminAuth.Use(admin_auth)

	admin.POST("/signup", adminHandler.CreateAccount)
	admin.POST("/signin", adminHandler.Signin)
	admin.GET("/subjects", adminHandler.GetAvailableSubjectsHandler)
	adminAuth.GET("/me", adminHandler.GetProfileDataHandler)
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
	adminAuth.GET("/pending-delete-request-count", adminHandler.PendingDeleteRequestCountHandler)
	adminAuth.DELETE("/files/delete/:file_id", adminHandler.DeleteFileHandler)
	adminAuth.DELETE("/faculty/delete/:faculty_id", adminHandler.DeleteFacultyHandler)
	adminAuth.PUT("/faculty/update/:faculty_id",adminHandler.UpdateFacultyHandler)
	/*
	 Paymemt system routes
	*/

	mainAdmin := e.Group("/main-admin")
	mainAdminAuth := mainAdmin.Group("")
	mainAdminAuth.Use(college_auth)

	mainAdmin.POST("/college/signin", mainAdminHandler.CollegeLoginRequestHandler)
	mainAdminAuth.GET("/me", mainAdminHandler.GetProfileDataHandler)
	mainAdminAuth.POST("/machine/recharge", mainAdminHandler.RechargeMachineHandler)
	mainAdminAuth.GET("/machine/balance/:machine_id", mainAdminHandler.GetMachineBalanceHandler)
	mainAdminAuth.GET("/machine/recharge-history/:machine_id", mainAdminHandler.GetRechargeMachineHistoryHandler)
	mainAdminAuth.GET("/college/machines", mainAdminHandler.GetMachinesByCollegeID)
	mainAdminAuth.GET("/rfid/recharge-history/:machine_id", mainAdminHandler.GetRFIDRechargeHistoryHandler)
	mainAdminAuth.POST("/machine/user/create", mainAdminHandler.CreateRechargeMachineUser)
	mainAdminAuth.GET("/revenue", fileHandler.CalculateTotalRevenueHandler)
	mainAdminAuth.GET("/printjobs", fileHandler.FetchPrintJobsHandler)
	mainAdminAuth.GET("/recent-printjobs", fileHandler.GetRecentPrintJobsHandler)
	mainAdminAuth.GET("/machine/users/:machine_id", mainAdminHandler.GetMachineUsersByMachineIdHandler)
	mainAdminAuth.DELETE("/machine/user/:machine_id/:user_id", mainAdminHandler.DeleteMachineUserHandler)
	mainAdminAuth.POST("/machine/card/:card_id", mainAdminHandler.CardDeativationHandler)
	mainAdminAuth.GET("/rfid/cards", mainAdminHandler.GetAllCardsHandler)
	mainAdminAuth.PUT("/machine/user/update/:user_id", mainAdminHandler.UpdateMachineUserPassword)

	/*machine user routes*/
	machineUser := e.Group("/machine-user")
	machineUser.POST("/login-user", mainAdminHandler.LoginRechargeMachineUser)

	machineUserAuth := machineUser.Group("")
	machineUserAuth.Use(machine_user_auth)
	machineUserAuth.POST("/rfid/initiate/:machine_id", mainAdminHandler.InitializeCard)
	machineUserAuth.POST("/recharge-rfid/:machine_id", mainAdminHandler.RechargeRFIDHandler) //need to inform frontend team
	machineUserAuth.GET("/fetchConnectedMachines/:machine_no", mainAdminHandler.FetchConnectedMachinesHandler)
	machineUserAuth.GET("/fetchMachineBalance/:machine_id", mainAdminHandler.GetMachineBalanceHandler)
	machineUserAuth.GET("/recharge-history/:machine_id", mainAdminHandler.GetRFIDRechargeHistoryHandler)
	machineUserAuth.GET("/rfid/details/:card_id", mainAdminHandler.GetRFIDCardDetails)
	machineUserAuth.GET("/rfid/balance/:card_id", mainAdminHandler.GetRFIDCardBalance)

	superAdmin := e.Group("/super-admin")
	superAdminAuth := superAdmin.Group("")
	superAdminAuth.Use(super_admin_auth)

	superAdmin.POST("/signup", superAdminHandler.CreateSuperAdmin)
	superAdmin.POST("/signin", superAdminHandler.LoginSuperAdminHandler)
	superAdminAuth.GET("/me", superAdminHandler.GetProfileHandler)
	superAdminAuth.POST("/college/create", superAdminHandler.CreateCollege)
	superAdminAuth.POST("/college/recharge/:college_id", superAdminHandler.RechargeCollege)
	superAdminAuth.GET("/college/recharge/history/:college_id", superAdminHandler.GetCollegeRechargeHistory)
	superAdminAuth.GET("/colleges", superAdminHandler.GetCollegesBySuperAdmin)
	superAdminAuth.GET("/college/:college_id", superAdminHandler.GetCollegeDetails)
	superAdminAuth.DELETE("/college/delete/:college_id", superAdminHandler.DeleteCollege)
	superAdminAuth.POST("/college/machine/create", superAdminHandler.CreateMachineHandler)
	superAdminAuth.GET("/college/machine/:college_id", superAdminHandler.GetMachinesByCollegeID)
	superAdminAuth.GET("/college/count", superAdminHandler.GetTotalCollegesCount)
	superAdminAuth.GET("/recharge/volume", superAdminHandler.GetTotalRechargeVolume)
	superAdminAuth.GET("/college/recharge/volume/:college_id", superAdminHandler.GetTotalRechargeVolumeByCollege)
	superAdminAuth.GET("/recharge/history", superAdminHandler.GetOverallCollgeRechargeHistory)
	superAdminAuth.GET("/machine/count", superAdminHandler.GetTotalMachineCount)
	superAdminAuth.GET("/machine/count/:college_id", superAdminHandler.GetTotalMachinesCountByCollege)
	superAdminAuth.GET("/rfid/cards/:college_id", superAdminHandler.GetRFIDCardsByCollegeId)
	superAdminAuth.POST("/rfid/card/:card_id", mainAdminHandler.CardDeativationHandler)
	superAdminAuth.DELETE("/college/machine/delete/:machine_id", superAdminHandler.DeleteRechargeMachine)
	superAdminAuth.PUT("/machine/update/:machine_id", superAdminHandler.UpdateMachine)
	superAdminAuth.PUT("/college/machine/update/:card_id", superAdminHandler.UpdateRFIDCard)
	superAdminAuth.DELETE("/rfid/delete/:card_id", superAdminHandler.DeleteRFIDCardHandler)
	superAdminAuth.PUT("/college/update/:college_id", superAdminHandler.UpdateCollege)

}
