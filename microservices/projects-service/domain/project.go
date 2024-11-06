package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Project struct {
	Id             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name           string             `bson:"name" json:"name"`
	CompletionDate time.Time          `bson:"completionDate" json:"completionDate"`
	MinMembers     int32              `bson:"minMembers" json:"minMembers"`
	MaxMembers     int32              `bson:"maxMembers" json:"maxMembers"`
	//Manager        User
	//Members        []User
}

type Projects []*Project
