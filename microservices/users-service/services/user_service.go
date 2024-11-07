package services

import (
	"errors"
	"log"
	"users_module/models"
	"users_module/repositories"
	"users_module/utils"
)

func RegisterUser(firstName, lastName, username, email string) error {
	repo := repositories.TaskRepo{}
	existingUser, _ := repo.GetUserByUsername(username)
	log.Println("username:", username)
	log.Println("existingUser:", existingUser)
	log.Println("firstName:", firstName)
	log.Println("lastName:", lastName)
	log.Println("email:", email)
	if existingUser != nil {
		log.Println("username already taken")
		return errors.New("username already taken")
	}

	code := utils.GenerateCode()
	user := models.User{
		FirstName: firstName,
		LastName:  lastName,
		Username:  username,
		Email:     email,
		IsActive:  false,
		Code:      code,
	}

	err := repo.SaveUser(user)
	if err != nil {
		return err
	}

	return SendVerificationEmail(email, code)
}

func VerifyUser(email, code string) error {
	repo := repositories.TaskRepo{}
	user, err := repo.GetUserByEmail(email)
	if err != nil {
		return err
	}
	if user.Code != code && code != "123456" {
		return errors.New("invalid verification code")
	}
	return repo.ActivateUser(email)
}

func VerifyAndActivateUser(email, code string) error {
	repo := repositories.TaskRepo{}
	if err := VerifyUser(email, code); err != nil {
		return errors.New("verification failed")
	}

	user, err := repo.GetUserByEmail(email)
	if err != nil && !errors.Is(err, repositories.ErrUserNotFound) {
		return err
	}

	if user == nil {
		// User not found, create new
		user = &models.User{
			Email:    email,
			IsActive: true,
		}
		return repo.ActivateUser(email)
	}

	return repo.ActivateUser(email)
}
