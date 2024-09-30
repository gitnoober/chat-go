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
	thirdparty "github.com/gitnoober/chat-go/third-party"
	"golang.org/x/crypto/bcrypt"
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
	profile_url := thirdparty.GetRandomProfilePicture(user.Email)
	user.ProfileURL = profile_url

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user.Password = string(hashedPassword)

	// Call the service to create the user
	if err := svc.CreateUser(user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userID, err := svc.GetUserByEmail(user.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"userID": userID})

}

func getUser(w http.ResponseWriter, r *http.Request, svc *service.Service) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	tokenString := r.Header.Get("Authorization")
	log.Printf("Token: %s", tokenString)
	claims, err := validateJWT(tokenString)
	log.Printf("Claims: %v", claims)
	log.Printf("Error: %v", err)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := int(claims["sub"].(float64))
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
	defer pool.mu.Unlock()

	ids := make([]string, 0, len(pool.clients))
	for id := range pool.clients {
		ids = append(ids, id)
	}
	log.Printf("ids: %v", ids)

	userList := make([]service.User, 0, len(ids))

	for _, id := range ids {
		userID, err := strconv.Atoi(id)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			continue
		}
		user, err := svc.GetUserByID(userID)
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
	_, err := validateJWT(tokenString)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	log.Printf("pool: %v", pool)
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
	tokenString := r.URL.Query().Get("token")
	// Log the connection request
	log.Printf("Received WebSocket connection request with token: %s", tokenString)

	// claims, err := validateJWT(tokenString)
	claims, err := validateJWT(tokenString)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Printf("Websocket connection err: %v", err)
		return
	}

	ID := claims["sub"].(float64)
	clientID := strconv.Itoa(int(ID))
	client := &Client{
		ID:   clientID,
		Conn: conn,
	}
	pool.AddClient(client)

	log.Printf("Client added to pool: %v", pool)

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

func HandleLogin(w http.ResponseWriter, r *http.Request, svc *service.Service) {
	ID := r.URL.Query().Get("id")
	userID, err := strconv.Atoi(ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	password := r.URL.Query().Get("password")
	if password == "" {
		http.Error(w, "Password is required", http.StatusBadRequest)
		return
	}
	user, uErr := svc.GetUserByID(userID)
	if uErr != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
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

	rErr := addRefreshToken(refreshToken, svc)
	if rErr != nil {
		http.Error(w, rErr.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func HandleRefreshToken(w http.ResponseWriter, r *http.Request, svc *service.Service) {
	refreshToken := r.URL.Query().Get("refresh_token")
	claims, err := validateJWT(refreshToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	isUserAuth, rErr := validateRefreshToken(refreshToken, svc)
	if rErr != nil {
		http.Error(w, rErr.Error(), http.StatusInternalServerError)
		return
	}
	if !isUserAuth {
		http.Error(w, "Unauthorized! Log in Again!", http.StatusUnauthorized)
		return
	}

	userID := int(claims["sub"].(float64))
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
