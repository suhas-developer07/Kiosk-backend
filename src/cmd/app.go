package cmd

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	handler_Faculty "github.com/suhas-developer07/Kiosk-backend/src/internals/handlers/faculty_handler"
	handler_File "github.com/suhas-developer07/Kiosk-backend/src/internals/handlers/file_handler"
	handler_orchestrator "github.com/suhas-developer07/Kiosk-backend/src/internals/handlers/orchestrator"
	service_orchestrator "github.com/suhas-developer07/Kiosk-backend/src/internals/service/orchestrator"
	"github.com/suhas-developer07/Kiosk-backend/src/internals/middleware"
	facultyrepo "github.com/suhas-developer07/Kiosk-backend/src/internals/repository/faculty_repo"
	repository_Files "github.com/suhas-developer07/Kiosk-backend/src/internals/repository/files_repo"
	service_Faculty "github.com/suhas-developer07/Kiosk-backend/src/internals/service/faculty_service"
	service_File "github.com/suhas-developer07/Kiosk-backend/src/internals/service/file_service"
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

	auth := middleware.AuthMiddleware(sugar)

	filesRepo := repository_Files.NewFilesRepo(db, mongoClient)

	fileService := service_File.NewFileService(filesRepo, storage, sugar)

	fileHandler := handler_File.NewFileHandler(fileService, sugar)

	facultyRepo := facultyrepo.NewFacultyRepo(db, mongoClient)

	facultyService := service_Faculty.NewFacultyService(facultyRepo, sugar)

	facultyHandler := handler_Faculty.NewFacultyHandler(facultyService, sugar)

	orchestratorService := service_orchestrator.NewUploadService(filesRepo,facultyRepo,storage,sugar)
	orchestratorHandler := handler_orchestrator.NewUploadHandler(orchestratorService,sugar)

	SetupRouter(e, fileHandler, facultyHandler,orchestratorHandler,auth)

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status": "ok",
		})
	})

	return e
}
