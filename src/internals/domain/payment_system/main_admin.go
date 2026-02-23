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
	CollegeID      string `json:"college_id" validate:"required"`
	MachineID      string `json:"machine_id" validate:"required"`
	RechargeAmount string `json:"recharge_amount" validate:"required"`
}

type RechargeRFIDRequest struct {
	UserID         string `json:"user_id" bson:"user_id"`
	MachineID      string `json:"machine_id" bson:"machine_id" validate:"required"`
	CollegeID      string `json:"college_id" bson:"college_id" validate:"required"`
	RechargeAmount string `json:"recharge_amount" bson:"recharge_amount" validate:"required"`
}

type RechargerRFIDHistory struct {
	SuperAdminID   string `json:"super_admin_id" bson:"super_admin_id"`
	CollegeID      string `json:"college_id" bson:"college_id"`
	MachineID      string `json:"machine_id" bson:"machine_id"`
	MachineName    string `json:"machine_name" bson:"machine_name"`
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

/*
 RECHARGE MACHINE USER MODELS
*/

type User struct {
	UserID      string `json:"user_id" bson:"user_id"`
	MachineID   string `json:"machine_id" bson:"machine_id"`
	MachineName string `json:"machine_name"`
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
