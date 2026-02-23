package cmd

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	handler_Admin "github.com/suhas-developer07/Kiosk-backend/src/internals/handlers/admin"
	handler_Faculty "github.com/suhas-developer07/Kiosk-backend/src/internals/handlers/faculty_handler"
	handler_File "github.com/suhas-developer07/Kiosk-backend/src/internals/handlers/file_handler"
	handler_orchestrator "github.com/suhas-developer07/Kiosk-backend/src/internals/handlers/orchestrator"
	payment_handler "github.com/suhas-developer07/Kiosk-backend/src/internals/handlers/payment_system"
	"github.com/suhas-developer07/Kiosk-backend/src/internals/middleware"
	adminRepo "github.com/suhas-developer07/Kiosk-backend/src/internals/repository/admin"
	facultyrepo "github.com/suhas-developer07/Kiosk-backend/src/internals/repository/faculty_repo"
	repository_Files "github.com/suhas-developer07/Kiosk-backend/src/internals/repository/files_repo"
	payment_repository "github.com/suhas-developer07/Kiosk-backend/src/internals/repository/payment_system"
	service_Admin "github.com/suhas-developer07/Kiosk-backend/src/internals/service/admin"
	service_Faculty "github.com/suhas-developer07/Kiosk-backend/src/internals/service/faculty_service"
	service_File "github.com/suhas-developer07/Kiosk-backend/src/internals/service/file_service"
	service_orchestrator "github.com/suhas-developer07/Kiosk-backend/src/internals/service/orchestrator"
	payment_service "github.com/suhas-developer07/Kiosk-backend/src/internals/service/payment_system"

	"github.com/suhas-developer07/Kiosk-backend/src/pkg/filestore"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

func Start(mongoClient *mongo.Client) *echo.Echo {
	e := echo.New()

	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			log.Println("Echo received path:", c.Request().URL.Path)
			return next(c)
		}
	})

	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())

	e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			echo.GET,
			echo.POST,
			echo.PUT,
			echo.DELETE,
			echo.PATCH,
			echo.OPTIONS,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
		},
	}))

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	s3Client := s3.NewFromConfig(cfg)

	storage := filestore.NewS3Storage(
		s3Client,
		os.Getenv("FILES_BUCKET"),
	)

	db := mongoClient.Database("kiosk_db")

	Faculty_auth := middleware.FacultyAuthMiddleware(sugar)

	Admin_auth := middleware.AdminAuthMiddleware(sugar)

	Machine_User_Auth := middleware.MachineUserAuth(sugar)

	Super_Admin_Auth := middleware.SuperAdminAuthMiddleware(sugar)

	College_Auth := middleware.CollegeAuth(sugar)

	filesRepo := repository_Files.NewFilesRepo(db, mongoClient)

	fileService := service_File.NewFileService(filesRepo, storage, sugar)

	fileHandler := handler_File.NewFileHandler(fileService, sugar)

	facultyRepo := facultyrepo.NewFacultyRepo(db, mongoClient)

	facultyService := service_Faculty.NewFacultyService(facultyRepo, sugar)

	facultyHandler := handler_Faculty.NewFacultyHandler(facultyService, sugar)

	orchestratorService := service_orchestrator.NewOrchestrateService(filesRepo,facultyRepo,storage,sugar)

	orchestratorHandler := handler_orchestrator.NewOrchestrateHandler(orchestratorService,sugar)

	adminRepo := adminRepo.NewAdminRepo(db,mongoClient)

	adminService := service_Admin.NewAdminService(adminRepo,facultyRepo,filesRepo,storage,sugar)

	adminHandler := handler_Admin.NewAdminHandler(adminService,sugar)

	MainAdminRepo := payment_repository.NewMainAdminRepo(db,mongoClient)

	MainAdminService := payment_service.NewMainAdminService(MainAdminRepo, sugar)

	MainAdminHandler := payment_handler.NewMainAdminHandler(MainAdminService, sugar)

	SuperAdminRepo := payment_repository.NewSuperAdminRepo(db,mongoClient)

	SuperAdminService := payment_service.NewSuperAdminService(SuperAdminRepo, sugar)

	SuperAdminHandler := payment_handler.NewSuperAdminHandler(SuperAdminService, sugar)

	SetupRouter(e, fileHandler, facultyHandler,orchestratorHandler,adminHandler,MainAdminHandler,SuperAdminHandler,Faculty_auth,Admin_auth,Machine_User_Auth,College_Auth,Super_Admin_Auth)
	

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status": "ok",
		})
	})

	return e
}
