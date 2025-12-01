package utils

import (
	"errors"
	"net/url"
	"strings"

	models "github.com/suhas-developer07/Kiosk-backend/src/internals/Models"
	// "go.mongodb.org/mongo-driver/bson/primitive"
)

func ValidateFileInput(f *models.File) error{

	if strings.TrimSpace(f.FileName) == ""{
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