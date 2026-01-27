package faculties

import "errors"

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrInvalidID          = errors.New("invalid object_id")
	ErrFacultyNotFound    = errors.New("Faculty not found")
)
