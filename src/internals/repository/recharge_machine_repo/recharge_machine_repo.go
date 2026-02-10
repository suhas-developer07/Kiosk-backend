package rechargemachinerepo

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	apperrors "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/errors"
	model "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/recharge_machine"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type RechargeMachineRepo struct {
	client                    *mongo.Client
	MachineCollection         *mongo.Collection
	RechargeHistoryCollection *mongo.Collection
	RechargeRFIDHistoryCollection *mongo.Collection
	WardensCollection         *mongo.Collection
	MainAdminCollection       *mongo.Collection
}

func NewMachineRechargeRepo(db *mongo.Database, client *mongo.Client) *RechargeMachineRepo {
	return &RechargeMachineRepo{
		client: client,
		MachineCollection: db.Collection("machine_collection"),
		RechargeHistoryCollection: db.Collection("Recharge_History_Collection"),
		RechargeRFIDHistoryCollection: db.Collection("Recharge_RFID_History_Collection"),
		WardensCollection: db.Collection("Wardens_Collection"),
		MainAdminCollection: db.Collection("Main_Admin_Collection"),
	}
}

func (r *RechargeMachineRepo) CreateAccount(ctx context.Context, req model.MainAdmin) error {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"email": req.Email}

	var exists struct {
		ID primitive.ObjectID `bson:"_id"`
	}

	err := r.MainAdminCollection.FindOne(ctx, filter).Decode(&exists)

	switch {
	case errors.Is(err, mongo.ErrNoDocuments):
		_, err := r.MainAdminCollection.InsertOne(ctx, req)
		if err != nil {
			return fmt.Errorf("db error while creating account: %w", err)
		}
		return nil

	case err != nil:
		return fmt.Errorf("db error while checking email: %w", err)

	default:
		return apperrors.ErrEmailAlreadyExists
	}
}

func (r *RechargeMachineRepo) CreateMachine(ctx context.Context, machine model.Machine) error {
	filter := bson.M{
		"machine_no": machine.MachineNo,
	}

	var existing model.Machine
	err := r.MachineCollection.FindOne(ctx, filter).Decode(&existing)
	if err == nil {
		return errors.New("a machine with this number already exists for the selected college")
	}
	if err != mongo.ErrNoDocuments {
		return errors.New("unable to check for existing machines. Please try again later")
	}

	machine.Balance = "0"
	_, err = r.MachineCollection.InsertOne(ctx, machine)
	if err != nil {
		return errors.New("failed to create the machine. Please try again later")
	}
	return nil
}


func (r *RechargeMachineRepo) GetMachineBalance(ctx context.Context, machineID string) (string, error) {
	var machine struct {
		Balance string `bson:"balance"`
	}
	err := r.MachineCollection.FindOne(ctx, bson.M{"machine_id": machineID}).Decode(&machine)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", errors.New("machine not found")
		}
		return "", errors.New("unable to fetch machine balance at the moment")
	}
	if _, err := strconv.ParseFloat(machine.Balance, 64); err != nil {
		return "", errors.New("stored machine balance is in an invalid format")
	}
	return machine.Balance, nil
}

func (r *RechargeMachineRepo) GetMachineDetails(ctx context.Context, machineID string) (*model.Machine, error) {
	var machine model.Machine
	err := r.MachineCollection.FindOne(ctx, bson.M{"_id": machineID}).Decode(&machine)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("machine not found")
		}
		return nil, errors.New("unable to fetch machine details at the moment")
	}
	return &machine, nil
}

func (r *RechargeMachineRepo) UpdateMachineBalance(ctx context.Context, machineID string, newBalance string) error {
	if _, err := strconv.ParseFloat(newBalance, 64); err != nil {
		return errors.New("invalid balance value provided")
	}
	res, err := r.MachineCollection.UpdateOne(ctx,
		bson.M{"machine_id": machineID},
		bson.M{"$set": bson.M{"balance": newBalance}},
	)
	if err != nil {
		return errors.New("failed to update machine balance due to a database error")
	}
	if res.MatchedCount == 0 {
		return errors.New("machine not found")
	}
	return nil
}

func (r *RechargeMachineRepo) InsertRechargeHistory(ctx context.Context, history model.MachineRechargeHistory) error {
	_, err := r.RechargeHistoryCollection.InsertOne(ctx, history)
	if err != nil {
		return errors.New("failed to save recharge history")
	}
	return nil
}

func (r *RechargeMachineRepo) InsertRechargeRFIDHistory(ctx context.Context, history model.RechargerRFIDHistory) error {
	var user struct {
		UserName string `bson:"user_name"`
	}
	err := r.WardensCollection.FindOne(ctx, bson.M{"user_id": history.UserID}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {

			history.UserName = ""
		} else {
			return errors.New("failed to fetch warden details from wardens collection")
		}
	} else {

		history.UserName = user.UserName
	}

	_, err = r.RechargeRFIDHistoryCollection.InsertOne(ctx, history)
	if err != nil {
		return errors.New("failed to save recharge history")
	}
	return nil
}

func (r *RechargeMachineRepo) GetRFIDRechargeHistory(ctx context.Context, machineID string) ([]model.RechargerRFIDHistory, error) {
	var userMachineRechargeHistory []model.RechargerRFIDHistory
	cursor, err := r.RechargeRFIDHistoryCollection.Find(ctx, bson.M{"machine_id": machineID})
	if err != nil {
		return nil, errors.New("unable to retrieve user recharge history")
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var userMachineRechargeHistoryItem model.RechargerRFIDHistory
		if err := cursor.Decode(&userMachineRechargeHistoryItem); err != nil {
			return nil, errors.New("error reading user recharge history data")
		}
		userMachineRechargeHistory = append(userMachineRechargeHistory, userMachineRechargeHistoryItem)
	}
	if err := cursor.Err(); err != nil {
		return nil, errors.New("error occurred while processing user recharge history records")
	}
	return userMachineRechargeHistory, nil
}

func (r *RechargeMachineRepo) GetRechargeMachineHistory(ctx context.Context, machineID string) ([]model.MachineRechargeHistory, error) {
	var rechargeHistory []model.MachineRechargeHistory
	cursor, err := r.RechargeHistoryCollection.Find(ctx, bson.M{"machine_id": machineID})
	if err != nil {
		return nil, errors.New("unable to retrieve recharge history")
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var historyItem model.MachineRechargeHistory
		if err := cursor.Decode(&historyItem); err != nil {
			return nil, errors.New("error reading recharge history data")
		}
		rechargeHistory = append(rechargeHistory, historyItem)
	}
	if err := cursor.Err(); err != nil {
		return nil, errors.New("error occurred while processing recharge history records")
	}
	return rechargeHistory, nil
}

func (r *RechargeMachineRepo) GetWardenByEmail(ctx context.Context, email string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"email": email}
	var warden  model.User

	err := r.WardensCollection.FindOne(ctx, filter).Decode(&warden)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, apperrors.ErrFacultyNotFound
		}
		return nil, fmt.Errorf("db error while checking email: %w", err)
	}

	return &warden, nil
}

func (r *RechargeMachineRepo) CreateUser(ctx context.Context, user *model.User) error {
	_, err := r.WardensCollection.InsertOne(ctx, user)
	if err != nil {
		return errors.New("failed to create user")
	}
	return nil
}

func (r *RechargeMachineRepo) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.WardensCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("user not found")
		}
		return nil, errors.New("unable to fetch user details")
	}
	return &user, nil
}

func (r *RechargeMachineRepo) GetMachineDetailsByMachineNo(ctx context.Context, machineNo string) (model.Machine, error) {
	var machine model.Machine

	err := r.MachineCollection.FindOne(ctx, bson.M{"machine_no": machineNo}).Decode(&machine)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return model.Machine{}, errors.New("machine not found")
		}
		return model.Machine{}, errors.New("unable to fetch machine details at the moment")
	}
	return machine, nil
}

func (r *RechargeMachineRepo) GetMachineNameByID(ctx context.Context, machineID string) string {
	var machine struct {
	MachineName string `bson:"machine_name"`
	}
	err := r.MachineCollection.FindOne(ctx, bson.M{"machine_id": machineID}).Decode(&machine)
	if err != nil {
		return ""
	}
	return machine.MachineName
}