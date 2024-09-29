package service

import (
	"database/sql"
	"fmt"
)

type Service struct {
	mysqlDB *sql.DB
}

func NewService(
	mysqlDB *sql.DB,
) *Service {

	svc := &Service{
		mysqlDB: mysqlDB,
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

// CreateUser inserts a new user into the database
func (s *Service) CreateUser(user User) error {
	query := "INSERT INTO users (id, email, password, name, profile_url) VALUES (?, ?, ?, ?, ?)"
	_, err := s.mysqlDB.Exec(query, user.ID, user.Email, user.Password, user.Name, user.ProfileURL)
	if err != nil {
		return fmt.Errorf("error creating user: %v", err)
	}
	return nil
}

// GetUserByID retrieves a user by ID from the database
func (s *Service) GetUserByID(userID string) (*User, error) {
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
