package apperrors

import "errors"

/*
	Input / Validation errors
*/
var (
	ErrInvalidInput     = errors.New("invalid input")
	ErrInvalidID        = errors.New("invalid id")
	ErrInvalidGrade     = errors.New("invalid grade")
	ErrInvalidSubject   = errors.New("invalid subject")
	ErrInvalidCopies    = errors.New("invalid copies")
	ErrInvalidPassword  = errors.New("invalid password")
)

/*
	Resource / Not found errors
*/
var (
	ErrFileNotFound     = errors.New("file not found")
	ErrFacultyNotFound  = errors.New("faculty not found")
	ErrAdminNotFound    = errors.New("Admin not found")
)

/*
	Conflict / State errors
*/
var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrNoPendingRequest   = errors.New("no pending delete request")
	ErrConflict           = errors.New("conflict")
)

/*
	Infrastructure / Internal errors
*/
var (
	ErrDBFailure = errors.New("database failure")
	ErrInternal  = errors.New("internal server error")
)
