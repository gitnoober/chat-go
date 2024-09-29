package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("secret-key") // TOOD: Move this somewhere else

// Client represents a websocket client
type Client struct {
	ID string
	Conn *websocket.Conn
}

// Pool manages all active connections
type Pool struct {
	clients map[string]*Client
	mu sync.Mutex
}

// Create a new Pool
func newPool() *Pool{
	return &Pool{
		clients: make(map[string]*Client),
	}
}

// Add a new client to Pool
func (pool *Pool) AddClient(client *Client){
	pool.mu.Lock()

	defer pool.mu.Unlock()

	pool.clients[client.ID] = client
}

// Remove a client from the Pool
func (pool *Pool) RemoveClient(clientID string){
	pool.mu.Lock()
	defer pool.mu.Unlock()
	delete(pool.clients, clientID)
}

// Send a message to a specific client
func (pool *Pool) SendMessage(ReceiverID string, message string) error {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	receiver, ok := pool.clients[ReceiverID]
	if !ok{
		fmt.Errorf("client not found: %s", ReceiverID)
	}
	w, err := receiver.Conn.Writer(ctx, websocket.MessageText)
	if err != nil {
		return fmt.Errorf("failed to get writer: %w", err)
	}
	defer w.Close()

	// Write the message
	if _, err := w.Write([]byte (message)); err != nil {
		return fmt.Errorf("error sending message: %v", err)
	}

	return nil
}

// Validate JWT token and return claims
func validateJWT(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return token, nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	return claims, nil
}

// Splitmessage
func splitMessage(message string) []string{
	var parts[]string
	splitIdx := -1
	for i, r := range message {
		if r == ':'{
			splitIdx = i
			break
		}
	}
	if splitIdx > 0 {
		parts = append(parts, message[:splitIdx], message[splitIdx+1:])

	}
	return parts
}


// Handle incoming websocket connections
func handleWebSocket(pool *Pool, w http.ResponseWriter, r *http.Request){
	tokenString := r.URL.Query().Get("token")
	claims, err := validateJWT(tokenString)
	if err != nil{
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Printf("Websocket connection err: %v", err)
		return
	}

	clientID := claims["sub"].( string)
	client := &Client{
		ID: clientID,
		Conn: conn,
	}
	pool.AddClient(client)

	defer pool.RemoveClient(clientID)

	log.Printf("Client connected: %s", clientID)

	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
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

func main(){
	pool := newPool()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(pool, w, r)
	})

	srv := &http.Server{
		Addr: ":8080",
		WriteTimeout: 10 * time.Second,
		ReadTimeout: 10 * time.Second,
		IdleTimeout: 120 * time.Second,
	}

	// Channel to listen for interrupt signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	go func(){
		<- sigs
		log.Println("Received shutdown signal, shutting down gracefully.....")

		// Create a context for shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("Server forced to shutdown: %v", err)
		}
		log.Println("Server gracefully shutdown")
	}()

	log.Println("Starting server on :8080")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to listen and serve: %v", err)
	}
}
