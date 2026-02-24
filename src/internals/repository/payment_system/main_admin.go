package paymentsystem

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	apperrors "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/errors"
	model "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/payment_system"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	defaultQueryTimeout = 5 * time.Second

	collectionMachines            = "recharge_machines"
	collectionRechargeHistory     = "recharge_machine_history"
	collectionRechargeRFIDHistory = "rfid_recharge_history"
	users                         = "recharge_machine_users"
)

type MainAdminRepo struct {
	client                               *mongo.Client
	MachineCollection                    *mongo.Collection
	RechargeMachineHistoryCollection     *mongo.Collection
	RFIDRechargeMachineHistoryCollection *mongo.Collection
	UsersCollection                      *mongo.Collection
	CollegeCollection                    *mongo.Collection
}

func NewMainAdminRepo(db *mongo.Database, client *mongo.Client) *MainAdminRepo {
	return &MainAdminRepo{
		client:                               client,
		MachineCollection:                    db.Collection(collectionMachines),
		RechargeMachineHistoryCollection:     db.Collection(collectionRechargeHistory),
		RFIDRechargeMachineHistoryCollection: db.Collection(collectionRechargeRFIDHistory),
		UsersCollection:                      db.Collection(users),
		CollegeCollection:                    db.Collection(collectionColleges),
	}
}
/* College Repository methods */

func (r *MainAdminRepo) GetCollegeForLogin(ctx context.Context, email string) (collegeID, collegename, hashedPassword, superadminID string, err error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	filter := bson.M{"college_email": email}
	var college model.SuperAdminCollege
	err = r.CollegeCollection.FindOne(ctx, filter).Decode(&college)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", "", "", "", errors.New("college not found")
		}
		return "", "", "", "", errors.New("unable to retrieve college login details")
	}
	return college.CollegeID, college.CollegeName, college.CollegePassword, college.SuperAdminId, nil
}

func (r *MainAdminRepo) GetCollegeBalance(ctx context.Context, collegeID string) (string, error) {
	var college struct {
		Balance string `bson:"balance"`
	}
	err := r.CollegeCollection.FindOne(ctx, bson.M{"college_id": collegeID}).Decode(&college)
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

func (r *MainAdminRepo) UpdateCollegeBalance(ctx context.Context, collegeID string, newBalance string) error {
	if _, err := strconv.ParseFloat(newBalance, 64); err != nil {
		return errors.New("invalid balance value provided")
	}
	res, err := r.CollegeCollection.UpdateOne(ctx,
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

/*Recharge machine Repository methods */
func (repo *MainAdminRepo) GetMachinesByCollegeID(ctx context.Context, collegeID string) ([]model.Machine, error) {
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


func (r *MainAdminRepo) GetMachineBalance(ctx context.Context, machineID string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	var result struct {
		Balance string `bson:"balance"`
	}

	filter := bson.M{"machine_id": machineID}
	err := r.MachineCollection.FindOne(ctx, filter).Decode(&result)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", apperrors.ErrFileNotFound
		}
		return "", fmt.Errorf("failed to fetch machine balance: %w", err)
	}

	// Validate balance format
	if _, parseErr := strconv.ParseFloat(result.Balance, 64); parseErr != nil {
		return "", errors.New("machine balance is in an invalid format")
	}

	return result.Balance, nil
}

func (r *MainAdminRepo) GetMachineDetails(ctx context.Context, machineID string) (*model.Machine, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	var machine model.Machine

	filter := bson.M{"machine_id": machineID}
	err := r.MachineCollection.FindOne(ctx, filter).Decode(&machine)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, apperrors.ErrFileNotFound
		}
		return nil, fmt.Errorf("failed to fetch machine details: %w", err)
	}

	return &machine, nil
}

func (r *MainAdminRepo) UpdateMachineBalance(ctx context.Context, machineID string, newBalance string) error {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	// Validate balance format
	if _, err := strconv.ParseFloat(newBalance, 64); err != nil {
		return errors.New("invalid balance value provided")
	}

	filter := bson.M{"machine_id": machineID}
	update := bson.M{"$set": bson.M{"balance": newBalance}}

	result, err := r.MachineCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update machine balance: %w", err)
	}

	if result.MatchedCount == 0 {
		return apperrors.ErrFileNotFound
	}

	return nil
}

func (r *MainAdminRepo) InsertRechargeHistory(ctx context.Context, history model.MachineRechargeHistory) error {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	_, err := r.RechargeMachineHistoryCollection.InsertOne(ctx, history)
	if err != nil {
		return fmt.Errorf("failed to insert recharge history: %w", err)
	}

	return nil
}

func (r *MainAdminRepo) GetRechargeMachineHistory(ctx context.Context, machineID string) ([]model.MachineRechargeHistory, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	var history []model.MachineRechargeHistory

	filter := bson.M{"machine_id": machineID}
	cursor, err := r.RechargeMachineHistoryCollection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query machine recharge history: %w", err)
	}
	defer cursor.Close(ctx)

	// Decode all results
	for cursor.Next(ctx) {
		var record model.MachineRechargeHistory
		if decodeErr := cursor.Decode(&record); decodeErr != nil {
			return nil, fmt.Errorf("failed to decode recharge history record: %w", decodeErr)
		}
		history = append(history, record)
	}

	// Check for cursor errors
	if cursorErr := cursor.Err(); cursorErr != nil {
		return nil, fmt.Errorf("cursor error while reading recharge history: %w", cursorErr)
	}

	return history, nil
}

func (r *MainAdminRepo) GetMachineDetailsByMachineNo(ctx context.Context, machineNo string) (model.Machine, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	var machine model.Machine

	filter := bson.M{"machine_no": machineNo}
	err := r.MachineCollection.FindOne(ctx, filter).Decode(&machine)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return model.Machine{}, apperrors.ErrFileNotFound
		}
		return model.Machine{}, fmt.Errorf("failed to fetch machine by number: %w", err)
	}

	return machine, nil
}

func (r *MainAdminRepo) GetMachineNameByID(ctx context.Context, machineID string) string {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	var result struct {
		MachineName string `bson:"machine_name"`
	}

	filter := bson.M{"machine_id": machineID}
	err := r.MachineCollection.FindOne(ctx, filter).Decode(&result)

	if err != nil {
		return ""
	}

	return result.MachineName
}

func (r *MainAdminRepo) GetAvailableMachines(ctx context.Context) ([]model.Machine, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	var machines []model.Machine

	cursor, err := r.MachineCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to query available machines: %w", err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var machine model.Machine
		if decodeErr := cursor.Decode(&machine); decodeErr != nil {
			return nil, fmt.Errorf("failed to decode machine record: %w", decodeErr)
		}
		machines = append(machines, machine)
	}
	
	if cursorErr := cursor.Err(); cursorErr != nil {
		return nil, fmt.Errorf("cursor error while reading machines: %w", cursorErr)
	}

	return machines, nil
}

/*RFID Recharge Repository methods */
func (r *MainAdminRepo) InsertRechargeRFIDHistory(ctx context.Context, history model.RechargerRFIDHistory) error {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	var user struct {
		UserName string `bson:"user_name"`
	}

	userFilter := bson.M{"user_id": history.UserID}
	userErr := r.UsersCollection.FindOne(ctx, userFilter).Decode(&user)

	if userErr != nil {
		if errors.Is(userErr, mongo.ErrNoDocuments) {
			history.UserName = ""
		} else {
			return fmt.Errorf("failed to fetch user details: %w", userErr)
		}
	} else {
		history.UserName = user.UserName
	}

	_, insertErr := r.RFIDRechargeMachineHistoryCollection.InsertOne(ctx, history)
	if insertErr != nil {
		return fmt.Errorf("failed to insert RFID recharge history: %w", insertErr)
	}

	return nil
}

func (r *MainAdminRepo) GetRFIDRechargeHistory(ctx context.Context, machineID string) ([]model.RechargerRFIDHistory, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	var history []model.RechargerRFIDHistory

	filter := bson.M{"machine_id": machineID}
	cursor, err := r.RFIDRechargeMachineHistoryCollection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query RFID recharge history: %w", err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var record model.RechargerRFIDHistory
		if decodeErr := cursor.Decode(&record); decodeErr != nil {
			return nil, fmt.Errorf("failed to decode RFID recharge history record: %w", decodeErr)
		}
		history = append(history, record)
	}

	if cursorErr := cursor.Err(); cursorErr != nil {
		return nil, fmt.Errorf("cursor error while reading RFID recharge history: %w", cursorErr)
	}

	return history, nil
}

/* Recharge machine user repository methods */
func (r *MainAdminRepo) GetRechargeMachineUserByEmail(ctx context.Context, email string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	var warden model.User

	filter := bson.M{"email": email}
	err := r.UsersCollection.FindOne(ctx, filter).Decode(&warden)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, apperrors.ErrFacultyNotFound
		}
		return nil, fmt.Errorf("failed to query warden by email: %w", err)
	}

	return &warden, nil
}

// CreateUser creates a new warden user in the database
func (r *MainAdminRepo) CreateRechargeMachineUser(ctx context.Context, user *model.User) error {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	_, err := r.UsersCollection.InsertOne(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	return nil
}

func (r *MainAdminRepo) GetRechargeMachineUserById(ctx context.Context, id string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	opts := options.FindOne().SetProjection(bson.M{
		"user_name": 1,
		"_id":       0,
	})

	var result struct {
		Username string `bson:"user_name"`
	}

	filter := bson.M{
		"user_id": id,
	}

	err := r.UsersCollection.FindOne(ctx, filter, opts).Decode(&result)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", apperrors.ErrWardenNotFound
		}
		return "", fmt.Errorf("failed to get username by id: %w", err)
	}
	return result.Username, nil
}

