package faculties

import (
	"time"

	"github.com/suhas-developer07/Kiosk-backend/src/internals/domain/subjects"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// type Faculty struct {
// 	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
// 	Username           string             `bson:"username" json:"username"`
// 	Email              string             `bson:"email" json:"email"`
// 	Password           string             `bson:"password,omitempty" json:"password,omitempty"`
// 	GoogleID           string             `bson:"google_id,omitempty" json:"google_id,omitempty"`
// 	Profile            FacultyProfile     `bson:"profile,omitempty" json:"profile,omitempty"`
// 	IsProfileCompleted bool               `bson:"is_profile_completed" json:"is_profile_completed"`
// 	CreatedAt          time.Time          `bson:"created_at" json:"created_at"`
// 	UpdatedAt          time.Time          `bson:"updated_at" json:"updated_at"`
// }

// type FacultyProfile struct {
// 	FacultyID     string             `bson:"faculty_id,omitemty" json:"faculty_id"`
// 	Subjects      []subjects.Subject `bson:"subjects,omitempty" json:"subjects,omitempty"`
// 	Gender        string             `bson:"gender,omitempty" json:"gender,omitempty"`
// 	Qualification string             `bson:"qualification,omitempty" json:"qualification,omitempty"`
// 	Experience    int                `bson:"experience,omitempty" json:"experience,omitempty"`
// 	PhoneNumber   string             `bson:"phone_number,omitempty" json:"phone_number,omitempty"`
// }

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
	Subjects      []subjects.Subject `json:"subjects" validate:"required,dive"`
	Gender        string             `json:"gender" validate:"required,oneof=male female other"`
	Qualification string             `json:"qualification" validate:"required"`
	Experience    int                `json:"experience" validate:"required,min=0,max=50"`
	PhoneNumber   string             `json:"phone_number" validate:"required,e164"`
}

type Faculty struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Username      string             `bson:"username" json:"username"`
	Email         string             `bson:"email" json:"email"`
	Password      string             `bson:"password" json:"password"`
	Subjects      []subjects.Subject `bson:"subjects" json:"subjects"`
	Stream        string             `bson:"stream" json:"stream"`
	ClassHandling string             `bson:"class_handling" json:"class_handling"`
	PhoneNumber   string             `bson:"phone_number,omitempty" json:"phone_number,omitempty"`
	Gender        string             `bson:"gender" json:"gender"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}

type FacultyPayload struct {
	Username      string             `bson:"username" json:"username" validate:"required"`
	Email         string             `bson:"email" json:"email" validate:"required"`
	Password      string             `bson:"password" json:"password" validate:"required,min=6"`
	Subjects      []subjects.Subject `bson:"subjects" json:"subjects" validate:"required"`
	Stream        string             `bson:"stream" json:"stream" validate:"required,oneof=science commerce arts"`
	ClassHandling string             `bson:"class_handling" json:"class_handling"`
	PhoneNumber   string             `bson:"phone_number,omitempty" json:"phone_number,omitempty"`
	Gender        string             `bson:"gender" json:"gender" validate:"required,oneof=male female other"`
}

type FacultyUpdateRequest struct {
	Username      string             `json:"username"`
	Email         string             `json:"email"`
	Password      string             `json:"password"`
	Subjects      []subjects.Subject `json:"subjects"`
	Stream        string             `json:"stream"`
	ClassHandling string             `json:"class_handling"`
	PhoneNumber   string             `json:"phone_number"`
	Gender        string             `json:"gender"`
}