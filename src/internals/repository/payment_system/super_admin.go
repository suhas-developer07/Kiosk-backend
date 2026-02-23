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
	SUPER_ADMIN_COLLECTION              = "super_admins"
	COLLEGE_COLLECTION                  = "colleges"
	RECHARGE_COLLEGE_HISTORY_COLLECTION = "college_recharge_history"
	MACHINE_COLLECTION                  = "recharge_machines"
)

type SuperAdminRepository struct {
	client                           *mongo.Client
	SuperAdminCollection             *mongo.Collection
	CollegeCollection                *mongo.Collection
	CollegeRechargeHistoryCollection *mongo.Collection
	MachineCollection                *mongo.Collection
}

func NewSuperAdminRepo(db *mongo.Database, client *mongo.Client) *SuperAdminRepository {
	return &SuperAdminRepository{
		client:                           client,
		SuperAdminCollection:             db.Collection(SUPER_ADMIN_COLLECTION),
		CollegeCollection:                db.Collection(COLLEGE_COLLECTION),
		CollegeRechargeHistoryCollection: db.Collection(RECHARGE_COLLEGE_HISTORY_COLLECTION),
		MachineCollection:                db.Collection(MACHINE_COLLECTION),
	}
}

/*
Not implemented yet
*/
func (repo *SuperAdminRepository) CreateSuperAdmin(ctx context.Context, superadmin model.SuperAdmin) error {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()
	_, err := repo.SuperAdminCollection.InsertOne(ctx, superadmin)
	if err != nil {
		return errors.New("failed to create Super Admin account. Please try again later")
	}
	return nil
}

func (repo *SuperAdminRepository) CheckSuperAdminEmailExists(ctx context.Context, email string) (bool, error) {
	filter := bson.M{"super_admin_email": email}
	count, err := repo.SuperAdminCollection.CountDocuments(ctx, filter)
	if err != nil {
		return false, errors.New("unable to verify email. Please try again later")
	}
	return count > 0, nil
}

func (repo *SuperAdminRepository) GetSuperAdminByEmail(ctx context.Context, email string) (*model.SuperAdmin, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()
	filter := bson.M{"super_admin_email": email}
	var superAdmin model.SuperAdmin

	err := repo.SuperAdminCollection.FindOne(ctx, filter).Decode(&superAdmin)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("super admin not found")
		}
		return nil, errors.New("unable to retrieve super admin details")
	}

	return &superAdmin, nil
}

func (repo *SuperAdminRepository) GetSuperAdminBalance(ctx context.Context, superAdminID string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()
	filter := bson.M{"super_admin_id": superAdminID}
	var superAdmin model.SuperAdmin

	err := repo.SuperAdminCollection.FindOne(ctx, filter).Decode(&superAdmin)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "0", errors.New("superadmin not found")
		}
		return "0", errors.New("unable to fetch superadmin balance")
	}

	return superAdmin.Balance, nil
}

func (repo *SuperAdminRepository) UpdateSuperAdminBalance(ctx context.Context, superAdminID, newBalance string) error {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()
	filter := bson.M{"super_admin_id": superAdminID}
	update := bson.M{"$set": bson.M{"balance": newBalance}}
	res, err := repo.SuperAdminCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return errors.New("failed to update superadmin balance")
	}
	if res.MatchedCount == 0 {
		return errors.New("superadmin not found")
	}
	return nil
}

func (r *SuperAdminRepository) CreateCollege(ctx context.Context, college model.SuperAdminCollege) error {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	_, err := r.CollegeCollection.InsertOne(ctx, college)
	if err != nil {
		return errors.New("failed to create college")
	}
	return nil
}

func (repo *SuperAdminRepository) GetCollegeForLogin(ctx context.Context, email string) (collegeID, collegename, hashedPassword, superadminID string, err error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	filter := bson.M{"college_email": email}
	var college model.SuperAdminCollege
	err = repo.CollegeCollection.FindOne(ctx, filter).Decode(&college)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", "", "", "", errors.New("college not found")
		}
		return "", "", "", "", errors.New("unable to retrieve college login details")
	}
	return college.CollegeID, college.CollegeName, college.CollegePassword, college.SuperAdminId, nil
}

func (r *SuperAdminRepository) CheckCollegeEmailExists(ctx context.Context, email string) (bool, error) {
	filter := bson.M{"college_email": email}
	count, err := r.CollegeCollection.CountDocuments(ctx, filter)
	if err != nil {
		return false, errors.New("failed to check college email")
	}
	return count > 0, nil
}

func (repo *SuperAdminRepository) GetCollegeBalance(ctx context.Context, collegeID string) (string, error) {
	var college struct {
		Balance string `bson:"balance"`
	}
	err := repo.CollegeCollection.FindOne(ctx, bson.M{"college_id": collegeID}).Decode(&college)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", errors.New("college not found")
		}
		return "", errors.New("unable to fetch college balance at the moment")
	}
	if _, err := strconv.ParseFloat(college.Balance, 64); err != nil {
		return "", errors.New("stored college balance is in an invalid format")
	}
	return college.Balance, nil
}

func (repo *SuperAdminRepository) UpdateCollegeBalance(ctx context.Context, collegeID string, newBalance string) error {
	if _, err := strconv.ParseFloat(newBalance, 64); err != nil {
		return errors.New("invalid balance value provided")
	}
	res, err := repo.CollegeCollection.UpdateOne(ctx,
		bson.M{"college_id": collegeID},
		bson.M{"$set": bson.M{"balance": newBalance}},
	)
	if err != nil {
		return errors.New("failed to update college balance due to a database error")
	}
	if res.MatchedCount == 0 {
		return errors.New("college not found")
	}
	return nil
} 

func (repo *SuperAdminRepository) GetCollegesBySuperAdminID(ctx context.Context, adminID string) ([]model.SuperAdminCollege, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	filter := bson.M{"super_admin_id": adminID}
	cursor, err := repo.CollegeCollection.Find(ctx, filter)
	if err != nil {
		return nil, errors.New("failed to fetch colleges")
	}
	defer cursor.Close(ctx)

	var colleges []model.SuperAdminCollege
	for cursor.Next(ctx) {
		var college model.SuperAdminCollege
		if err := cursor.Decode(&college); err != nil {
			return nil, errors.New("error reading college data")
		}
		colleges = append(colleges, college)
	}

	if len(colleges) == 0 {
		return nil, errors.New("no colleges found for the given super admin")
	}

	return colleges, nil
}

func (repo *SuperAdminRepository) GetCollegeByID(ctx context.Context, collegeID string) (*model.SuperAdminCollege, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	filter := bson.M{"college_id": collegeID}
	var college model.SuperAdminCollege
	err := repo.CollegeCollection.FindOne(ctx, filter).Decode(&college)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("college not found")
		}
		return nil, errors.New("unable to retrieve college details")
	}
	return &college, nil
}

func (repo *SuperAdminRepository) DeleteCollege(ctx context.Context, collegeID string) error {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	filter := bson.M{"college_id": collegeID}
	res, err := repo.CollegeCollection.DeleteOne(ctx, filter)
	if err != nil {
		return errors.New("failed to delete college")
	}
	if res.DeletedCount == 0 {
		return errors.New("college not found")
	}
	return nil
}

func (r *SuperAdminRepository) InsertCollegeRechargeHistory(ctx context.Context, history model.CollegeRechargeHistory) error {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	_, err := r.CollegeRechargeHistoryCollection.InsertOne(ctx, history)
	if err != nil {
		return fmt.Errorf("failed to insert recharge history: %w", err)
	}

	return nil
}

func (repo *SuperAdminRepository) GetRechargeHistoryByCollegeID(ctx context.Context, collegeID string) ([]model.CollegeRechargeHistory, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	filter := bson.M{"college_id": collegeID}

	cursor, err := repo.CollegeRechargeHistoryCollection.Find(ctx, filter)
	if err != nil {
		return nil, errors.New("failed to retrieve recharge history")
	}
	defer cursor.Close(ctx)

	var recharges []model.CollegeRechargeHistory
	for cursor.Next(ctx) {
		var recharge model.CollegeRechargeHistory
		if err := cursor.Decode(&recharge); err != nil {
			return nil, errors.New("error reading recharge history data")
		}
		recharges = append(recharges, recharge)
	}

	if len(recharges) == 0 {
		return nil, errors.New("no recharge history found for this college")
	}

	return recharges, nil
}

/* Machine management repository methods*/

func (repo *SuperAdminRepository) CreateMachine(ctx context.Context, machine model.Machine) error {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	filter := bson.M{
		"machine_no": machine.MachineNo,
		"college_id": machine.CollegeId,
	}

	var existingMachine model.Machine
	err := repo.MachineCollection.FindOne(ctx, filter).Decode(&existingMachine)

	if err == nil {
		return apperrors.ErrMachineAlreadyExists
	}

	if !errors.Is(err, mongo.ErrNoDocuments) {
		return fmt.Errorf("failed to check existing machine: %w", err)
	}

	_, insertErr := repo.MachineCollection.InsertOne(ctx, machine)
	if insertErr != nil {
		return fmt.Errorf("failed to insert machine: %w", insertErr)
	}

	return nil
}

func (repo *SuperAdminRepository) GetMachinesByCollegeID(ctx context.Context, collegeID string) ([]model.Machine, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	filter := bson.M{"college_id": collegeID}
	cursor, err := repo.MachineCollection.Find(ctx, filter)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, apperrors.ErrCollegeNotFound
		}
		return nil, errors.New("failed to fetch machines")
	}
	defer cursor.Close(ctx)

	var machines []model.Machine
	for cursor.Next(ctx) {
		var machine model.Machine
		if err := cursor.Decode(&machine); err != nil {
			return nil, errors.New("error reading machine data")
		}
		machines = append(machines, machine)
	}

	return machines, nil
}


