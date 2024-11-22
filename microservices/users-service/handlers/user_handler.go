package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
	"users_module/models"
	proto "users_module/proto/users"
	"users_module/services"
)

type UserHandler struct {
	service services.UserService
	proto.UnimplementedUsersServiceServer
}

func NewUserHandler(service services.UserService) (UserHandler, error) {
	return UserHandler{
		service: service,
	}, nil
}

type RegisterRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Role      string `json:"role"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type VerifyRequest struct {
	Username string `json:"username"`
	Code     string `json:"code"`
}

type ChangePasswordRequest struct {
	Username        string `json:"username" binding:"required"`
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required,min=6"`
}

func (h UserHandler) RegisterHandler(ctx context.Context, req *proto.RegisterReq) (*proto.EmptyResponse, error) {
	captchaValid, err := h.verifyCaptcha(req.User.Key)
	if err != nil || !captchaValid {
		return nil, status.Error(codes.InvalidArgument, "Invalid or failed CAPTCHA verification")
	}
	user := req.User
	password, _ := HashPassword(user.Password)

	err = h.service.RegisterUser(user.Firstname, user.Lastname, user.Username, user.Email, password, user.Role)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "bad request ...")
	}

	return nil, nil
}

func (h UserHandler) LoginUserHandler(ctx context.Context, req *proto.LoginReq) (*proto.LoginRes, error) {
	// Pokušaj dohvatanja korisnika

	log.Println("Usao u handler login")

	captchaValid, err := h.verifyCaptcha(req.LoginUser.Key)
	if err != nil || !captchaValid {
		return nil, status.Error(codes.InvalidArgument, "Invalid or failed CAPTCHA verification")
	}

	user, err := h.service.GetUserByUsername(req.LoginUser.Username)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "User not found ...")
	}

	// Provera lozinke
	if !CheckPassword(user.Password, req.LoginUser.Password) {
		return nil, status.Error(codes.Unauthenticated, "Invalid username or password ...")
	}

	// Provera da li je korisnik aktivan
	if !user.IsActive {
		return nil, status.Error(codes.PermissionDenied, "User is not active ...")
	}

	// Generisanje JWT tokena
	token, err := GenerateJWT(user)
	if err != nil {
		return nil, status.Error(codes.Internal, "Error generating token ...")
	}

	// Vraćanje odgovora sa tokenom
	return &proto.LoginRes{
		Message: "Login successful",
		Token:   token,
	}, nil
}

func (h UserHandler) VerifyHandler(ctx context.Context, req *proto.VerifyReq) (*proto.EmptyResponse, error) {

	err := h.service.VerifyAndActivateUser(req.VerifyUser.Username, req.VerifyUser.Code)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "bad request ...")
	}

	return nil, nil
}

func (h UserHandler) GetUserByUsername(ctx context.Context, req *proto.GetUserByUsernameReq) (*proto.GetUserByUsernameRes, error) {

	user, err := h.service.GetUserByUsername(req.Username)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "bad request ...")
	}
	protoUser := &proto.UserL{
		Id:        user.Id.Hex(), // Konverzija MongoDB ObjectID u string
		Firstname: user.FirstName,
		Lastname:  user.LastName,
		Username:  user.Username,
		Email:     user.Email,
		IsActive:  user.IsActive,
		Code:      user.Code,
		Role:      user.Role,
	}
	response := &proto.GetUserByUsernameRes{User: protoUser}
	return response, nil
}

func (h UserHandler) DeleteUserByUsername(ctx context.Context, req *proto.GetUserByUsernameReq) (*proto.EmptyResponse, error) {

	err := h.service.DeleteUserByUsername(req.Username)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "bad request ...")
	}
	return nil, nil
}

func GenerateJWT(user *models.User) (string, error) {
	var secretKey = []byte("matija_AFK")
	// Kreiraj claims (podatke koji se šalju u tokenu)
	claims := jwt.MapClaims{
		"user_role": user.Role,
		"username":  user.Username,
		"id":        user.Id,
		"exp":       time.Now().Add(time.Hour * 24).Unix(), // Tokom od 24 sata
	}

	// Kreiraj token sa HMAC algoritmom i našim claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Potpiši token sa tajnim ključem
	signedToken, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	fmt.Println(signedToken)
	return signedToken, nil
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

func (h *UserHandler) ChangePassword(ctx context.Context, req *proto.ChangePasswordReq) (*proto.EmptyResponse, error) {

	err := h.service.ChangePassword(req.ChangeUser.Username, req.ChangeUser.CurrentPassword, req.ChangeUser.NewPassword)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "bad request ...")
	}

	return nil, nil
}

func (h UserHandler) verifyCaptcha(captchaResponse string) (bool, error) {

	secretKey := os.Getenv("CAPTCHA_SECRET_KEY_TEST")

	verificationURL := "https://www.google.com/recaptcha/api/siteverify"

	// Request body
	data := fmt.Sprintf("secret=%s&response=%s", secretKey, captchaResponse)

	// Salje post za verifikaciju
	resp, err := http.Post(verificationURL, "application/x-www-form-urlencoded", bytes.NewBuffer([]byte(data)))
	if err != nil {
		return false, fmt.Errorf("failed to verify CAPTCHA: %v", err)
	}
	defer resp.Body.Close()

	// Cita response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read CAPTCHA verification response: %v", err)
	}

	// Parsira JSON
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return false, fmt.Errorf("failed to parse CAPTCHA verification response: %v", err)
	}

	// Provjera uspjeha
	if success, ok := response["success"].(bool); ok && success {
		return true, nil
	}
	return false, nil
}
