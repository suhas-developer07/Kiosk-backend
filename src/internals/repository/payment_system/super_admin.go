package paymentsystem

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	apperrors "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/errors"
	model "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/payment_system"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	collectionSuperAdmins            = "super_admins"
	collectionColleges               = "colleges"
	collectionCollegeRechargeHistory = "college_recharge_history"
)

type SuperAdminRepository struct {
	client                           *mongo.Client
	superAdminCollection             *mongo.Collection
	collegeCollection                *mongo.Collection
	collegeRechargeHistoryCollection *mongo.Collection
	machineCollection                *mongo.Collection
}

func NewSuperAdminRepo(db *mongo.Database, client *mongo.Client) *SuperAdminRepository {
	return &SuperAdminRepository{
		client:                           client,
		superAdminCollection:             db.Collection(collectionSuperAdmins),
		collegeCollection:                db.Collection(collectionColleges),
		collegeRechargeHistoryCollection: db.Collection(collectionCollegeRechargeHistory),
		machineCollection:                db.Collection(collectionMachines),
	}
}

/* Super Admin Repository Methods */

func (r *SuperAdminRepository) CreateSuperAdmin(ctx context.Context, superAdmin model.SuperAdmin) error {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	_, err := r.superAdminCollection.InsertOne(ctx, superAdmin)
	if err != nil {
		return fmt.Errorf("failed to insert super admin: %w", err)
	}

	return nil
}

func (r *SuperAdminRepository) CheckSuperAdminEmailExists(ctx context.Context, email string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	count, err := r.superAdminCollection.CountDocuments(ctx, bson.M{"super_admin_email": email})
	if err != nil {
		return false, fmt.Errorf("failed to check super admin email existence: %w", err)
	}

	return count > 0, nil
}

func (r *SuperAdminRepository) GetSuperAdminByEmail(ctx context.Context, email string) (*model.SuperAdmin, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	var superAdmin model.SuperAdmin

	err := r.superAdminCollection.FindOne(ctx, bson.M{"super_admin_email": email}).Decode(&superAdmin)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, apperrors.ErrSuperAdminNotFound
		}
		return nil, fmt.Errorf("failed to retrieve super admin by email: %w", err)
	}

	return &superAdmin, nil
}

func (r *SuperAdminRepository) UpdateSuperAdminBalance(ctx context.Context, superAdminID, newBalance string) error {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	if _, err := strconv.ParseFloat(newBalance, 64); err != nil {
		return errors.New("invalid balance value provided")
	}

	res, err := r.superAdminCollection.UpdateOne(ctx,
		bson.M{"super_admin_id": superAdminID},
		bson.M{"$set": bson.M{"balance": newBalance}},
	)
	if err != nil {
		return fmt.Errorf("failed to update super admin balance: %w", err)
	}
	if res.MatchedCount == 0 {
		return apperrors.ErrSuperAdminNotFound
	}

	return nil
}

/*  College Repository Methods  */

func (r *SuperAdminRepository) CreateCollege(ctx context.Context, college model.SuperAdminCollege) error {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	_, err := r.collegeCollection.InsertOne(ctx, college)
	if err != nil {
		return fmt.Errorf("failed to insert college: %w", err)
	}

	return nil
}

func (r *SuperAdminRepository) CheckCollegeEmailExists(ctx context.Context, email string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	count, err := r.collegeCollection.CountDocuments(ctx, bson.M{"college_email": email})
	if err != nil {
		return false, fmt.Errorf("failed to check college email existence: %w", err)
	}

	return count > 0, nil
}

func (r *SuperAdminRepository) GetCollegeBalance(ctx context.Context, collegeID string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	var result struct {
		Balance string `bson:"balance"`
	}

	err := r.collegeCollection.FindOne(ctx, bson.M{"college_id": collegeID}).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", apperrors.ErrCollegeNotFound
		}
		return "", fmt.Errorf("failed to fetch college balance: %w", err)
	}

	if _, err := strconv.ParseFloat(result.Balance, 64); err != nil {
		return "", errors.New("stored college balance is in an invalid format")
	}

	return result.Balance, nil
}

func (r *SuperAdminRepository) UpdateCollegeBalance(ctx context.Context, collegeID, newBalance string) error {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	if _, err := strconv.ParseFloat(newBalance, 64); err != nil {
		return errors.New("invalid balance value provided")
	}

	res, err := r.collegeCollection.UpdateOne(ctx,
		bson.M{"college_id": collegeID},
		bson.M{"$set": bson.M{"balance": newBalance}},
	)
	if err != nil {
		return fmt.Errorf("failed to update college balance: %w", err)
	}
	if res.MatchedCount == 0 {
		return apperrors.ErrCollegeNotFound
	}

	return nil
}

func (r *SuperAdminRepository) GetCollegesBySuperAdminID(ctx context.Context, adminID string) ([]model.SuperAdminCollege, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	cursor, err := r.collegeCollection.Find(ctx, bson.M{"super_admin_id": adminID})
	if err != nil {
		return nil, fmt.Errorf("failed to query colleges by super admin: %w", err)
	}
	defer cursor.Close(ctx)

	var colleges []model.SuperAdminCollege
	for cursor.Next(ctx) {
		var college model.SuperAdminCollege
		if err := cursor.Decode(&college); err != nil {
			return nil, fmt.Errorf("failed to decode college record: %w", err)
		}
		colleges = append(colleges, college)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error while reading colleges: %w", err)
	}

	return colleges, nil
}

func (r *SuperAdminRepository) GetCollegeByID(ctx context.Context, collegeID string) (*model.SuperAdminCollege, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	var college model.SuperAdminCollege

	err := r.collegeCollection.FindOne(ctx, bson.M{"college_id": collegeID}).Decode(&college)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, apperrors.ErrCollegeNotFound
		}
		return nil, fmt.Errorf("failed to retrieve college by ID: %w", err)
	}

	return &college, nil
}

func (r *SuperAdminRepository) DeleteCollege(ctx context.Context, collegeID string) error {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	res, err := r.collegeCollection.DeleteOne(ctx, bson.M{"college_id": collegeID})
	if err != nil {
		return fmt.Errorf("failed to delete college: %w", err)
	}
	if res.DeletedCount == 0 {
		return apperrors.ErrCollegeNotFound
	}

	return nil
}

func (r *SuperAdminRepository) InsertCollegeRechargeHistory(ctx context.Context, history model.CollegeRechargeHistory) error {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	_, err := r.collegeRechargeHistoryCollection.InsertOne(ctx, history)
	if err != nil {
		return fmt.Errorf("failed to insert college recharge history: %w", err)
	}

	return nil
}

func (r *SuperAdminRepository) GetRechargeHistoryByCollegeID(ctx context.Context, collegeID string) ([]model.CollegeRechargeHistory, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	cursor, err := r.collegeRechargeHistoryCollection.Find(ctx, bson.M{"college_id": collegeID})
	if err != nil {
		return nil, fmt.Errorf("failed to query college recharge history: %w", err)
	}
	defer cursor.Close(ctx)

	var recharges []model.CollegeRechargeHistory
	for cursor.Next(ctx) {
		var recharge model.CollegeRechargeHistory
		if err := cursor.Decode(&recharge); err != nil {
			return nil, fmt.Errorf("failed to decode recharge history record: %w", err)
		}
		recharges = append(recharges, recharge)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error while reading recharge history: %w", err)
	}

	return recharges, nil
}

/* Machine Repository Methods */

func (r *SuperAdminRepository) CreateMachine(ctx context.Context, machine model.Machine) error {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	filter := bson.M{
		"machine_no": machine.MachineNo,
		"college_id": machine.CollegeId,
	}

	var existing model.Machine
	err := r.machineCollection.FindOne(ctx, filter).Decode(&existing)
	if err == nil {
		return apperrors.ErrMachineAlreadyExists
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return fmt.Errorf("failed to check existing machine: %w", err)
	}

	_, err = r.machineCollection.InsertOne(ctx, machine)
	if err != nil {
		return fmt.Errorf("failed to insert machine: %w", err)
	}

	return nil
}

func (r *SuperAdminRepository) GetMachinesByCollegeID(ctx context.Context, collegeID string) ([]model.Machine, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	cursor, err := r.machineCollection.Find(ctx, bson.M{"college_id": collegeID})
	if err != nil {
		return nil, fmt.Errorf("failed to query machines by college: %w", err)
	}
	defer cursor.Close(ctx)

	var machines []model.Machine
	for cursor.Next(ctx) {
		var machine model.Machine
		if err := cursor.Decode(&machine); err != nil {
			return nil, fmt.Errorf("failed to decode machine record: %w", err)
		}
		machines = append(machines, machine)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error while reading machines: %w", err)
	}

	return machines, nil
}