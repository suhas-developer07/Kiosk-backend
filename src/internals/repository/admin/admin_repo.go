package admin

import (
	"context"
	"errors"
	"fmt"
	"time"

	model "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/admin"
	apperrors "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AdminRepo struct {
	client *mongo.Client
	Admin  *mongo.Collection
}

func NewAdminRepo(db *mongo.Database, client *mongo.Client) *AdminRepo {
	return &AdminRepo{
		client: client,
		Admin:  db.Collection("admin"),
	}
}

func (r *AdminRepo) GetAdminByEmail(ctx context.Context, email string) (*model.Admin, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"email": email}
	var Admin model.Admin

	err := r.Admin.FindOne(ctx, filter).Decode(&Admin)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, apperrors.ErrAdminNotFound
		}
		return nil, fmt.Errorf("db error while checking email: %w", err)
	}

	return &Admin, nil
}

func (r *AdminRepo) CreateAccount(ctx context.Context, req model.Admin) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"email": req.Email}

	var exists struct {
		ID primitive.ObjectID `bson:"_id"`
	}

	err := r.Admin.FindOne(ctx, filter).Decode(&exists)

	switch {
	case err == nil:
		return apperrors.ErrEmailAlreadyExists

	case errors.Is(err, mongo.ErrNoDocuments):
		// do nothing -> proced to insert
	default:
		return fmt.Errorf("database error during email check: %w", err)
	}

	_, err = r.Admin.InsertOne(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to insert faculty account: %w", err)
	}

	return nil
}

func (r *AdminRepo) GetAdminByID(ctx context.Context, adminID primitive.ObjectID) (model.Admin, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{
		"_id": adminID,
	}

	var admin model.Admin

	err := r.Admin.FindOne(ctx, filter).Decode(&admin)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return model.Admin{}, apperrors.ErrAdminNotFound
		}
		return model.Admin{}, fmt.Errorf("db error: %v", err)
	}

	return admin, nil
}

func (r *AdminRepo) DeleteFaculty(ctx context.Context, facultyId primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{
		"_id": facultyId,
	}

	res,err := r.Admin.DeleteOne(ctx, filter)

	if err != nil {
		return fmt.Errorf("db error :%v", err)
	}

	if res.DeletedCount == 0 {
		return fmt.Errorf("db deletion error:%v",err)
	}
	return nil
}
