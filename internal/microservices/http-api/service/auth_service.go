package service

// upgrade to OAUTH2.1
import (
	"errors"
	"fmt"
	"mangahub/internal/config"
	"mangahub/internal/microservices/http-api/models"
	"mangahub/internal/microservices/http-api/repository"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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
	Login(username, password, email string) (accessToken, refreshToken string, user *models.User, err error)
	RefreshAccessToken(refreshToken string) (newAccessToken, newRefreshToken string, err error)
	ValidateToken(tokenString string) (*Claims, error)
	RevokeToken(refreshToken string) error
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

// claims structure for JWT tokens
type Claims struct {
	UserID               string   `json:"user_id"`
	Username             string   `json:"username"`
	Email                string   `json:"email"`
	Scopes               []string `json:"scopes"` // e.g., ["read", "write"] => list of permissions granted
	Role                 string   `json:"role"`
	jwt.RegisteredClaims          // embeds standard claims like exp, aud, subject, issuer
}

// ScopeString returns the scopes as a single space-separated string => e.g., "read write"
func (c Claims) ScopeString() string {
	return strings.Join(c.Scopes, " ")
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
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
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
func (s *authService) Login(username, password, email string) (string, string, *models.User, error) {
	// Find user
	var user *models.User
	var err error
	if username != "" {
		user, err = s.userRepo.FindByUsername(username)
	} else if email != "" {
		user, err = s.userRepo.FindByEmail(email)
	}
	if err != nil {
		// User not found we use dummy compare to mitigate timing attacks (always take same time)
		_ = bcrypt.CompareHashAndPassword([]byte("$2a$10$7EqJtq98hPqEX7fNZaFWoOhi6Cq1h0u3b0j3Z6h5y5jY5f5h5F5eW"), []byte(password))
		return "", "", nil, ErrInvalidCredentials
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", "", nil, ErrInvalidCredentials
	}

	// Generate access token (short-lived, 15 min)
	accessToken, err := s.generateAccessTokenWithScopes(user) // default role is "user"
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

// this version is simple JWT generation without OAUTH2.1 specifics
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

// generateAccessTokenWithScopes: generates an access token with specific scopes based on user role or custom scopes.
func (s *authService) generateAccessTokenWithScopes(user *models.User, customScopes ...string) (string, error) {
	// Define default scopes based on role
	scopeByRole := map[string][]string{
		"admin": {"read:*", "write:*", "delete:*", "admin:*"},
		"user":  {"read:manga", "write:comment", "write:profile", "write:community_chat"},
	}

	// Get custom scopes if provided, else use default based on role
	var scopes []string
	if len(customScopes) > 0 {
		scopes = customScopes
	} else {
		scopes = scopeByRole[user.Role]
	}

	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		Scopes:   scopes,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "mangahub",
			Subject:   user.ID,
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// generateAccessTokenWithRequestedScopes: generates an access token with specific requested scopes after validating them against allowed scopes.
// This is useful for OAUTH2.1 where clients can request specific scopes during authorization.
func (s *authService) generateAccessTokenWithRequestedScopes(user *models.User, requestedScopes []string) (string, error) {
	// Define allowed scopes based on user role
	allowedScopesByRole := map[string][]string{
		"admin": {"read:*", "write:*", "delete:*", "admin:*"},
		"user":  {"read:manga", "write:comment", "write:profile", "write:community_chat"},
	}
	allowed := allowedScopesByRole[user.Role]

	// Filter requested scopes to only those allowed for this role
	var grantedScopes []string
	for _, requested := range requestedScopes {
		// set up wildcard support
		prefix := requested[:strings.Index(requested, ":")+1]
		wildcard := prefix + "*"
		if contains(allowed, requested) || contains(allowed, wildcard) {
			grantedScopes = append(grantedScopes, requested)
		}
	}

	return s.generateAccessTokenWithScopes(user, grantedScopes...)
}

// Contains checks if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item || s == "*" || matchesWildcard(s, item) {
			return true
		}
	}
	return false
}

// matchesWildcard checks if the item matches the pattern with wildcard support
func matchesWildcard(pattern, item string) bool {
	// Simple wildcard matching: "read:*" matches "read:products", "read:orders", etc.
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' { // ends with *
		return strings.HasPrefix(item, pattern[:len(pattern)-1]) //return true if item starts with pattern prefix
	}
	return pattern == item
}

// generateRefreshToken: creates a new refresh token for the user and stores it in the database.
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

func (s *authService) RefreshAccessToken(refreshTokenString string) (string, string, error) {
	// Validate refresh token
	refreshToken, err := s.refreshTokenRepo.FindByToken(refreshTokenString)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	// Check expiration
	if time.Now().After(refreshToken.ExpiresAt) {
		s.refreshTokenRepo.Delete(refreshToken.ID)
		return "", "", errors.New("refresh token expired")
	}

	// Check if revoked
	if refreshToken.Revoked {
		s.refreshTokenRepo.Delete(refreshToken.ID)
		return "", "", errors.New("refresh token revoked")
	}

	// Get user
	user, err := s.userRepo.FindByID(refreshToken.UserID)
	if err != nil {
		return "", "", err
	}
	// Rotate refresh token
	// Invalidate the old refresh token
	if err := s.refreshTokenRepo.Revoke(refreshToken.ID); err != nil {
		s.refreshTokenRepo.Delete(refreshToken.ID)
		return "", "", err
	}
	// Issue a new access token
	newAccessToken, err := s.generateAccessTokenWithScopes(user)
	if err != nil {
		return "", "", err
	}
	// Issue a new refresh token
	newRefreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return "", "", err
	}
	// Generate new access token
	return newAccessToken, newRefreshToken, nil
}

func (s *authService) ValidateToken(tokenString string) (*Claims, error) {
	// prepare empty claims struct for parsing(prevent panic)
	claims := &Claims{}
	// parse with claims
	token, err := jwt.ParseWithClaims(tokenString, claims,
		// take jwt.token and check signing method
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("invalid signing method")
			}
			// if valid, return the secret key for validation
			return []byte(s.jwtSecret), nil
		})

	// check for parsing errors / invalid signing method
	if err != nil {
		return nil, err
	}
	// check token validity
	if !token.Valid {
		return nil, ErrInvalidToken
	}

	// check additional information in claims
	// check expiration
	if claims.ExpiresAt == nil || time.Now().After(claims.ExpiresAt.Time) {
		return nil, ErrExpiredToken
	}

	// check issuer
	if claims.Issuer != "mangahub" {
		return nil, ErrInvalidToken
	}
	// check subject
	if claims.Subject == "" {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func (s *authService) RevokeToken(refreshTokenString string) error {
	// Validate refresh token
	refreshToken, err := s.refreshTokenRepo.FindByToken(refreshTokenString)
	if err != nil {
		fmt.Println("Error finding refresh token:", err)
		return nil // return nil to ignore leakage of token validity
	}
	// Revoke the token
	if err := s.refreshTokenRepo.Revoke(refreshToken.ID); err != nil {
		fmt.Println("Error revoking refresh token:", err)
		return nil // Ignore errors during revocation
	}

	fmt.Println("Refresh token revoked successfully")
	return nil
}
