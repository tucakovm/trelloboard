package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// User struct represents a user in the database
type User struct {
	Id        primitive.ObjectID `bson:"_id,omitempty" json:"id"` // MongoDB's unique identifier for each document
	FirstName string             `bson:"firstname" json:"firstname"`
	LastName  string             `bson:"lastname" json:"lastname"`
	Username  string             `bson:"username" json:"username"`
	Email     string             `bson:"email" json:"email"`
	Password  string             `bson:"password" json:"password"`
	IsActive  bool               `bson:"is_active" json:"is_active"`
	Code      string             `bson:"code" json:"code"`
	Role      string             `bson:"role" json:"role"`
}
