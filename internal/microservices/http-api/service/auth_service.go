package service

import (
	"errors"
	"mangahub/internal/config"
	"mangahub/internal/microservices/http-api/middleware/auth"
	"mangahub/internal/microservices/http-api/models"
	"mangahub/internal/microservices/http-api/repository"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrNameInUse          = errors.New("username already in use")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrExpiredToken       = errors.New("token has expired")
	ErrEmailInUse         = errors.New("email already in use")
)

type AuthService interface {
	Register(username, password, email string) (*models.User, error)
	Login(username, password string) (accessToken, refreshToken string, user *models.User, err error)
	RefreshAccessToken(refreshToken string) (newAccessToken string, err error)
	ValidateToken(tokenString string) (*jwt.Token, error)
}

type authService struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	jwtSecret        string
	accessTokenTTL   time.Duration
	refreshTokenTTL  time.Duration
}

func NewAuthService(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	cfg *config.Config,
) AuthService {
	return &authService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtSecret:        cfg.JWTSecret,
		accessTokenTTL:   cfg.AccessTokenTTL,  // 15 minutes
		refreshTokenTTL:  cfg.RefreshTokenTTL, // 7 days
	}
}

// Todo: add context
// Todo: add logging
// Todo: add metrics
// Todo: add DTOs for input rather than passing raw strings
// Register: registers a new user with the given username, password, and email.
func (s *authService) Register(username, password, email string) (*models.User, error) {
	// Check if user exists
	if _, err := s.userRepo.FindByUsername(username); err == nil {
		return nil, ErrNameInUse
	}

	// Check if email exists
	if _, err := s.userRepo.FindByEmail(email); err == nil {
		return nil, ErrEmailInUse
	}
	// Hash password
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Create user Struct
	user := &models.User{
		ID:       uuid.New().String(),
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
	}

	// Save user Struct to DB
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login: authenticates a user and returns access and refresh tokens upon successful login.
func (s *authService) Login(username, password string) (string, string, *models.User, error) {
	// Find user
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		// User not found we use dummy compare to mitigate timing attacks (always take same time)
		auth.VerifyPassword("$2a$10$7EqJtq98hPqEX7fNZaFWoOHi6VbU5h6K9v8u5rO0m3j0h6dX5r8e", password)
		return "", "", nil, ErrInvalidCredentials
	}

	// Verify password
	if err := auth.VerifyPassword(user.Password, password); err != nil {
		return "", "", nil, ErrInvalidCredentials
	}

	// Generate access token (short-lived, 15 min)
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return "", "", nil, err
	}

	// Generate refresh token (long-lived, 7 days)
	refreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return "", "", nil, err
	}

	return accessToken, refreshToken, user, nil
}

func (s *authService) generateAccessToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(s.accessTokenTTL).Unix(),
		"iat":      time.Now().Unix(),
		"type":     "access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *authService) generateRefreshToken(user *models.User) (string, error) {
	refreshToken := &models.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		Token:     uuid.New().String(), // Simple UUID as refresh token
		ExpiresAt: time.Now().Add(s.refreshTokenTTL),
	}

	if err := s.refreshTokenRepo.Create(refreshToken); err != nil {
		return "", err
	}

	return refreshToken.Token, nil
}

func (s *authService) RefreshAccessToken(refreshTokenString string) (string, error) {
	// Validate refresh token
	refreshToken, err := s.refreshTokenRepo.FindByToken(refreshTokenString)
	if err != nil {
		return "", errors.New("invalid refresh token")
	}

	// Check expiration
	if time.Now().After(refreshToken.ExpiresAt) {
		s.refreshTokenRepo.Delete(refreshToken.ID)
		return "", errors.New("refresh token expired")
	}

	// Get user
	user, err := s.userRepo.FindByID(refreshToken.UserID)
	if err != nil {
		return "", err
	}

	// Generate new access token
	return s.generateAccessToken(user)
}

func (s *authService) ValidateToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return token, nil
}
