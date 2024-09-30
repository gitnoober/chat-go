package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/joho/godotenv"
	"go.uber.org/ratelimit"

	"github.com/gitnoober/chat-go/config"
	"github.com/gitnoober/chat-go/service"
)


// Client represents a websocket client
type Client struct {
	ID   string
	Conn *websocket.Conn
}

// Pool manages all active connections
type Pool struct {
	clients map[string]*Client
	mu      sync.Mutex
}

// Create a new Pool
func newPool() *Pool {
	return &Pool{
		clients: make(map[string]*Client),
	}
}

// Add a new client to Pool
func (pool *Pool) AddClient(client *Client) {
	pool.mu.Lock()

	defer pool.mu.Unlock()

	pool.clients[client.ID] = client
}

// Remove a client from the Pool
func (pool *Pool) RemoveClient(clientID string) {
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
	if !ok {
		fmt.Printf("client not found: %s", ReceiverID)
	}
	w, err := receiver.Conn.Writer(ctx, websocket.MessageText)
	if err != nil {
		return fmt.Errorf("failed to get writer: %w", err)
	}
	defer w.Close()

	// Write the message
	if _, err := w.Write([]byte(message)); err != nil {
		return fmt.Errorf("error sending message: %v", err)
	}

	return nil
}

// Splitmessage
func splitMessage(message string) []string {
	var parts []string
	splitIdx := -1
	for i, r := range message {
		if r == ':' {
			splitIdx = i
			break
		}
	}
	if splitIdx > 0 {
		parts = append(parts, message[:splitIdx], message[splitIdx+1:])

	}
	return parts
}

var jwtSecret []byte

func init() {
    // Load the .env file
    err := godotenv.Load()
    if err != nil {
        log.Fatalf("Error loading .env file: %v", err)
    }

    // Set jwtSecret from environment variable
    secret := os.Getenv("JWT_SECRET")
    if secret == "" {
        log.Fatalf("JWT_SECRET environment variable not set")
    }
    jwtSecret = []byte(secret)

	gravatarSecret := os.Getenv("GRAVATAR_ACCESS_KEY")
	if gravatarSecret == "" {
		log.Fatalf("GRAVATAR_ACCESS_KEY environment variable not set")
	}
}

func main() {
	// Capture connection properties
	var db *sql.DB

	cfg := config.LoadConfig()

	db, err := config.ConnectMysql(cfg, db)
	if err != nil {
		log.Fatal(err)
	}
	redisDB := config.ConnectRedis(cfg)
	svc := service.NewService(db, redisDB)

	pool := newPool()

	rl := ratelimit.New(100) // per second

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		rl.Take()
		HandleWebSocket(pool, w, r)
	})
	http.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		rl.Take()
		HandleUser(w, r, svc)
	})
	http.HandleFunc("/online-users", func(w http.ResponseWriter, r *http.Request) {
		rl.Take()
		HandleGetAllActiveConn(pool, w, r, svc)
	})
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		rl.Take()
		HandleLogin(w, r, svc)
	})
	http.HandleFunc("/refresh", func(w http.ResponseWriter, r *http.Request) {
		rl.Take()
		HandleRefreshToken(w, r, svc)
	})

	srv := &http.Server{
		Addr:         ":8080",
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Channel to listen for interrupt signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	go func() {
		<-sigs
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
