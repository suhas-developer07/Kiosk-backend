package rechargemachine

import (
	"time"
)

type MainAdmin struct {
	ID        string    `bson:"_id,omitempty" json:"id,omitempty"`
	Username  string    `bson:"username" json:"username"`
	Email     string    `bson:"email" json:"email"`
	Password  string    `bson:"password" json:"password"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

type Machine struct {
	MachineID   string `bson:"machine_id,omitempty" json:"machine_id,omitempty"`
	MachineNo   string `json:"machine_no" bson:"machine_no" validate:"required"`
	MachineName string `json:"machine_name" bson:"machine_name" validate:"required"`
	Balance     string `json:"balance" bson:"balance"`
}

type MachineRechargeHistory struct {
	MachineID      string `json:"machine_id" bson:"machine_id"`
	RechargeAmount string `json:"recharge_amount" bson:"recharge_amount"`
	Date           string `json:"date" bson:"date"`
	Time           string `json:"time" bson:"time"`
}

type MachineCreateRequest struct {
	MachineNo   string `json:"machine_no" bson:"machine_no" validate:"required"`
	MachineName string `json:"machine_name" bson:"machine_name" validate:"required"`
}

type MachineRechargeRequest struct {
	MachineID      string `json:"machine_id" validate:"required"`
	RechargeAmount string `json:"recharge_amount" validate:"required"`
}

type RechargeRFIDRequest struct {
	UserID         string `json:"user_id" bson:"user_id"`
	MachineID      string `json:"machine_id" bson:"machine_id" validate:"required"`
	RechargeAmount string `json:"recharge_amount" bson:"recharge_amount" validate:"required"`
}

type RechargerRFIDHistory struct {
	MachineID      string `json:"machine_id" bson:"machine_id"`
	UserID         string `json:"user_id" bson:"user_id"`
	UserName       string `json:"user_name" bson:"user_name"`
	RechargeAmount string `json:"recharge_amount" bson:"recharge_amount"`
	Date           string `json:"date" bson:"date"`
	Time           string `json:"time" bson:"time"`
}

type MachineBalanceResponse struct {
	UserID    string `json:"user_id" bson:"user_id"`
	MachineID string `json:"machine_id" bson:"machine_id"`
	Balance   string `json:"balance" bson:"balance"`
}

type CreateMainAdminPayload struct {
	Name     string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type SigninMainAdminPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

//warden models

type User struct {
	UserID    string `json:"user_id" bson:"user_id"`
	MachineID string `json:"machine_id" bson:"machine_id"`
	MachineName string `json:"machine_name"`
	UserName  string `json:"user_name" bson:"user_name" validate:"required"`
	Password  string `json:"password" bson:"password" validate:"required"`
	Email     string `json:"email" bson:"email" validate:"required,email"`
	CreatedAt string `json:"created_at" bson:"created_at"`
}

type UserAccessCreateRequest struct {
	MachineId string `json:"machine_id" validate:"required"`
	UserId    string `json:"user_id"`
	UserName  string `json:"user_name" validate:"required"`
	Password  string `json:"password" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
}

type UserAccessLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UserAccessDeleteRequest struct {
	UserId string `json:"user_id" validate:"required"`
}
