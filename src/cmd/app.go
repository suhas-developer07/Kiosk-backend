package cmd

import (
	"github.com/labstack/echo/v4"
	repository_Files "github.com/suhas-developer07/Kiosk-backend/src/internals/repository/Files_repo"
	service_File "github.com/suhas-developer07/Kiosk-backend/src/internals/service"
	"github.com/suhas-developer07/Kiosk-backend/src/pkg/filestore"
	"go.mongodb.org/mongo-driver/mongo"
)

func Start(mongoClient *mongo.Client)*echo.Echo {
	e := echo.New()

	db := mongoClient.Database("kiosk_db")

	var storage filestore.FileStorage

	storage = filestore.NewLocalStorage("uploads")

	Files_repo := repository_Files.NewFilesRepo(db,mongoClient)
	File_service := service_File.NewFileService(Files_repo,storage)

	SetupRouter(
		e,
		File_service,
	)
	return e
}