package Files

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type File struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Title        string             `bson:"title" json:"title"`
	Description  string             `bson:"description" json:"description"`
	FileURL      string             `bson:"file_url" json:"file_url"`
	Grade        string             `bson:"grade" json:"grade"`
	Subject      string             `bson:"subject" json:"subject"`
	Category     string             `bson:"category" json:"category"`
	FacultyID    primitive.ObjectID `bson:"faculty_id" json:"faculty_id"`
	GroupAllowed string             `bson:"group_allowed" json:"group_allowed"`
	FileType     string             `bson:"file_type" json:"file_type"`
	UploadedAt   primitive.DateTime `bson:"uploaded_at" json:"uploaded_at"`
}

type FileUploadRequest struct {
	Title        string             `json:"file_name"`
	Description  string             `json:"description"`
	Grade        string             `json:"grade"`
	Subject      string             `json:"subject"`
	Category     string             `json:"category"`
	FacultyID    primitive.ObjectID `json:"faculty_id"`
	GroupAllowed string             `json:"group_allowed"`
	FileType     string             `json:"file_type"`
}

type PrintJob struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	FileID              primitive.ObjectID `bson:"file_id" json:"file_id"`
	Copies              int                `bson:"copies" json:"copies"`
	PrintingSide        string             `bson:"printing_side" json:"printing_side"`
	PrintingMode        string             `bson:"printing_mode" json:"printing_mode"`
	PageRange           string             `bson:"page_range" json:"page_range"`
	PageLayout          string             `bson:"paper_size" json:"PageLayout"`
	Price               int                `bson:"price" json:"price"`
	OrderStatus         string             `bson:"order_status" json:"order_status"`
	Token               string             `bson:"token" json:"token"`
	TotalSheetsRequired int                `bson:"total_sheets_required" json:"total_sheets_required"`
	CreatedAt           time.Time          `bson:"created_at" json:"created_at"`
}
type PrintJobPayload struct {
	FileID       primitive.ObjectID `json:"file_id" validate:"required"`
	FileName     string             `json:"file_name" validate:"required,min=3"`
	Copies       int                `json:"copies" validate:"required,min=1,max=100"`
	PrintingSide string             `json:"printing_side" validate:"required,oneof=single double"`
	PrintingMode string             `json:"printing_mode" validate:"required,oneof=color bw"`
	PageRange    string             `json:"page_range" validate:"omitempty"`
	PageLayout   string             `json:"PageLayout" validate:"required,oneof=2-up 4-up 1-up"`
}
