package db

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	apperrors "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/errors"
	domain "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/files"
	model "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/payment_system"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type FilesRepo struct {
	client              *mongo.Client
	FilesCollection     *mongo.Collection
	PrintJobCollection  *mongo.Collection
	RFIDCardsCollection *mongo.Collection
}

func NewFilesRepo(db *mongo.Database, client *mongo.Client) *FilesRepo {
	return &FilesRepo{
		client:             client,
		FilesCollection:    db.Collection("files"),
		PrintJobCollection: db.Collection("PrintJobs"),
		RFIDCardsCollection: db.Collection("rfid_cards"),
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
		return false, apperrors.ErrInvalidID //change
	}

	filter := bson.M{"_id": objectID}
	var result domain.File

	err = r.FilesCollection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, apperrors.ErrFileNotFound
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
		return fmt.Errorf("%w: failed to insert print job: %v", apperrors.ErrDBFailure, err)
	}

	return nil
}

func (r *FilesRepo) GetFileKeyfromtheFileID(ctx context.Context, req string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	type S3Key struct {
		File_key string `bson:"file_key"`
	}

	var res S3Key

	objectID, err := primitive.ObjectIDFromHex(req)
	if err != nil {
		return "", apperrors.ErrInvalidID
	}

	filter := bson.M{
		"_id": objectID,
	}

	err = r.FilesCollection.FindOne(ctx, filter, options.FindOne().SetProjection(bson.M{"file_key": 1})).Decode(&res)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", apperrors.ErrFileNotFound
		}

		fmt.Printf("Error in database layer:error:%v", err)
		return "", fmt.Errorf("%w: %v", apperrors.ErrDBFailure, err)
	}

	fmt.Println("File_key:", res.File_key)
	return res.File_key, nil
}

func (r *FilesRepo) GetRecentUploadedFilesByFacultyID(ctx context.Context, FacultyID string) ([]domain.File, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(FacultyID)
	if err != nil {
		return nil, fmt.Errorf("Faculty not found in the database")
	}

	filter := bson.M{"faculty_id": objectID}
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

func (r *FilesRepo) FetchPrintJobs(ctx context.Context) ([]domain.PrintJob, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.M{"created_at": -1})

	cursor, err := r.PrintJobCollection.Find(ctx, bson.D{}, opts)

	if err != nil {
		return nil, fmt.Errorf("db.Find Error:%w", err)
	}
	defer cursor.Close(ctx)

	var printJobs []domain.PrintJob
	if err = cursor.All(ctx, &printJobs); err != nil {
		return nil, fmt.Errorf("cursore decode error:%w", err)
	}

	if len(printJobs) == 0 {
		return []domain.PrintJob{}, nil
	}
	return printJobs, nil
}

func (r *FilesRepo) DeleteFileRecord(ctx context.Context, fileID string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(fileID)
	if err != nil {
		return apperrors.ErrInvalidID
	}

	filter := bson.M{"_id": objectID}

	result, err := r.FilesCollection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return apperrors.ErrFileNotFound
	}

	return nil
}

func (r *FilesRepo) GetTotalFilesCount(ctx context.Context) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	count, err := r.FilesCollection.CountDocuments(ctx, bson.D{})

	if err != nil {
		return 0, fmt.Errorf("db.Find Error:%w", err)
	}

	if count == 0 {
		return 0, nil
	}
	return count, err
}

func (r *FilesRepo) DeleteFileRequest(ctx context.Context, fileID string, reason string) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if strings.TrimSpace(reason) == "" {
		return errors.New("delete reason is required")
	}

	objectID, err := primitive.ObjectIDFromHex(fileID)
	if err != nil {
		return errors.New("invalid file id")
	}

	filter := bson.M{
		"_id": objectID,
	}

	update := bson.M{
		"$set": bson.M{
			"delete_request": bson.M{
				"status": "pending",
				"reason": reason,
			},
			"updated_at": time.Now(),
		},
	}

	result, err := r.FilesCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("db update failed: %w", err)
	}

	if result.MatchedCount == 0 {
		return errors.New("file not found or delete already requested")
	}

	return nil
}

func (r *FilesRepo) GetPendingDeleteRequestFiles(ctx context.Context) ([]domain.File, error) {

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	filter := bson.M{
		"delete_request.status": "pending",
	}

	cursor, err := r.FilesCollection.Find(ctx, filter)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("db.find error:%v", err)
	}

	var files []domain.File

	if err := cursor.All(ctx, &files); err != nil {
		return nil, fmt.Errorf("cursor decode error:%v", err)
	}
	return files, nil
}

func (r *FilesRepo) DeleteFilePermanently(ctx context.Context, fileID string) error {

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(fileID)
	if err != nil {
		return apperrors.ErrInvalidInput
	}

	filter := bson.M{
		"_id":                   objectID,
		"delete_request.status": "pending",
	}

	result, err := r.FilesCollection.DeleteOne(ctx, filter)
	if err != nil {
		fmt.Println("DB Error:", err)
		return apperrors.ErrInternal
	}

	if result.DeletedCount == 0 {
		return apperrors.ErrNoPendingRequest
	}

	return nil
}

func (r *FilesRepo) RejectDeleteRequest(ctx context.Context, fileID string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(fileID)
	if err != nil {
		return apperrors.ErrInvalidInput
	}

	filter := bson.M{
		"_id":                   objectID,
		"delete_request.status": "pending",
	}

	update := bson.M{
		"$set": bson.M{
			"delete_request.status": "rejected",
			"delete_request.reason": "",
		},
	}

	result, err := r.FilesCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return apperrors.ErrNoPendingRequest
	}

	return nil
}

func (r *FilesRepo) PendingDeleteRequestCount(ctx context.Context) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{
		"delete_request.status": "pending",
	}

	cursor, err := r.FilesCollection.Find(ctx, filter)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return 0, nil
		}
		return 0, err
	}

	var files []domain.File
	if err = cursor.All(ctx, &files); err != nil {
		return 0, fmt.Errorf("cursor decode error: %w", err)
	}
	defer cursor.Close(ctx)

	var count int64 = 0
	for range files {
		count++
	}
	return count, nil
}

func (r *FilesRepo) RecentlyUploadedFiles(ctx context.Context) ([]domain.File, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.M{"uploaded_at": -1}).SetLimit(5)

	cursor, err := r.FilesCollection.Find(ctx, bson.M{}, opts)

	var files []domain.File

	if err = cursor.All(ctx, &files); err != nil {
		return nil, fmt.Errorf("cursor decode error: %w", err)
	}
	defer cursor.Close(ctx)

	return files, nil
}

func (r *FilesRepo) TotalFiles(ctx context.Context) ([]domain.File, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.M{"uploaded_at": -1})

	cursor, err := r.FilesCollection.Find(ctx, bson.M{}, opts)
	if err != nil {
		fmt.Print("error", err)
		return nil, err
	}

	var files []domain.File

	if err = cursor.All(ctx, &files); err != nil {
		return nil, fmt.Errorf("cursor decode error: %w", err)
	}
	defer cursor.Close(ctx)

	return files, nil
}

func (r *FilesRepo) GetPendingDeleteRequests(ctx context.Context, facultyId string) ([]domain.File, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(facultyId)
	if err != nil {
		return nil, fmt.Errorf("invalid faculty id")
	}

	filter := bson.M{
		"faculty_id":            objectID,
		"delete_request.status": "pending",
	}

	cursor, err := r.FilesCollection.Find(ctx, filter)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("db.Find error:%v", err)
	}
	defer cursor.Close(ctx)

	var files []domain.File
	if err = cursor.All(ctx, &files); err != nil {
		return nil, fmt.Errorf("cursor decode error:%v", err)
	}

	return files, nil
}

func (r *FilesRepo) CalculateTotalRevenue(ctx context.Context) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cursor, err := r.PrintJobCollection.Find(ctx, bson.M{})
	if err != nil {
		return 0, fmt.Errorf("db.Find error:%v", err)
	}
	defer cursor.Close(ctx)

	var printJobs []domain.PrintJob
	if err = cursor.All(ctx, &printJobs); err != nil {
		return 0, fmt.Errorf("cursor decode error:%v", err)
	}

	totalRevenue := 0
	for _, job := range printJobs {
		totalRevenue += job.Amount
	}

	return totalRevenue, nil
}

func (r *FilesRepo) GetRecentPrintJobs(ctx context.Context) ([]domain.PrintJob, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.M{"created_at": -1}).SetLimit(20)

	cursor, err := r.PrintJobCollection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("db.Find error:%v", err)
	}
	defer cursor.Close(ctx)

	var printJobs []domain.PrintJob
	if err = cursor.All(ctx, &printJobs); err != nil {
		return nil, fmt.Errorf("cursor decode error:%v", err)
	}

	return printJobs, nil
}

func (r *FilesRepo) GetUsnByCardId(ctx context.Context,cardId string)(string,error){
	ctx,cancel := context.WithTimeout(ctx,5*time.Second)
	defer cancel()

	filter := bson.M{
		"card_id":cardId,
	}

	var card  model.RFIDCard

	err := r.RFIDCardsCollection.FindOne(ctx,filter).Decode(&card)

	if err != nil {
		return "",fmt.Errorf("db.Find error:%v",err)
	}

	return card.USN,nil
}