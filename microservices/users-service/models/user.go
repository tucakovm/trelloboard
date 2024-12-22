package models

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

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

type ErrRespTmp struct {
	URL        string
	Method     string
	StatusCode int
}

func (e ErrRespTmp) Error() string {
	return fmt.Sprintf("temporary error [status code %d] for request: HTTP %s\t%s", e.StatusCode, e.Method, e.URL)
}

type ErrResp struct {
	URL        string
	Method     string
	StatusCode int
}

func (e ErrResp) Error() string {
	return fmt.Sprintf("error [status code %d] for request: HTTP %s\t%s", e.StatusCode, e.Method, e.URL)
}

type ErrCircuitBreakerOpen struct {
	Message string
}

func (e ErrCircuitBreakerOpen) Error() string {
	return e.Message
}
