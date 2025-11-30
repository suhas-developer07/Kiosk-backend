package filesrepo

import (
	"context"

	models "github.com/suhas-developer07/Kiosk-backend/src/internals/Models"
	"go.mongodb.org/mongo-driver/mongo"
)

type FilesRepo struct {
	FilesCollection   *mongo.Collection
	PrintJobsCollection *mongo.Collection
}

func NewFilesRepo(db *mongo.Database) *FilesRepo {
	return &FilesRepo{
		FilesCollection:     db.Collection("files"),
		PrintJobsCollection: db.Collection("print_jobs"),
	}
}

func (r *FilesRepo) InsertFile(ctx context.Context,file models.File) (*mongo.InsertOneResult, error) {
	id,err := r.FilesCollection.InsertOne(ctx,file);

	if err != nil {
		return nil, err
	}

	return id, nil
}

func (r *FilesRepo) InsertPrintJob(ctx context.Context,job models.PrintJob) (*mongo.InsertOneResult, error) {
	id,err := r.PrintJobsCollection.InsertOne(ctx,job);

	if err != nil {
		return nil, err
	}

	return id, nil
}

