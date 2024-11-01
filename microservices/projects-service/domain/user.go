package domain

import (
	"github.com/google/uuid"
)

type User struct {
	Id        uuid.UUID
	Username  string
	Password  string
	Firstname string
	Lastname  string
	Email     string
}
