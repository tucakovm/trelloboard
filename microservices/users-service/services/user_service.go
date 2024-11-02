package services

import (
	"errors"
	"users_module/models"
	"users_module/repositories"
	"users_module/utils"
)

func RegisterUser(firstName, lastName, username, email string) error {
	code := utils.GenerateCode()
	user := models.User{
		FirstName: firstName,
		LastName:  lastName,
		Username:  username,
		Email:     email,
		IsActive:  false,
		Code:      code,
	}
	err := repositories.SaveUser(user)
	if err != nil {
		return err
	}
	return SendVerificationEmail(email, code)
}

func VerifyUser(email, code string) error {
	user, err := repositories.GetUserByEmail(email)
	if err != nil {
		return err
	}
	if user.Code != code {
		return errors.New("invalid verification code")
	}
	return repositories.ActivateUser(email)
}
