package paymentsystem

import (
	"context"
	"errors"
	"strconv"
	"strings"

	model "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/payment_system"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	SUPER_ADMIN_COLLECTION              = "super_admins"
	COLLEGE_COLLECTION                  = "colleges"
	RECHARGE_COLLEGE_HISTORY_COLLECTION = "college_recharge_history"
)

type SuperAdminRepository struct {
	client                           *mongo.Client
	SuperAdminCollection             *mongo.Collection
	CollegeCollection                *mongo.Collection
	RechargeCollegeHistoryCollection *mongo.Collection
}

func NewSuperAdminRepo(db *mongo.Database, client *mongo.Client) *SuperAdminRepository {
	return &SuperAdminRepository{
		client:                           client,
		SuperAdminCollection:             db.Collection(SUPER_ADMIN_COLLECTION),
		CollegeCollection:                db.Collection(COLLEGE_COLLECTION),
		RechargeCollegeHistoryCollection: db.Collection(RECHARGE_COLLEGE_HISTORY_COLLECTION),
	}
}

/*
Not implemented yet
func (repo *SuperAdminRepository) CreateSuperAdmin(ctx context.Context, superadmin domain.SuperAdminCreateRequest) error {
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
*/

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

func (repo *SuperAdminRepository) RechargeCollege(ctx context.Context, recharge model.CollegeRecharge, superAdminID string) error {
	// 1. Fetch superadmin
	superAdminFilter := bson.M{"super_admin_id": superAdminID}
	var superAdmin model.SuperAdmin
	err := repo.SuperAdminCollection.FindOne(ctx, superAdminFilter).Decode(&superAdmin)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("superadmin not found")
		}
		return errors.New("unable to fetch superadmin balance")
	}

	// 2. Validate recharge amount (must be integer, not negative or float)
	if strings.Contains(recharge.RechargeAmount, ".") {
		return errors.New("recharge amount must be a whole number")
	}

	rechargeAmountInt, err := strconv.Atoi(recharge.RechargeAmount)
	if err != nil || rechargeAmountInt <= 0 {
		return errors.New("invalid recharge amount, must be positive integer")
	}

	// 3. Check superadmin balance
	superBalanceInt, err := strconv.Atoi(superAdmin.Balance)
	if err != nil {
		return errors.New("invalid superadmin balance format in database")
	}
	if rechargeAmountInt > superBalanceInt {
		return errors.New("insufficient superadmin balance for recharge")
	}

	// 4. Fetch college
	var college model.SuperAdminCollege
	collegeFilter := bson.M{"college_id": recharge.CollegeID}
	err = repo.CollegeCollection.FindOne(ctx, collegeFilter).Decode(&college)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("college not found")
		}
		return errors.New("unable to retrieve college details")
	}

	// 5. Calculate new balances
	oldCollegeBalanceInt, err := strconv.Atoi(college.Balance)
	if err != nil {
		return errors.New("invalid college balance format in database")
	}

	newCollegeBalance := strconv.Itoa(oldCollegeBalanceInt + rechargeAmountInt)
	newSuperBalance := strconv.Itoa(superBalanceInt - rechargeAmountInt)

	// 6. Update both balances in DB
	_, err = repo.CollegeCollection.UpdateOne(ctx, collegeFilter, bson.M{"$set": bson.M{"balance": newCollegeBalance}})
	if err != nil {
		return errors.New("failed to update college balance")
	}

	_, err = repo.CollegeCollection.Database().Collection("super_admin").UpdateOne(ctx, superAdminFilter, bson.M{"$set": bson.M{"balance": newSuperBalance}})
	if err != nil {
		return errors.New("failed to update superadmin balance")
	}

	// 7. Record recharge transaction
	_, err = repo.RechargeCollegeHistoryCollection.InsertOne(ctx, recharge)
	if err != nil {
		return errors.New("failed to record recharge transaction")
	}

	return nil
}

func (repo *SuperAdminRepository) GetRechargeHistoryByCollegeID(ctx context.Context, collegeID string) ([]model.CollegeRecharge, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	filter := bson.M{"college_id": collegeID}

	cursor, err := repo.RechargeCollegeHistoryCollection.Find(ctx, filter)
	if err != nil {
		return nil, errors.New("failed to retrieve recharge history")
	}
	defer cursor.Close(ctx)

	var recharges []model.CollegeRecharge
	for cursor.Next(ctx) {
		var recharge model.CollegeRecharge
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
