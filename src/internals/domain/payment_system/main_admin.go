package paymentsystem

/*
RECHARGE MACHINE AND RFID MODELS
*/

type MachineRechargeHistory struct {
	SuperAdminID   string `json:"super_admin_id" bson:"super_admin_id"`
	CollegeID      string `json:"college_id" bson:"college_id"`
	MachineID      string `json:"machine_id" bson:"machine_id"`
	MachineName    string `json:"machine_name" bson:"machine_name"`
	RechargeAmount string `json:"recharge_amount" bson:"recharge_amount"`
	Date           string `json:"date" bson:"date"`
	Time           string `json:"time" bson:"time"`
}

type MachineRechargeRequest struct {
	MachineID      string `json:"machine_id" validate:"required"`
	RechargeAmount string `json:"recharge_amount" validate:"required"`
}

type RechargeRFIDRequest struct {
	UserID         string `json:"user_id" bson:"user_id"`
	CardID         string `json:"card_id" bson:"card_id"`
	MachineID      string `json:"machine_id" bson:"machine_id" validate:"required"`
	RechargeAmount string `json:"recharge_amount" bson:"recharge_amount" validate:"required"`
}

type RechargerRFIDHistory struct {
	SuperAdminID   string `json:"super_admin_id" bson:"super_admin_id"`
	CollegeID      string `json:"college_id" bson:"college_id"`
	MachineID      string `json:"machine_id" bson:"machine_id"`
	MachineName    string `json:"machine_name" bson:"machine_name"`
	UserID         string `json:"user_id" bson:"user_id"`
	UserName       string `json:"user_name" bson:"user_name"`
	CardID         string `json:"card_id" bson:"card_id"`
	USN            string `json:"usn" bson:"usn"`
	RechargeAmount string `json:"recharge_amount" bson:"recharge_amount"`
	Date           string `json:"date" bson:"date"`
	Time           string `json:"time" bson:"time"`
}

type MachineBalanceResponse struct {
	UserID    string `json:"user_id" bson:"user_id"`
	MachineID string `json:"machine_id" bson:"machine_id"`
	Balance   string `json:"balance" bson:"balance"`
}

type RFIDCard struct {
	CardID    string `json:"card_id" bson:"card_id"`
	USN       string `json:"usn" bson:"usn"`
	Balance   string `json:"balance" bson:"balance"`
	CollegeID string `json:"college_id" bson:"college_id"`
}

type InitializeCardRequest struct {
	MachineID      string `json:"machine_id" bson:"machine_id" validate:"required"`
	CardID         string `json:"card_id" bson:"card_id" validate:"required"`
	UserID         string `json:"user_id" bson:"user_id" validate:"required"`
	USN            string `json:"usn" bson:"usn" validate:"required"`
	RechargeAmount string `json:"recharge_amount" bson:"recharge_amount" validate:"required"`
}

/*
 RECHARGE MACHINE USER MODELS
*/

type User struct {
	UserID      string `json:"user_id" bson:"user_id"`
	MachineID   string `json:"machine_id" bson:"machine_id"`
	MachineName string `json:"machine_name"`
	College_id  string `json:"college_id"`
	UserName    string `json:"user_name" bson:"user_name" validate:"required"`
	Password    string `json:"password" bson:"password" validate:"required"`
	Email       string `json:"email" bson:"email" validate:"required,email"`
	CreatedAt   string `json:"created_at" bson:"created_at"`
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
