package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
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

	stringUserID := user.ID
	intUserID, err := strconv.Atoi(stringUserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := generateToken(intUserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"token": token})

}

func getUser(w http.ResponseWriter, r *http.Request, svc *service.Service) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	tokenString := r.Header.Get("Authorization")
	claims, err := validateTestJWT(tokenString)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := claims["sub"].(int)
	log.Printf("User ID: %d", userID)

	// Get user from service
	user, err := svc.GetUserByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)

}

func getAllActiveConn(pool *Pool, w http.ResponseWriter, svc *service.Service) {
	pool.mu.Lock()

	ids := make([]int, 0, len(pool.clients))
	for id := range pool.clients {
		i, err := strconv.Atoi(id)
		if err != nil {
			ids = append(ids, i)
		}
	}
	pool.mu.Unlock()

	userList := make([]service.User, 0, len(ids))

	for _, id := range ids {
		user, err := svc.GetUserByID(id)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		userList = append(userList, *user)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userList)
}

func HandleGetAllActiveConn(pool *Pool, w http.ResponseWriter, r *http.Request, svc *service.Service) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	tokenString := r.Header.Get("Authorization")
	claims, err := validateTestJWT(tokenString)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID := claims["sub"].(string)

	if userID == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}
	getAllActiveConn(pool, w, svc)
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
	tokenString := r.Header.Get("Authorization")
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

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	ID := r.URL.Query().Get("id")
	userID, err := strconv.Atoi(ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	accessToken, err := generateToken(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	refreshToken, err := generateRefreshToken(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func HandleRefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshToken := r.URL.Query().Get("refresh_token")
	claims, err := validateJWT(refreshToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	userID := claims["sub"].(int)
	accessToken, err := generateToken(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := map[string]string{
		"access_token": accessToken,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
