package main

import (
	"log"
	"net/http"
	"github.com/gitnoober/chat-go/service"
)

type User struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	Name       string `json:"name"`
	ProfileURL string `json:"profile_url"`
}

func createUser(w http.ResponseWriter, r *http.Request, svc *service.Service) {
	//
}

func getUser(w http.ResponseWriter, r *http.Request, svc *service.Service) {
	userID := r.URL.Query().Get("id")

	if userID == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	// Get user from database

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
