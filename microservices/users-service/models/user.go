package models

type User struct {
	ID        int
	FirstName string
	LastName  string
	Username  string
	Email     string
	IsActive  bool
	Code      string
}
