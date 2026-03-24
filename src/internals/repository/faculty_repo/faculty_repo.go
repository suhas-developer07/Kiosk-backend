package facultyrepo

import (
	"context"
	"errors"
	"fmt"
	"time"

	apperrors "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/errors"

	domain "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/faculties"
	"github.com/suhas-developer07/Kiosk-backend/src/internals/domain/subjects"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
		return apperrors.ErrEmailAlreadyExists

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
			return nil, domain.ErrFacultyNotFound
		}
		return nil, fmt.Errorf("db error while checking email: %w", err)
	}

	return &faculty, nil
}

// func (r *FacultyRepo) UpdateProfile(
// 	ctx context.Context,
// 	id primitive.ObjectID,
// 	profile domain.FacultyProfile,
// ) error {

// 	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
// 	defer cancel()

// 	update := bson.M{
// 		"$set": bson.M{
// 			"profile":              profile,
// 			"is_profile_completed": true,
// 			"updated_at":           time.Now(),
// 		},
// 	}

// 	result, err := r.FacultyCollection.UpdateByID(ctx, id, update)
// 	if err != nil {
// 		return fmt.Errorf("db: failed updating profile: %w", err)
// 	}

// 	if result.MatchedCount == 0 {
// 		return domain.ErrFacultyNotFound
// 	}

// 	return nil
// }

func (r *FacultyRepo) GetFacultyProfileByID(
	ctx context.Context,
	id primitive.ObjectID,
) (domain.Faculty, error) {

	var faculty domain.Faculty

	err := r.FacultyCollection.FindOne(
		ctx,
		bson.M{"_id": id},
	).Decode(&faculty)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Faculty{}, domain.ErrFacultyNotFound
		}
		return domain.Faculty{}, err
	}

	return faculty, nil
}

func (r *FacultyRepo) HasSubject(
	ctx context.Context,
	facultyID primitive.ObjectID,
	subject subjects.Subject,
) (bool, error) {

	filter := bson.M{
		"_id":      facultyID,
		"subjects": subject,
	}

	err := r.FacultyCollection.FindOne(ctx, filter).Err()

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (r *FacultyRepo) GetFacultyByID(ctx context.Context, facultyId string) (domain.Faculty, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(facultyId)
	if err != nil {
		return domain.Faculty{}, domain.ErrInvalidID
	}

	var faculty domain.Faculty

	filter := bson.M{
		"_id": objID,
	}

	err = r.FacultyCollection.FindOne(ctx, filter).Decode(&faculty)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Faculty{}, fmt.Errorf("db:There is faculty found in this id")
		}
		return domain.Faculty{}, err
	}

	return faculty, nil
}

func (r *FacultyRepo) AddFaculty(ctx context.Context, req domain.Faculty) error {
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
		return fmt.Errorf("failed to add a faculty: %w", err)
	}

	return nil
}

func (r *FacultyRepo) GetFaculties(ctx context.Context) ([]domain.Faculty, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var Faculty []domain.Faculty

	findOptions := options.Find()

	cursor, err := r.FacultyCollection.Find(ctx, bson.M{}, findOptions)

	if err != nil {
		return nil, fmt.Errorf("db.Find Error:%w", err)
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &Faculty); err != nil {
		return nil, fmt.Errorf("cursor decode error:%v", err)
	}

	if len(Faculty) == 0 {
		return []domain.Faculty{}, nil
	}
	return Faculty, nil
}

func (r *FacultyRepo) GetFacultiesByStream(ctx context.Context, stream string) ([]domain.Faculty, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var Faculty []domain.Faculty

	fmt.Println("stream", stream)

	filter := bson.M{
		"stream": stream,
	}

	cursor, err := r.FacultyCollection.Find(ctx, filter)

	if err != nil {
		return nil, fmt.Errorf("db.Find Error:%w", err)
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &Faculty); err != nil {
		return nil, fmt.Errorf("cursor decode error:%v", err)
	}

	if len(Faculty) == 0 {
		return []domain.Faculty{}, nil
	}
	return Faculty, nil
}

func (r *FacultyRepo) UpdateFaculty(
	ctx context.Context,
	facultyId primitive.ObjectID,
	updateData bson.M,
) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{
		"_id": facultyId,
	}

	update := bson.M{
		"$set": updateData,
	}

	result, err := r.FacultyCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update faculty: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("faculty not found")
	}

	return nil
}
