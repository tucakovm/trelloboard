package handlers

import (
	"encoding/json"
	"net/http"
	"users_module/services"
)

type RegisterRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Email     string `json:"email"`
}

type VerifyRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	json.NewDecoder(r.Body).Decode(&req)
	err := services.RegisterUser(req.FirstName, req.LastName, req.Username, req.Email)
	if err != nil {
		http.Error(w, "Registration failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Verification email sent"))
}

func VerifyHandler(w http.ResponseWriter, r *http.Request) {
	var req VerifyRequest
	json.NewDecoder(r.Body).Decode(&req)
	err := services.VerifyUser(req.Email, req.Code)
	if err != nil {
		http.Error(w, "Verification failed", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User verified successfully"))
}
