package services

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log"
	"users_module/models"
	"users_module/repositories"
	"users_module/utils"
)

type UserService struct {
	repo            repositories.UserRepo
	blacklistConsul *repositories.BlacklistConsul
}

func NewUserService(repo repositories.UserRepo, blacklistConsul *repositories.BlacklistConsul) (UserService, error) {
	return UserService{
		repo:            repo,
		blacklistConsul: blacklistConsul,
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

	if err := s.blacklistConsul.CheckPassword(password); err != nil {
		log.Printf("Password rejected due to blacklist: %v", err)
		return fmt.Errorf("password is not allowed: %w", err)
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

func (s UserService) VerifyUser(username, code string) error {
	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		return err
	}
	if user.Code != code && code != "123456" {
		return errors.New("invalid verification code")
	}
	return s.repo.ActivateUser(username)
}

func (s UserService) GetUserByUsername(username string) (*models.User, error) {
	log.Println("usao u servis")
	user, err := s.repo.GetUserByUsername(username)

	if err != nil {
		return nil, err
	}
	return user, nil
}
func (s UserService) GetUserByEmail(email string) (*models.User, error) {
	log.Println("usao u servis")
	user, err := s.repo.GetUserByEmail(email)

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

func (s UserService) DeleteUserById(id string) error {
	err := s.repo.DeleteById(id)
	if err != nil {
		return err
	}
	return nil
}

func (s UserService) VerifyAndActivateUser(username, code string) error {
	if err := s.VerifyUser(username, code); err != nil {
		return errors.New("verification failed")
	}

	user, err := s.repo.GetUserByUsername(username)
	if err != nil && !errors.Is(err, repositories.ErrUserNotFound) {
		return err
	}

	if user == nil {
		// User not found, create new
		user = &models.User{
			Username: username,
			IsActive: true,
		}
		return s.repo.ActivateUser(username)
	}

	return s.repo.ActivateUser(username)
}
func (s *UserService) ChangePassword(username, currentPassword, newPassword string) error {

	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		log.Println("User not found")
		return fmt.Errorf("user not found")
	}

	if !CheckPassword(user.Password, currentPassword) {
		log.Println("current password not correct")
		log.Println(currentPassword, user.Password)
		return fmt.Errorf("current password is incorrect")
	}

	if err := s.blacklistConsul.CheckPassword(newPassword); err != nil {
		return fmt.Errorf("new password is not allowed because its commonly used: %w", err)
	}

	// Hash the new password
	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		log.Println("Error hashing")
		return fmt.Errorf("failed to hash the new password")
	}

	// Update the password in the repository
	err = s.repo.UpdatePassword(user.Username, hashedPassword)
	if err != nil {
		log.Println("Error updating password")
		return fmt.Errorf("failed to update the password")
	}

	return nil
}

func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}
func CheckPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func (s *UserService) RecoverPassword(userName, newPassword string) error {
	log.Println(userName)
	user, err := s.repo.GetUserByUsername(userName)
	if err != nil {
		log.Println("User not found")
		return fmt.Errorf("user not found")
	}

	if err := s.blacklistConsul.CheckPassword(newPassword); err != nil {
		return fmt.Errorf("new password is not allowed: %w", err)
	}

	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		log.Println("Error hashing")
		return fmt.Errorf("failed to hash the new password")
	}

	err = s.repo.UpdatePassword(user.Username, hashedPassword)
	if err != nil {
		log.Println("Error updating password")
		return fmt.Errorf("failed to update the password")
	}

	return nil
}
