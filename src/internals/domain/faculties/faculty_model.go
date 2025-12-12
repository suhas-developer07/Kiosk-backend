package faculties

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Faculty struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Username           string             `bson:"username" json:"username"`
	Email              string             `bson:"email" json:"email"`
	Password           string             `bson:"password,omitempty" json:"password,omitempty"`
	GoogleID           string             `bson:"google_id,omitempty" json:"google_id,omitempty"`
	Profile            FacultyProfile     `bson:"profile,omitempty" json:"profile,omitempty"`
	IsProfileCompleted bool               `bson:"is_profile_completed" json:"is_profile_completed"`
	CreatedAt          time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt          time.Time          `bson:"updated_at" json:"updated_at"`
}

type FacultyProfile struct {
	FacultyID     string    `bson:"faculty_id,omitemty" json:"faculty_id"`
	Subjects      []Subject `bson:"subjects,omitempty" json:"subjects,omitempty"`
	Gender        string    `bson:"gender,omitempty" json:"gender,omitempty"`
	Qualification string    `bson:"qualification,omitempty" json:"qualification,omitempty"`
	Experience    int       `bson:"experience,omitempty" json:"experience,omitempty"`
	PhoneNumber   string    `bson:"phone_number,omitempty" json:"phone_number,omitempty"`
}

type Subject struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	SubjectCode string             `bson:"subject_code" json:"subject_code"`
	SubjectName string             `bson:"subject_name" json:"subject_name"`
}

type AccoutCreationPayload struct {
	Name     string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type SigninPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}


type UpdateProfilePayload struct {
	Subjects      []Subject `json:"subjects" validate:"required,dive"`
	Gender        string    `json:"gender" validate:"required,oneof=male female other"`
	Qualification string    `json:"qualification" validate:"required"`
	Experience    int       `json:"experience" validate:"required,min=0,max=50"`
	PhoneNumber   string    `json:"phone_number" validate:"required,e164"`
}
