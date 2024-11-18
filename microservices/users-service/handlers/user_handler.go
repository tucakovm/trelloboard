package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"io/ioutil"
	"net/http"
	"time"
	"users_module/models"
	"users_module/services"

	"github.com/golang-jwt/jwt/v4"
)

type UserHandler struct {
	service services.UserService
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

func (h UserHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}
	password, _ := HashPassword(req.Password)
	if err != nil {
		http.Error(w, `{"error": "Failed to hash password"}`, http.StatusInternalServerError)
		return
	}

	err = h.service.RegisterUser(req.FirstName, req.LastName, req.Username, req.Email, password, req.Role)
	if err != nil {
		http.Error(w, `{"error": "Registration failed"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Verification email sent"})
}

func (h UserHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
	}

	// Pokušaj dohvatanja korisnika
	user, err := h.service.GetUserByUsername(req.Username)
	if err != nil {
		http.Error(w, `{"error": "User not found"}`, http.StatusUnauthorized)
		return
	}

	if !CheckPassword(user.Password, req.Password) {
		http.Error(w, `{"error": "Invalid username or password"}`, http.StatusInternalServerError)
		return
	}

	if user.IsActive == false { // assuming Password is an exported field
		http.Error(w, `{"error": "User is not active"}`, http.StatusUnauthorized)
		return
	}

	token, err := GenerateJWT(user)
	if err != nil {
		http.Error(w, `{"error": "Error generating token"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Login successful",
		"token":   token,
	})
}

func (h UserHandler) VerifyHandler(w http.ResponseWriter, r *http.Request) {
	var req VerifyRequest

	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	err = h.service.VerifyAndActivateUser(req.Username, req.Code)
	if err != nil {
		http.Error(w, `{"error": "Verification or activation failed"}`, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User verified and saved successfully"})
}

func (h UserHandler) GetUserByUsername(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	user, err := h.service.GetUserByUsername(username)
	if err != nil {
		http.Error(w, `{"Bad request"}`, http.StatusBadRequest)
		return
	}

	userJson, err := json.Marshal(user)
	if err != nil {
		http.Error(w, `{"error": "Decoding error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(userJson)
}

func (h UserHandler) DeleteUserByUsername(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS , DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	err := h.service.DeleteUserByUsername(username)
	if err != nil {
		http.Error(w, `{"Bad request"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
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

func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS , DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read the request body", http.StatusBadRequest)
		return
	}

	var request ChangePasswordRequest
	if err := json.Unmarshal(body, &request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.service.ChangePassword(request.Username, request.CurrentPassword, request.NewPassword)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error changing password: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "Password updated successfully"}`))
}
