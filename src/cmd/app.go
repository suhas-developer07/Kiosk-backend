package cmd

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	handler_File "github.com/suhas-developer07/Kiosk-backend/src/internals/handlers/File_Handler"
	repository_Files "github.com/suhas-developer07/Kiosk-backend/src/internals/repository/Files_repo"
	service_File "github.com/suhas-developer07/Kiosk-backend/src/internals/service"
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

	SetupRouter(e, fileHandler)

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status": "ok",
		})
	})

	return e
}
