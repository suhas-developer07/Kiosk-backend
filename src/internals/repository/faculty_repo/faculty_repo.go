package facultyrepo

import (
	"context"
	"errors"
	"fmt"
	"time"

	domain "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/faculties"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FacultyRepo struct {
	client            *mongo.Client
	FacultyCollection *mongo.Collection
}

func NewFacultyRepo(db *mongo.Database, client *mongo.Client) *FacultyRepo {
	return &FacultyRepo{
		client:            client,
		FacultyCollection: db.Collection("faculties"),
	}
}

func (r *FacultyRepo) CreateAccount(ctx context.Context, req domain.Faculty) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"email": req.Email}

	var exists struct {
		ID primitive.ObjectID `bson:"_id"`
	}

	err := r.FacultyCollection.FindOne(ctx, filter).Decode(&exists)

	switch {
	case err == nil:
		return domain.ErrEmailAlreadyExists

	case errors.Is(err, mongo.ErrNoDocuments):
		// do nothing -> proced to insert
	default:
		return fmt.Errorf("database error during email check: %w", err)
	}

	_, err = r.FacultyCollection.InsertOne(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to insert faculty account: %w", err)
	}

	return nil
}
func (r *FacultyRepo) GetFacultyByEmail(ctx context.Context, email string) (*domain.Faculty, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"email": email}
	var faculty domain.Faculty

	err := r.FacultyCollection.FindOne(ctx, filter).Decode(&faculty)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("db error while checking email: %w", err)
	}

	return &faculty, nil
}

func (r *FacultyRepo) UpdateProfile(
	ctx context.Context,
	id primitive.ObjectID,
	profile domain.FacultyProfile,
) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"profile":              profile,
			"is_profile_completed": true,
			"updated_at":           time.Now(),
		},
	}

	result, err := r.FacultyCollection.UpdateByID(ctx, id, update)
	if err != nil {
		return fmt.Errorf("db: failed updating profile: %w", err)
	}

	if result.MatchedCount == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}
