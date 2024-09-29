package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gitnoober/chat-go/service"
)

func createUser(w http.ResponseWriter, r *http.Request, svc *service.Service) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var user service.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Call the service to create the user
	if err := svc.CreateUser(user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User created successfully"})

}

func getUser(w http.ResponseWriter, r *http.Request, svc *service.Service) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := r.URL.Query().Get("id")

	if userID == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	// Get user from service
	user, err := svc.GetUserByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)

}

func handleUser(w http.ResponseWriter, r *http.Request, svc *service.Service) {
	switch requestType := r.Method; requestType {
	case http.MethodPost:
		log.Print("Inside POST method")
		createUser(w, r, svc)
	case http.MethodGet:
		log.Print("Inside GET method")
		getUser(w, r, svc)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
