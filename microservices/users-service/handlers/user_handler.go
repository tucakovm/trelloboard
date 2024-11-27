package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	otelCodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	"users_module/models"
	proto "users_module/proto/users"
	"users_module/services"
)

type UserHandler struct {
	service        services.UserService
	projectService proto.ProjectServiceClient
	proto.UnimplementedUsersServiceServer
	Tracer trace.Tracer
}

func NewUserHandler(service services.UserService, projectService proto.ProjectServiceClient, tracer trace.Tracer) (UserHandler, error) {
	return UserHandler{
		service:        service,
		projectService: projectService,
		Tracer:         tracer,
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
	ctx, span := h.Tracer.Start(ctx, "h.register")
	defer span.End()
	captchaValid, err := h.verifyCaptcha(req.User.CaptchaResponse)
	if err != nil || !captchaValid {
		err := errors.New("bad request ...")
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.InvalidArgument, "Invalid or failed CAPTCHA verification")
	}

	user := req.User
	password, _ := HashPassword(user.Password)

	err = h.service.RegisterUser(user.Firstname, user.Lastname, user.Username, user.Email, password, user.Role, ctx)
	if err != nil {
		err := errors.New("bad request ...")
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.InvalidArgument, "bad request ...")
	}

	return nil, nil
}

func (h UserHandler) LoginUserHandler(ctx context.Context, req *proto.LoginReq) (*proto.LoginRes, error) {
	ctx, span := h.Tracer.Start(ctx, "h.login")
	defer span.End()
	// Pokušaj dohvatanja korisnika

	log.Println("Usao u handler login")

	captchaValid, err := h.verifyCaptcha(req.LoginUser.Key)
	if err != nil || !captchaValid {
		err := errors.New("bad request ...")
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.InvalidArgument, "Invalid or failed CAPTCHA verification")
	}

	user, err := h.service.GetUserByUsername(req.LoginUser.Username, ctx)
	if err != nil {
		err := errors.New("bad request ...")
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.InvalidArgument, "User not found ...")
	}

	// Provera lozinke
	if !CheckPassword(user.Password, req.LoginUser.Password) {
		err := errors.New("bad request ...")
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.Unauthenticated, "Invalid username or password ...")
	}

	// Provera da li je korisnik aktivan
	if !user.IsActive {
		err := errors.New("bad request ...")
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.PermissionDenied, "User is not active ...")
	}

	// Generisanje JWT tokena
	token, err := GenerateJWT(user)
	if err != nil {
		err := errors.New("bad request ...")
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.Internal, "Error generating token ...")
	}

	// Vraćanje odgovora sa tokenom
	return &proto.LoginRes{
		Message: "Login successful",
		Token:   token,
	}, nil
}

func (h UserHandler) VerifyHandler(ctx context.Context, req *proto.VerifyReq) (*proto.EmptyResponse, error) {
	ctx, span := h.Tracer.Start(ctx, "h.verify")
	defer span.End()

	err := h.service.VerifyAndActivateUser(req.VerifyUser.Username, req.VerifyUser.Code, ctx)
	if err != nil {
		err := errors.New("bad request ...")
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.InvalidArgument, "bad request ...")
	}

	return nil, nil
}

func (h UserHandler) GetUserByUsername(ctx context.Context, req *proto.GetUserByUsernameReq) (*proto.GetUserByUsernameRes, error) {

	ctx, span := h.Tracer.Start(ctx, "h.getUserByUsername")
	defer span.End()

	user, err := h.service.GetUserByUsername(req.Username, ctx)
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
	ctx, span := h.Tracer.Start(ctx, "h.DeleteUserByUsername")
	defer span.End()
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}
	authHeader := md.Get("authorization")
	if len(authHeader) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing authorization header")
	}
	tokenString := strings.TrimPrefix(authHeader[0], "Bearer ")
	if tokenString == "" {
		return nil, status.Error(codes.Unauthenticated, "invalid token format")
	}
	claims, err := parseJWT(tokenString)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
	}
	role, ok := claims["user_role"].(string)
	if !ok || role == "" {
		return nil, status.Error(codes.Unauthenticated, "role not found in token")
	}

	userOnProjectReq := &proto.UserOnProjectReq{
		Id:   req.Username,
		Role: role,
	}
	projServiceResponse, err := h.projectService.UserOnProject(ctx, userOnProjectReq)
	if err != nil {
		err := errors.New("Error checking project")
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.Internal, "Error checking project")
	}
	if projServiceResponse.OnProject {

		err := errors.New("User is assigned to a project.")
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.Internal, "User is assigned to a project.")
	}
	err = h.service.DeleteUserById(req.Username, ctx)
	if err != nil {
		err := errors.New("Bad request.")
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.InvalidArgument, "Bad request.")
	}
	return nil, nil
}

func GenerateJWT(user *models.User) (string, error) {
	var secretKey = []byte(os.Getenv("JWT_SECRET_KEY"))
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
	ctx, span := h.Tracer.Start(ctx, "h.changePassword")
	defer span.End()
	err := h.service.ChangePassword(req.ChangeUser.Username, req.ChangeUser.CurrentPassword, req.ChangeUser.NewPassword, ctx)
	if err != nil {
		err := errors.New("bad request")
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.InvalidArgument, "bad request ...")
	}

	return nil, nil
}

func (h UserHandler) verifyCaptcha(captchaResponse string) (bool, error) {

	secretKey := os.Getenv("CAPTCHA_SECRET_KEY")

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

func parseJWT(tokenString string) (jwt.MapClaims, error) {
	secret := []byte(os.Getenv("JWR_SECRET_KEY"))
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

func (h UserHandler) MagicLink(ctx context.Context, req *proto.MagicLinkReq) (*proto.EmptyResponse, error) {
	ctx, span := h.Tracer.Start(ctx, "h.magicLink")
	defer span.End()
	user, err := h.service.GetUserByEmail(req.MagicLink.Email, ctx)
	if err != nil {
		err := errors.New("user not found")
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.InvalidArgument, "User not found")
	}

	token, err := GenerateJWT(user)
	if err != nil {
		err := errors.New("err gen token")
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.Internal, "Error generating token")
	}

	frontendURL := "http://localhost:4200"
	magicLink := fmt.Sprintf("%s/magic-login?token=%s", frontendURL, token)

	err = services.SendMagicLinkEmail(user.Email, magicLink)
	if err != nil {
		err := errors.New("error sending email")
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.Internal, "Error sending email")
	}

	return &proto.EmptyResponse{}, nil
}

func (h UserHandler) RecoveryLink(ctx context.Context, req *proto.RecoveryLinkReq) (*proto.EmptyResponse, error) {
	ctx, span := h.Tracer.Start(ctx, "h.recoveryLink")
	defer span.End()
	user, err := h.service.GetUserByEmail(req.RecoveryLink.Email, ctx)
	if err != nil {
		err := errors.New("user not found")
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.NotFound, "User not found")
	}

	baseFrontendURL := "http://localhost:4200"

	// Generate the recovery URL
	recoveryURL := fmt.Sprintf("%s/change-password?username=%s&email=%s",
		baseFrontendURL,
		url.QueryEscape(user.Username),
		url.QueryEscape(user.Email),
	)

	// Prepare the subject and body for the email
	subject := "Password Recovery"
	body := fmt.Sprintf("Hi %s,\n\nClick the button below to recover your password:\n\n%s\n\nIf you did not request this, please ignore this email.",
		user.Username, recoveryURL)

	// Send the email
	err = services.SendEmail(user.Email, subject, body)
	if err != nil {
		err := errors.New("failed to send recovery email")
		span.SetStatus(otelCodes.Error, err.Error())
		return nil, status.Error(codes.Internal, "Failed to send recovery email")
	}

	return &proto.EmptyResponse{}, nil
}

func (h *UserHandler) RecoverPassword(ctx context.Context, req *proto.RecoveryPasswordRequest) (*proto.EmptyResponse, error) {
	ctx, span := h.Tracer.Start(ctx, "h.recoverPass")
	defer span.End()
	if req == nil {
		err := errors.New("invalid request payload")
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("RecoverPassword field is nil in request")
		return nil, errors.New("invalid request payload")
	}

	log.Printf("RecoverPassword request: username=%s, newPassword=%s", req.Username, req.NewPassword)

	log.Println("req.UserName")
	log.Println(req.Username)
	log.Println("req.NewPassword")
	log.Println(req.NewPassword)
	log.Println(req)
	err := h.service.RecoverPassword(req.Username, req.NewPassword, ctx)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())
		log.Println("Error in service:", err)
		return nil, err
	}

	return &proto.EmptyResponse{}, nil
}
