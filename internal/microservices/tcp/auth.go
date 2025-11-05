package tcp

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type TCPAuthService struct {
	jwtSecret string
}

func NewTCPAuthService(jwtSecret string) *TCPAuthService {
	return &TCPAuthService{jwtSecret: jwtSecret}
}

func (a *TCPAuthService) ValidateToken(tokenString string) (string, string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(a.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return "", "", fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", errors.New("invalid token claims")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", "", errors.New("user_id claim is not a string")
	}

	username, ok := claims["username"].(string)
	if !ok {
		return "", "", errors.New("username claim is not a string")
	}

	return userID, username, nil
}
