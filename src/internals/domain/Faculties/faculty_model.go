package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Faculty struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    FacultyID string             `bson:"faculty_id" json:"faculty_id"`
    Email     string             `bson:"email" json:"email"`
    Password  string             `bson:"password" json:"password"`
    GoogleID  string             `bson:"google_id,omitempty" json:"google_id,omitempty"`
}
