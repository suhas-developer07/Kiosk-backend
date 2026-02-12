package apperrors

import "errors"

/*
Input / Validation errors
*/
var (
	ErrInvalidInput         = errors.New("invalid input")
	ErrInvalidID            = errors.New("invalid id")
	ErrInvalidGrade         = errors.New("invalid grade")
	ErrInvalidSubject       = errors.New("invalid subject")
	ErrInvalidCopies        = errors.New("invalid copies")
	ErrInvalidPassword      = errors.New("invalid password")
	ErrInvalidClassHandling = errors.New("invalid class handling - class must be (1PUC or 2PUC)")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrInvalidAmount        = errors.New("invalid amount")
	ErrInvalidMachineNumber = errors.New("invalid machine number")
)

/*
Resource / Not found errors
*/
var (
	ErrFileNotFound    = errors.New("file not found")
	ErrFacultyNotFound = errors.New("faculty not found")
	ErrAdminNotFound   = errors.New("admin not found")
	ErrMachineNotFound = errors.New("machine not found")
	ErrUserNotFound    = errors.New("user not found")
	ErrHistoryNotFound = errors.New("history not found")
	ErrWardenNotFound  = errors.New("warden not found")
)

/*
Conflict / State errors
*/
var (
	ErrEmailAlreadyExists   = errors.New("email already exists")
	ErrNoPendingRequest     = errors.New("no pending delete request")
	ErrConflict             = errors.New("conflict")
	ErrMachineAlreadyExists = errors.New("a machine with this number already exists")
	ErrInsufficientBalance  = errors.New("insufficient balance")
	ErrDuplicateTransaction = errors.New("duplicate transaction detected")
)

/*
Infrastructure / Internal errors
*/
var (
	ErrDBFailure           = errors.New("database failure")
	ErrInternal            = errors.New("internal server error")
	ErrConnectionFailed    = errors.New("connection failed")
	ErrTimeout             = errors.New("operation timeout")
	ErrTransactionRollback = errors.New("transaction rollback")
)
