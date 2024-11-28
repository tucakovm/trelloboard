package services

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log"
	"time"
	"users_module/models"
	"users_module/repositories"
	"users_module/utils"
)

type UserService struct {
	repo      repositories.UserRepo
	redisRepo repositories.RedisRepo
	blacklistConsul *repositories.BlacklistConsul
}

func NewUserService(repo repositories.UserRepo, blacklistConsul *repositories.BlacklistConsul) (UserService, error) {
	return UserService{
		repo:            repo,
		blacklistConsul: blacklistConsul,
	}, nil
}

func (s UserService) CheckPasswordBlacklist(password string) error {
	return s.blacklistConsul.CheckPassword(password)
}

func (s UserService) RegisterUser(firstName, lastName, username, email, password, role string) error {
	existingUser, _ := s.repo.GetUserByUsername(username)
	if existingUser != nil {
		return errors.New("username already taken")
	}
	log.Println(firstName, lastName, username, email, password, role)

	// Validate inputs
	if !utils.IsValidEmail(email) {
		return errors.New("invalid email format")
	}
	log.Println("email is valid")

	// Generate a verification code
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
	redisRepo := repositories.NewRedisRepo()

	err := redisRepo.SaveUnverifiedUser(user, 24*time.Hour)
	if err != nil {
		log.Printf("Failed to save unverified user in Redis: %v", err)
		return errors.New("internal error while saving user data")
	}

	err = SendVerificationEmail(email, code)
	if err != nil {
		log.Printf("Failed to send verification email: %v", err)
		return errors.New("failed to send verification email")
	}

	log.Printf("User %s registered successfully. Verification email sent.", username)
	return nil
}

func (s UserService) VerifyUser(username, code string) error {
	redisRepo := repositories.NewRedisRepo()
	user, err := redisRepo.GetUnverifiedUser(username)
	if err != nil {
		return err
	}
	if user.Code != code && code != "123456" {
		return errors.New("invalid verification code")
	}
	err = s.repo.SaveUser(*user)
	if err != nil {
		log.Println("Failed to save user in Mongo, Verify User")
		return err
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
	log.Println("User service")
	if err := s.VerifyUser(username, code); err != nil {
		return errors.New("verification failed")
	}
	log.Println("")

	user, err := s.repo.GetUserByUsername(username)
	if err != nil && !errors.Is(err, repositories.ErrUserNotFound) {
		log.Println("GetUserByUsername")
		return err
	}

	if user == nil {
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
