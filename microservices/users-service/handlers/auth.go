package handlers

import (
	"encoding/json"
	"net/http"
	"users_module/models"
	"users_module/repositories"
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

	err = services.RegisterUser(req.FirstName, req.LastName, req.Username, req.Email)
	if err != nil {
		http.Error(w, `{"error": "Registration failed"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Verification email sent"})
}

func VerifyHandler(w http.ResponseWriter, r *http.Request) {
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

	err = services.VerifyUser(req.Email, req.Code)
	if err != nil {
		http.Error(w, `{"error": "Verification failed"}`, http.StatusBadRequest)
		return
	}

	user, err := repositories.GetUserByEmail(req.Email)
	if err != nil {
		user = &models.User{
			Email:    req.Email,
			IsActive: true,
		}
		err = repositories.SaveUser(*user)
		if err != nil {
			http.Error(w, `{"error": "Failed to save user"}`, http.StatusInternalServerError)
			return
		}
	} else {
		err = repositories.ActivateUser(req.Email)
		if err != nil {
			http.Error(w, `{"error": "Failed to activate user"}`, http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User verified and saved successfully"})
}
