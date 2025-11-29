package cmd

import (
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
)

func Start(mongoClient *mongo.Client)*echo.Echo {
	e := echo.New()

	return e
}