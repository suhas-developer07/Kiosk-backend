package Files

import "errors"

var (
    ErrInvalidGrade   = errors.New("invalid grade, must be 1PUC or 2PUC")
    ErrInvalidSubject = errors.New("invalid subject")
)
