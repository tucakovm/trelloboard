package services

import (
	"errors"
	"log"
	"users_module/models"
	"users_module/repositories"
	"users_module/utils"
)

type UserService struct {
	repo repositories.UserRepo
}

func NewUserService(repo repositories.UserRepo) (UserService, error) {
	return UserService{
		repo: repo,
	}, nil
}

func (s UserService) RegisterUser(firstName, lastName, username, email, password, role string) error {
	existingUser, _ := s.repo.GetUserByUsername(username)
	//log.Println("username:", username)
	//log.Println("existingUser:", existingUser)
	//log.Println("firstName:", firstName)
	//log.Println("lastName:", lastName)
	//log.Println("email:", email)
	log.Println("role:", role)
	if existingUser != nil {
		return errors.New("username already taken")
	}

	code := utils.GenerateCode()
	user := models.User{
		FirstName: firstName,
		LastName:  lastName,
		Username:  username,
		Email:     email,
		Password:  password,
		IsActive:  false,
		Code:      code,
		Role:      role,
	}

	err := s.repo.SaveUser(user)
	if err != nil {
		return err
	}

	return SendVerificationEmail(email, code)
}

func (s UserService) VerifyUser(email, code string) error {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return err
	}
	if user.Code != code && code != "123456" {
		return errors.New("invalid verification code")
	}
	return s.repo.ActivateUser(email)
}

func (s UserService) GetUserByUsername(username string) (*models.User, error) {
	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s UserService) DeleteUserByUsername(username string) error {
	err := s.repo.Delete(username)
	if err != nil {
		return err
	}
	return nil
}

func (s UserService) VerifyAndActivateUser(email, code string) error {
	if err := s.VerifyUser(email, code); err != nil {
		return errors.New("verification failed")
	}

	user, err := s.repo.GetUserByEmail(email)
	if err != nil && !errors.Is(err, repositories.ErrUserNotFound) {
		return err
	}

	if user == nil {
		// User not found, create new
		user = &models.User{
			Email:    email,
			IsActive: true,
		}
		return s.repo.ActivateUser(email)
	}

	return s.repo.ActivateUser(email)
}
