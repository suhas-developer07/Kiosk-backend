package admin

import "go.mongodb.org/mongo-driver/mongo"

type AdminRepo struct {
	client *mongo.Client
	Admin  *mongo.Collection
}

func NewAdminRepo(db *mongo.Database,client *mongo.Client) *AdminRepo{
	return &AdminRepo{
		client: client,
		Admin: db.Collection("admin") ,
	}
}

