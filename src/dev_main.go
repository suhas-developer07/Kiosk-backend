package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/suhas-developer07/Kiosk-backend/src/cmd"
)

func init() {

	godotenv.Load()

	serverMode := os.Getenv("SERVER_MODE")

	switch serverMode {
	case "dev":
		log.Println("Running server in development mode...")
	case "prod":
		log.Println("Running server in production mode...")
	default:
		log.Fatal("please set SERVER_MODE={dev|prod}")
	}
}

func main() {

	cfg := cmd.Config{
		URI:         os.Getenv("MONGO_URI"),
		MaxPoolSize: 50,
		MinPoolSize: 5,
		Timeout:     10 * time.Second,
	}

	client, err := cmd.InitMongo(cfg)
	if err != nil {
		log.Fatalf("Cannot init Mongo: %v", err)
	}

	server := cmd.Start(client)

	go func() {
		if err := server.Start(os.Getenv("SERVER_LISTEN_ADDR")); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Println("Shutting down server...")

	if err := cmd.DisconnectMongo(); err != nil {
		log.Printf("Error disconnecting Mongo: %v", err)
	}

	log.Println("Server stopped gracefully.")
}
