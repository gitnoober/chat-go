package main

import (
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Validate JWT token and return hardcoded claims for testing purposes
func validateTestJWT(tokenString string) (jwt.MapClaims, error) {
	// Hardcoded claims for testing
	// hardcodedClaims := jwt.MapClaims{
	// 	"sub":  "user123",         // User ID
	// 	"name": "John Doe",        // User's name
	// 	"iat":  time.Now().Unix(), // Issued at timestamp
	// 	"exp":  time.Now().Add(1 * time.Hour).Unix(), // Expiration timestamp
	// }
	log.Println("Token string:", tokenString)

	// Simulating validation of the token (you can skip the actual check here)
	if tokenString == "valid_token" { // Use a placeholder for a valid token check
		hardcodedClaims := jwt.MapClaims{
			"sub":  1,                                    // User ID
			"name": "John Doe",                           // User's name
			"iat":  time.Now().Unix(),                    // Issued at timestamp
			"exp":  time.Now().Add(1 * time.Hour).Unix(), // Expiration timestamp
		}
		log.Println("Hardcoded claims:", hardcodedClaims)
		return hardcodedClaims, nil
	}
	if tokenString == "valid_string2" {
		hardcodedClaims := jwt.MapClaims{
			"sub":  2,                                    // User ID
			"name": "Jon Winslow",                        // User's name
			"iat":  time.Now().Unix(),                    // Issued at timestamp
			"exp":  time.Now().Add(1 * time.Hour).Unix(), // Expiration timestamp
		}
		log.Println("Hardcoded claims:", hardcodedClaims)
		return hardcodedClaims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// Validate JWT token and return claims
func validateJWT(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}
	log.Printf("Claims later part: %v", claims)

	return claims, nil
}

func generateToken(userID int) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func generateRefreshToken(userID int) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(7 * 24 * time.Hour).Unix(), // 7 days
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
