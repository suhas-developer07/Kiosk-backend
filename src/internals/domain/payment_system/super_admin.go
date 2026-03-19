package paymentsystem

/*
SUPER ADMIN AUTH MODELS
*/
type SuperAdmin struct {
	SuperAdminId       string `json:"super_admin_id" bson:"super_admin_id"`
	SuperAdminName     string `json:"super_admin_name" bson:"super_admin_name" validate:"required"`
	SuperAdminEmail    string `json:"super_admin_email" bson:"super_admin_email" validate:"required,email"`
	SuperAdminPassword string `json:"super_admin_password" bson:"super_admin_password" validate:"required,min=8,strongpassword"`
	CreatedAt          string `json:"created_at" bson:"created_at"`
}

type SuperAdminCreateRequest struct {
	SuperAdminId       string `json:"super_admin_id" bson:"super_admin_id"`
	SuperAdminName     string `json:"super_admin_name" bson:"super_admin_name" validate:"required"`
	SuperAdminEmail    string `json:"super_admin_email" bson:"super_admin_email" validate:"required,email"`
	SuperAdminPassword string `json:"super_admin_password" bson:"super_admin_password" validate:"required,min=6"`
	Balance            string `json:"balance" bson:"balance"`
}

type SuperAdminLoginRequest struct {
	SuperAdminEmail    string `json:"super_admin_email" bson:"super_admin_email" validate:"required,email"`
	SuperAdminPassword string `json:"super_admin_password" bson:"super_admin_password" validate:"required"`
}

type SuperAdminResponse struct {
	Token        string `json:"token"`
	BalanceToken string `json:"balance_token"`
	Balance      string `json:"balance"`
}

type SuperAdminBalanceToken struct {
	Token string `json:"token"`
}

/*
COLLEGE AND MACHINE MODELS
*/

type SuperAdminCollege struct {
	SuperAdminId    string `json:"super_admin_id" bson:"super_admin_id"`
	CollegeID       string `json:"college_id" bson:"college_id"`
	CollegeName     string `json:"college_name" bson:"college_name" validate:"required"`
	CollegeEmail    string `json:"college_email" bson:"college_email" validate:"required,email"`
	CollegePhone    string `json:"college_phone" bson:"college_phone" validate:"required,e164"`
	CollegePassword string `json:"college_password" bson:"college_password" validate:"required,min=8,strongpassword"`
	CollegeAddress  string `json:"college_address" bson:"college_address" validate:"required"`
	Balance         string `json:"balance" bson:"balance"`
	CreatedAt       string `json:"created_at" bson:"created_at"`
}

type CollegeRechargeHistory struct {
	RechargeID     string `json:"recharge_id" bson:"recharge_id"`
	CollegeID      string `json:"college_id" bson:"college_id"`
	CollegeName    string `json:"college_name" bson:"college_name"`
	SuperAdminId   string `json:"super_admin_id" bson:"super_admin_id"`
	RechargeAmount string `json:"recharge_amount" bson:"recharge_amount"`
	Date           string `json:"date" bson:"date"`
	Time           string `json:"time" bson:"time"`
}

type CollegeCreateRequest struct {
	CollegeName     string `json:"college_name" bson:"college_name" validate:"required"`
	CollegeEmail    string `json:"college_email" bson:"college_email" validate:"required,email"`
	CollegePhone    string `json:"college_phone" bson:"college_phone" validate:"required"`
	CollegePassword string `json:"college_password" bson:"college_password" validate:"required,min=8"`
	CollegeAddress  string `json:"college_address" bson:"college_address" validate:"required"`
}

type CollegeResponse struct {
	SuperAdminId string `json:"super_admin_id"`
	CollegeID    string `json:"college_id"`
	CollegeName  string `json:"college_name"`
	Balance      string `json:"balance"`
}

type CollegeLoginRequest struct {
	CollegeEmail    string `json:"college_email" bson:"college_email" validate:"required,email"`
	CollegePassword string `json:"college_password" bson:"college_password" validate:"required,min=8,strongpassword"`
}

type CollegeTokenResponse struct {
	Token       string `json:"token"`
	CollegeName string `json:"college_name"`
	Balance     string `json:"balance"`
}

type CollegeRechargeRequest struct {
	SuperAdminId   string `json:"super_admin_id" bson:"super_admin_id" validate:"required"`
	CollegeID      string `json:"college_id" bson:"college_id" validate:"required"`
	RechargeAmount string `json:"recharge_amount" bson:"recharge_amount" validate:"required"`
}

type CollegeRechargeHistoryResponse struct {
	Status  string                   `json:"status"`
	Message string                   `json:"message"`
	Data    []CollegeRechargeHistory `json:"data"`
}

/* Machine management models */

type Machine struct {
	MachineID    string `bson:"machine_id,omitempty" json:"machine_id,omitempty"`
	MachineNo    string `json:"machine_no" bson:"machine_no" validate:"required"`
	MachineName  string `json:"machine_name" bson:"machine_name" validate:"required"`
	CollegeId    string `json:"college_id" bson:"college_id"`
	SuperAdminId string `json:"super_admin_id" bson:"super_admin_id"`
	Balance      string `json:"balance" bson:"balance"`
}

type MachineCreateRequest struct {
	MachineNo   string `json:"machine_no" bson:"machine_no" validate:"required"`
	MachineName string `json:"machine_name" bson:"machine_name" validate:"required"`
	CollegeId   string `json:"college_id" bson:"college_id" validate:"required"`
}

type UpdateRFIDCard struct {
	USN string `json:"usn"`
}
