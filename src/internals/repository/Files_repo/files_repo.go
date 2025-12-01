package filesrepo

import (
	"context"

	domain "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/Files"
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

func (r *FilesRepo) InsertFile(ctx context.Context,file domain.File) ( error) {
	_,err := r.FilesCollection.InsertOne(ctx,file);

	if err != nil {
		return  err
	}

	return  nil
}

func (r *FilesRepo) InsertPrintJob(ctx context.Context,job domain.PrintJob) (*mongo.InsertOneResult, error) {
	id,err := r.PrintJobsCollection.InsertOne(ctx,job);

	if err != nil {
		return nil, err
	}

	return id, nil
}

