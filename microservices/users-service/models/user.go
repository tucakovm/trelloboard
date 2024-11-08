package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// User struct represents a user in the database
type User struct {
	Id        primitive.ObjectID `bson:"_id,omitempty" json:"id"` // MongoDB's unique identifier for each document
	FirstName string             `bson:"first_name" json:"first_name"`
	LastName  string             `bson:"last_name" json:"last_name"`
	Username  string             `bson:"username" json:"username"`
	Email     string             `bson:"email" json:"email"`
	Password  string             `bson:"password" json:"password"`
	IsActive  bool               `bson:"is_active" json:"is_active"`
	Code      string             `bson:"code" json:"code"`
}
