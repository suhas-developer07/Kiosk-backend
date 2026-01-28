package admin

import "go.mongodb.org/mongo-driver/mongo"

type AdminRepo struct {
	client *mongo.Client
	Admin  *mongo.Collection
}

