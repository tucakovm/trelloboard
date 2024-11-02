package repositories

import (
	"errors"
	"users_module/models"
)

var users = []models.User{}

func SaveUser(user models.User) error {
	users = append(users, user)
	return nil
}

func GetUserByEmail(email string) (*models.User, error) {
	for _, user := range users {
		if user.Email == email {
			return &user, nil
		}
	}
	return nil, errors.New("user not found")
}

func ActivateUser(email string) error {
	for i, user := range users {
		if user.Email == email {
			users[i].IsActive = true
			return nil
		}
	}
	return errors.New("user not found")
}
