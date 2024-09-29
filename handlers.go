package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/coder/websocket"
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

func HandleUser(w http.ResponseWriter, r *http.Request, svc *service.Service) {
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


// Handle incoming websocket connections
func HandleWebSocket(pool *Pool, w http.ResponseWriter, r *http.Request) {
	tokenString := r.URL.Query().Get("token")
	// Log the connection request
	log.Printf("Received WebSocket connection request with token: %s", tokenString)

	// claims, err := validateJWT(tokenString)
	claims, err := validateTestJWT(tokenString)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Printf("Websocket connection err: %v", err)
		return
	}

	clientID := claims["sub"].(string)
	client := &Client{
		ID:   clientID,
		Conn: conn,
	}
	pool.AddClient(client)

	defer pool.RemoveClient(clientID)

	log.Printf("Client connected: %s", clientID)

	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*100)
		defer cancel()

		// Read the message from the client
		_, reader, err := conn.Reader(ctx)
		if err != nil {
			log.Printf("Read Error: $%v", err)
			break
		}

		// Read the entire message from the io.Reader
		message, err := io.ReadAll(reader)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		// log.Println("Received message:", string(message))

		// Assume the message format is "receiverID:message"
		parts := splitMessage(string(message))
		if len(parts) != 2 {
			log.Println("Invalid message format")
			continue
		}
		receiverID, msg := parts[0], parts[1]

		// Send the message to the intended recipient
		if err := pool.SendMessage(receiverID, msg); err != nil {
			log.Printf("Send message error: %v", err)
		}
	}
}