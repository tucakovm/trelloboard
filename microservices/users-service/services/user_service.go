package services

import (
	"context"
	"errors"
	"fmt"
	otelCodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"
	"log"
	"time"
	"users_module/models"
	"users_module/repositories"
	"users_module/utils"
)

type UserService struct {
	repo repositories.UserRepo
	//redisRepo       repositories.RedisRepo
	blacklistConsul *repositories.BlacklistConsul
	Tracer          trace.Tracer
}

func NewUserService(repo repositories.UserRepo, blacklistConsul *repositories.BlacklistConsul, Tracer trace.Tracer) (UserService, error) {
	return UserService{
		repo: repo,
		//redisRepo:       redisRepo,
		blacklistConsul: blacklistConsul,
		Tracer:          Tracer,
	}, nil
}

func (s UserService) CheckPasswordBlacklist(password string) error {
	return s.blacklistConsul.CheckPassword(password)
}

/*func (s UserService) RegisterUser(firstName, lastName, username, email, password, role string) error {
	existingUser, _ := s.repo.GetUserByUsername(username)

	repo   repositories.UserRepo
	Tracer trace.Tracer
}*/

/*func NewUserService(repo repositories.UserRepo, tracer trace.Tracer) (UserService, error) {
	return UserService{
		repo:   repo,
		Tracer: tracer,
	}, nil
}*/

func (s UserService) RegisterUser(firstName, lastName, username, email, password, role string, ctx context.Context) error {
	ctx, span := s.Tracer.Start(ctx, "s.register")
	defer span.End()
	existingUser, _ := s.repo.GetUserByUsername(username, ctx)
	//log.Println("username:", username)
	//log.Println("existingUser:", existingUser)
	//log.Println("firstName:", firstName)
	//log.Println("lastName:", lastName)
	//log.Println("email:", email)
	log.Println("role:", role)
	if existingUser != nil {
		err := errors.New("username already taken")
		span.SetStatus(otelCodes.Error, err.Error())
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

func (s UserService) VerifyUser(username, code string, ctx context.Context) error {
	ctx, span := s.Tracer.Start(ctx, "s.verify")
	defer span.End()
	redisRepo := repositories.NewRedisRepo()
	user, err := redisRepo.GetUnverifiedUser(username)

	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return err
	}
	if user.Code != code {
		return errors.New("invalid verification code")
	}

	err = s.repo.SaveUser(*user, ctx)
	if err != nil {
		log.Println("Failed to save user in Mongo, Verify User")
		return err
	}

	return s.repo.ActivateUser(username, ctx)
}

func (s UserService) GetUserByUsername(username string, ctx context.Context) (*models.User, error) {
	ctx, span := s.Tracer.Start(ctx, "s.getUserByUsername")
	defer span.End()
	log.Println("usao u servis")
	user, err := s.repo.GetUserByUsername(username, ctx)

	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, err
	}
	return user, nil
}
func (s UserService) GetUserByEmail(email string, ctx context.Context) (*models.User, error) {
	ctx, span := s.Tracer.Start(ctx, "s.getUserByEmail")
	defer span.End()
	log.Println("usao u servis")
	user, err := s.repo.GetUserByEmail(email, ctx)

	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, err
	}
	return user, nil
}

func (s UserService) DeleteUserByUsername(username string, ctx context.Context) error {
	ctx, span := s.Tracer.Start(ctx, "s.deleteUserByUsername")
	defer span.End()
	err := s.repo.Delete(username, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return err
	}
	return nil
}

func (s UserService) DeleteUserById(id string, ctx context.Context) error {
	ctx, span := s.Tracer.Start(ctx, "s.deleteUserById")
	defer span.End()
	err := s.repo.DeleteById(id, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		return err
	}
	return nil
}

func (s UserService) VerifyAndActivateUser(username, code string, ctx context.Context) error {
	ctx, span := s.Tracer.Start(ctx, "s.VerifyAndActivateUser")
	defer span.End()
	if err := s.VerifyUser(username, code, ctx); err != nil {

		return errors.New("verification failed")
	}
	log.Println("")

	user, err := s.repo.GetUserByUsername(username, ctx)
	if err != nil && !errors.Is(err, repositories.ErrUserNotFound) {
		log.Println("GetUserByUsername")

		span.SetStatus(otelCodes.Error, err.Error())
		return err
	}

	if user == nil {
		user = &models.User{
			Username: username,
			IsActive: true,
		}
		return s.repo.ActivateUser(username, ctx)
	}

	return s.repo.ActivateUser(username, ctx)
}
func (s *UserService) ChangePassword(username, currentPassword, newPassword string, ctx context.Context) error {
	ctx, span := s.Tracer.Start(ctx, "s.changePass")
	defer span.End()
	user, err := s.repo.GetUserByUsername(username, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
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
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("Error hashing")
		return fmt.Errorf("failed to hash the new password")
	}

	// Update the password in the repository
	err = s.repo.UpdatePassword(user.Username, hashedPassword, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
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

func (s *UserService) RecoverPassword(userName, newPassword string, ctx context.Context) error {
	ctx, span := s.Tracer.Start(ctx, "s.recoverPass")
	defer span.End()
	log.Println(userName)
	user, err := s.repo.GetUserByUsername(userName, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("User not found")
		return fmt.Errorf("user not found")
	}

	if err := s.blacklistConsul.CheckPassword(newPassword); err != nil {
		return fmt.Errorf("new password is not allowed: %w", err)
	}

	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("Error hashing")
		return fmt.Errorf("failed to hash the new password")
	}

	err = s.repo.UpdatePassword(user.Username, hashedPassword, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("Error updating password")
		return fmt.Errorf("failed to update the password")
	}

	return nil
}
