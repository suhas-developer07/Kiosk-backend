package utils

import (
	"errors"
	"net/url"
	"regexp"
	"strings"

	domain "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/Files"
)

func ValidateFileInput(f *domain.File) error{

	if strings.TrimSpace(f.Title) == ""{
		return errors.New("file name is required")
	}
	if strings.TrimSpace(f.FileURL) == ""{
		return errors.New("file URL is required")
	}
	if _,err := url.ParseRequestURI(f.FileURL);err != nil {
		return errors.New("invalid file URL")
	}
	
	return nil
}

var pageRangeRegex = regexp.MustCompile(`^(\d+(-\d+)?)(,\d+(-\d+)?)*$`)

func ValidatePrintJobPayload(p domain.PrintJobPayload) error {
	if p.FileID.IsZero() {
		return errors.New("file_id is required and must be a valid ObjectID")
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

