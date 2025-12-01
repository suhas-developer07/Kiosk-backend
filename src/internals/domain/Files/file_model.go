package Files

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type File struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	FileName     string             `bson:"file_name" json:"file_name"`
	FileURL      string             `bson:"file_url" json:"file_url"`
	Description  string             `bson:"description" json:"description"`
	Subject      string             `bson:"subject" json:"subject"`
	FacultyID    primitive.ObjectID `bson:"faculty_id" json:"faculty_id"`
	FacultyName  string             `bson:"faculty_name" json:"faculty_name"`
	GroupAllowed string             `bson:"group_allowed" json:"group_allowed"`
	Type         string             `bson:"type" json:"type"`
	Date         primitive.DateTime `bson:"date" json:"date"`
}

type FileUploadRequest struct {
	FileName     string             `json:"file_name"`
	Description  string             `json:"description"`
	Subject      string             `json:"subject"`
	FacultyID    primitive.ObjectID `json:"faculty_id"`
	GroupAllowed string             `json:"group_allowed"`
	Type         string             `json:"type"`
}

type PrintJob struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	FileID              primitive.ObjectID `bson:"file_id" json:"file_id"`
	Copies              int                `bson:"copies" json:"copies"`
	PrintingSide        string             `bson:"printing_side" json:"printing_side"`
	PrintingMode        string             `bson:"printing_mode" json:"printing_mode"`
	PageRange           string             `bson:"page_range" json:"page_range"`
	PaperSize           string             `bson:"paper_size" json:"paper_size"`
	Price               float64            `bson:"price" json:"price"`
	OrderStatus         string             `bson:"order_status" json:"order_status"`
	Token               string             `bson:"token" json:"token"`
	TotalSheetsRequired int                `bson:"total_sheets_required" json:"total_sheets_required"`
}
