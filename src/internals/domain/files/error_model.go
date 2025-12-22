package Files

import "errors"

var (
	ErrInvalidGrade    = errors.New("invalid grade, must be 1PUC or 2PUC")
	ErrInvalidSubject  = errors.New("invalid subject")
	ErrInvalidCopies   = errors.New("copies must be between 1 and 100")
	ErrFileNotFound    = errors.New("File not found")
	ErrDBFailure       = errors.New("database failure")
	ErrInvalidID       = errors.New("invalid object id")
)
