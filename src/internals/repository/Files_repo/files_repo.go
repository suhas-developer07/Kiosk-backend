package db

import (
	"context"

	domain "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/Files"
	"go.mongodb.org/mongo-driver/mongo"
)

type FilesRepo struct {
	client          *mongo.Client
	FilesCollection *mongo.Collection
}

func NewFilesRepo(db *mongo.Database, client *mongo.Client) *FilesRepo {
	return &FilesRepo{
		client:          client,
		FilesCollection: db.Collection("files"),
	}
}

func (r *FilesRepo) SaveFileRecord(ctx context.Context, file domain.File) error {
	_, err := r.FilesCollection.InsertOne(ctx, file)
	return err
}

func (r *FilesRepo) WithTransaction(
	ctx context.Context,
	fn func(sc mongo.SessionContext) error,
) error {

	session, err := r.client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	return mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {

		if err := session.StartTransaction(); err != nil {
			return err
		}

		if err := fn(sc); err != nil {
			_ = session.AbortTransaction(sc)
			return err
		}

		return session.CommitTransaction(sc)
	})
}
