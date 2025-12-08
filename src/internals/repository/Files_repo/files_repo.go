package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	domain "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/Files"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type FilesRepo struct {
	client             *mongo.Client
	FilesCollection    *mongo.Collection
	PrintJobCollection *mongo.Collection
}

func NewFilesRepo(db *mongo.Database, client *mongo.Client) *FilesRepo {
	return &FilesRepo{
		client:             client,
		FilesCollection:    db.Collection("files"),
		PrintJobCollection: db.Collection("PrintJobs"),
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

func (r *FilesRepo) GetFileByGradeAndSubject(
	ctx context.Context,
	grade string,
	subject string,
) ([]domain.File, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{
		"grade":   grade,
		"subject": subject,
	}

	opts := options.Find().SetSort(bson.M{"uploaded_at": -1})

	cursor, err := r.FilesCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("db.Find error: %w", err)
	}
	defer cursor.Close(ctx)

	var files []domain.File
	if err = cursor.All(ctx, &files); err != nil {
		return nil, fmt.Errorf("cursor decode error: %w", err)
	}

	if len(files) == 0 {
		return []domain.File{}, nil
	}

	return files, nil
}
func (r *FilesRepo) GetFileByID(ctx context.Context, id string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return false, domain.ErrInvalidID
	}

	filter := bson.M{"_id": objectID}
	var result domain.File

	err = r.FilesCollection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, domain.ErrFileNotFound
		}
		return false, fmt.Errorf("%w: %v", domain.ErrDBFailure, err)
	}

	return true, nil
}

func (r *FilesRepo) CreatePrintJob(ctx context.Context, req domain.PrintJob) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.PrintJobCollection.InsertOne(ctx, req)
	if err != nil {
		return fmt.Errorf("%w: failed to insert print job: %v", domain.ErrDBFailure, err)
	}

	return nil
}
