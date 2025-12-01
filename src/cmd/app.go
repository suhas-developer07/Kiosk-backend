package cmd

import (
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
	repository_Files "github.com/suhas-developer07/Kiosk-backend/src/internals/repository/Files_repo"
	service_File "github.com/suhas-developer07/Kiosk-backend/src/internals/service"
)

func Start(mongoClient *mongo.Client)*echo.Echo {
	e := echo.New()

	db := mongoClient.Database("kiosk_db")

	Files_repo := repository_Files.NewFilesRepo(db)
	File_service := service_File.NewFileService(Files_repo)

	SetupRouter(
		e,
		File_service,
	)
	return e
}