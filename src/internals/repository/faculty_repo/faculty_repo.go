package facultyrepo

import "go.mongodb.org/mongo-driver/mongo"

type FacultyRepo struct {
	client *mongo.Client
}

func NewFacultyRepo(db *mongo.Database,client *mongo.Client)*FacultyRepo{
	return &FacultyRepo{
		client: client,
	}
}


