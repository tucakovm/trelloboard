package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Task struct {
	Id          primitive.ObjectID `json:"id,omitempty" bson:"id"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
	Status      Status             `json:"status" bson:"status"`
	ProjectID   string             `json:"project_id" bson:"project_id"`
}

type Tasks []*Task
