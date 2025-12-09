package cmd

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	handler_Faculty "github.com/suhas-developer07/Kiosk-backend/src/internals/handlers/faculty_handler"
	handler_File "github.com/suhas-developer07/Kiosk-backend/src/internals/handlers/file_handler"
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

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	db := mongoClient.Database("kiosk_db")

	storage := filestore.NewLocalStorage("uploads")

	filesRepo := repository_Files.NewFilesRepo(db, mongoClient)

	fileService := service_File.NewFileService(filesRepo, storage, sugar)

	fileHandler := handler_File.NewFileHandler(fileService, sugar)

	facultyRepo := facultyrepo.NewFacultyRepo(db,mongoClient)

	facultyService :=service_Faculty.NewFacultyService(facultyRepo,sugar)

	facultyHandler := handler_Faculty.NewFacultyHandler(facultyService,sugar)

	SetupRouter(e, fileHandler,facultyHandler)

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status": "ok",
		})
	})

	return e
}
