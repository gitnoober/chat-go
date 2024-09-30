package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Service struct {
	mysqlDB *sql.DB
	redisDB *redis.Client
}

func NewService(
	mysqlDB *sql.DB,
	redisDB *redis.Client,
) *Service {

	svc := &Service{
		mysqlDB: mysqlDB,
		redisDB: redisDB,
	}
	return svc
}

type User struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	Name       string `json:"name"`
	ProfileURL string `json:"profile_url"`
}

// | Table | Create Table                                                                                                                                                                                                                                                                                                                       |
// +-------+------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
// | users | CREATE TABLE `users` (
//   `id` int NOT NULL AUTO_INCREMENT,
//   `email` varchar(255) NOT NULL,
//   `password` varchar(255) NOT NULL,
//   `name` varchar(255) NOT NULL,
//   `profile_url` varchar(255) NOT NULL,
//   PRIMARY KEY (`id`),
//   UNIQUE KEY `idx_email` (`email`)
// ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci |

// CreateUser inserts a new user into the database
func (s *Service) CreateUser(user User) error {
	query := "INSERT INTO users (email, password, name, profile_url) VALUES (?, ?, ?, ?)"
	_, err := s.mysqlDB.Exec(query, user.Email, user.Password, user.Name, user.ProfileURL)
	if err != nil {
		return fmt.Errorf("error creating user: %v", err)
	}
	return nil
}

// GetUserByID retrieves a user by ID from the database
func (s *Service) GetUserByID(userID int) (*User, error) {
	query := "SELECT id, email, password, name, profile_url FROM users WHERE id = ?"
	row := s.mysqlDB.QueryRow(query, userID)

	var user User
	if err := row.Scan(&user.ID, &user.Email, &user.Password, &user.Name, &user.ProfileURL); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("error retrieving user: %v", err)
	}

	return &user, nil
}

func (s *Service) GetUserByEmail(email string) (int, error) {
	query := "SELECT id FROM users WHERE email = ?"
	row := s.mysqlDB.QueryRow(query, email)

	var UserID int
	err := row.Scan(&UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("user not found")
		}
		return 0, fmt.Errorf("error retrieving user: %v", err)
	}
	return UserID, nil
}

func (s *Service) GetRedisData(key string) (string, error) {
	val, err := s.redisDB.Get(context.Background(), key).Result()
	if err != nil {
		return "", fmt.Errorf("error getting data from redis: %v", err)
	}
	return val, nil
}

func (s *Service) SetRedisData(key, value string, expiration time.Duration) error {
	err := s.redisDB.Set(context.Background(), key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("error setting data in redis: %v", err)
	}
	return nil
}