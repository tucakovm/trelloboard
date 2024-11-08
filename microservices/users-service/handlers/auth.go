package handlers

import (
	"encoding/json"
	"fmt"
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
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type VerifyRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
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

	err = h.service.RegisterUser(req.FirstName, req.LastName, req.Username, req.Email, req.Password)
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

	// Provera lozinke
	if user.Password != req.Password { // assuming Password is an exported field
		http.Error(w, `{"error": "Invalid username or password"}`, http.StatusUnauthorized)
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

	err = h.service.VerifyAndActivateUser(req.Email, req.Code)
	if err != nil {
		http.Error(w, `{"error": "Verification or activation failed"}`, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User verified and saved successfully"})
}

func GenerateJWT(user *models.User) (string, error) {
	var secretKey = []byte("tajni_kljuc")
	// Kreiraj claims (podatke koji se šalju u tokenu)
	claims := jwt.MapClaims{
		"user_role": user.Role,
		"username":  user.Username,
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
