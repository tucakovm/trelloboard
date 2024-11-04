package handlers

import (
	"encoding/json"
	"log"
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

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	json.NewDecoder(r.Body).Decode(&req)
	err := services.RegisterUser(req.FirstName, req.LastName, req.Username, req.Email)
	log.Println("username:", req.Username)
	log.Println("firstName:", req.FirstName)
	log.Println("lastName:", req.LastName)
	log.Println("email:", req.Email)
	if err != nil {
		http.Error(w, "Registration failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Verification email sent"))
}

func VerifyHandler(w http.ResponseWriter, r *http.Request) {
	var req VerifyRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	err = services.VerifyUser(req.Email, req.Code)
	if err != nil {
		http.Error(w, "Verification failed", http.StatusBadRequest)
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
			http.Error(w, "Failed to save user", http.StatusInternalServerError)
			return
		}
	} else {
		err = repositories.ActivateUser(req.Email)
		if err != nil {
			http.Error(w, "Failed to activate user", http.StatusInternalServerError)
			return
		}
	}

	// Send a success response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User verified and saved successfully"))
}
