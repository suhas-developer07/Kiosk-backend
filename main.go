package main

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	echoadapter "github.com/awslabs/aws-lambda-go-api-proxy/echo"
	"github.com/suhas-developer07/Kiosk-backend/src/cmd"
)

var (
	echoLambda *echoadapter.EchoLambda
	once       sync.Once
)

func initApp() {
	log.Println("Initializing Lambda app")

	client, err := cmd.InitMongo(cmd.Config{
		URI:         os.Getenv("MONGO_URI"),
		MaxPoolSize: 50,
		MinPoolSize: 5,
		Timeout:     30 * time.Second,
	})
	if err != nil {
		panic(err)
	}

	e := cmd.Start(client)
	echoLambda = echoadapter.New(e)
}

func handler(
	ctx context.Context,
	req events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	once.Do(initApp)
	return echoLambda.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(handler)
}
