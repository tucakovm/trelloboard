package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Task struct {
	Id          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Status      Status             `bson:"status" json:"status"`
	ProjectID   string             `bson:"project_id" json:"project_id"`
	Members     []User             `bson:"members" json:"members"`
}

type Tasks []*Task
