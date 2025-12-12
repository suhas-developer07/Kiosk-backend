package utils

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	validator "github.com/go-playground/validator/v10"
	
	faculty "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/faculties"
	domain "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/files"
)

func ValidateFileInput(f *domain.File) error {

	if strings.TrimSpace(f.Title) == "" {
		return errors.New("file name is required")
	}
	if strings.TrimSpace(f.FileURL) == "" {
		return errors.New("file URL is required")
	}
	if _, err := url.ParseRequestURI(f.FileURL); err != nil {
		return errors.New("invalid file URL")
	}

	return nil
}

var pageRangeRegex = regexp.MustCompile(`^(\d+(-\d+)?)(,\d+(-\d+)?)*$`)

func ValidatePrintJobPayload(p domain.PrintJobPayload) error {
	if p.FileID.IsZero() {
		return errors.New("file_id is required and must be a valid ObjectID")
	}

	if p.FileName == "" {
		return errors.New("missing required field: file_name")
	}

	if len(p.FileName) < 3 {
		return errors.New("file_name must be at least 3 characters long")
	}

	if p.Copies < 1 || p.Copies > 100 {
		return errors.New("copies must be between 1 and 100")
	}

	if p.PrintingSide != "single" && p.PrintingSide != "double" {
		return errors.New("printing_side must be 'single' or 'double'")
	}

	if p.PrintingMode != "color" && p.PrintingMode != "bw" {
		return errors.New("printing_mode must be 'color' or 'bw'")
	}

	if p.PageRange != "" && !pageRangeRegex.MatchString(p.PageRange) {
		return errors.New("invalid page_range format. Example: 1-5 or 2,3,7")
	}

	validPageLayout := map[string]bool{"2-up": true, "4-up": true, "1-up": true}
	if !validPageLayout[p.PageLayout] {
		return errors.New("Page Layout must be one of:2-up,4-up,1-up")
	}

	return nil
}

func ValidateAccountPayload(req faculty.AccoutCreationPayload) error {
	if req.Email == "" {
		return errors.New("email is required")
	}
	if !IsValidEmail(req.Email) {
		return errors.New("invalid email format")
	}
	if req.Password == "" {
		return errors.New("password required for non-google signup")
	}
	return nil
}

func IsValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	return re.MatchString(email)
}



func FormatValidationError(err error) string {
	if errs, ok := err.(validator.ValidationErrors); ok {
		e := errs[0]

		field := e.Field()
		tag := e.Tag()
		param := e.Param()

		switch tag {

		case "required":
			return fmt.Sprintf("missing required field: %s", field)

		case "min":
			return fmt.Sprintf("%s must be at least %s characters long", field, param)

		case "max":
			return fmt.Sprintf("%s must be at most %s characters long", field, param)

		case "oneof":
			return fmt.Sprintf("%s must be one of: %s", field, param)

		// Email rule
		case "email":
			return fmt.Sprintf("%s must be a valid email address", field)

		// Custom validators
		case "objectid":
			return "file_id must be a valid MongoDB ObjectID"

		case "pagerange":
			return "invalid page_range format. Example: 1-5 or 2,3,7"
		}

		// Fallback for unexpected tags
		return fmt.Sprintf("invalid value for field '%s'", field)
	}

	// Not a validator.ValidationErrors type
	return "invalid request payload"
}