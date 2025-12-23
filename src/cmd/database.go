package cmd

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mongoClient *mongo.Client
	mongoOnce   sync.Once
)

type Config struct {
	URI         string
	MaxPoolSize uint64
	MinPoolSize uint64
	Timeout     time.Duration
}
func InitMongo(cfg Config) (*mongo.Client, error) {
	var err error

	mongoOnce.Do(func() {
		clientOpts := options.Client().
			ApplyURI(cfg.URI).
			SetMaxPoolSize(cfg.MaxPoolSize).
			SetMinPoolSize(cfg.MinPoolSize)

		client, e := mongo.Connect(context.Background(), clientOpts)
		if e != nil {
			err = fmt.Errorf("mongo.connect error: %w", e)
			return
		}

		pingCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if e := client.Ping(pingCtx, nil); e != nil {
			err = fmt.Errorf("mongo.ping error: %w", e)
			return
		}

		mongoClient = client
		log.Println("Connected to MongoDB")
	})

	return mongoClient, err
}

func GetMongoClient() *mongo.Client {
	return mongoClient
}

func DisconnectMongo() error {
	if mongoClient == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := mongoClient.Disconnect(ctx); err != nil {
		return fmt.Errorf("mongo.Disconnect error :%w", err)
	}

	log.Println("MongoDB  Connection closed")

	return nil
}
