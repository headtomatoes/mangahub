package service

import (
	"context"
	"errors"
	"mangahub/internal/config"
	"mangahub/internal/microservices/http-api/models"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// MockUserRepository mocks the UserRepository interface
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) FindByID(id string) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetAllIDs(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// MockRefreshTokenRepository mocks the RefreshTokenRepository interface
type MockRefreshTokenRepository struct {
	mock.Mock
}

func (m *MockRefreshTokenRepository) Create(token *models.RefreshToken) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) FindByToken(token string) (*models.RefreshToken, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RefreshToken), args.Error(1)
}

func (m *MockRefreshTokenRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) Revoke(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) DeleteExpired() error {
	args := m.Called()
	return args.Error(0)
}

func TestRegister_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	cfg := &config.Config{
		JWTSecret:       "test-secret",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	}
	authService := NewAuthService(mockUserRepo, mockRefreshTokenRepo, cfg)

	mockUserRepo.On("FindByUsername", "testuser").Return(nil, gorm.ErrRecordNotFound)
	mockUserRepo.On("FindByEmail", "test@example.com").Return(nil, gorm.ErrRecordNotFound)
	mockUserRepo.On("Create", mock.AnythingOfType("*models.User")).Return(nil)

	user, err := authService.Register("testuser", "password123", "test@example.com")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "test@example.com", user.Email)
	mockUserRepo.AssertExpectations(t)
}

func TestRegister_UsernameExists(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	cfg := &config.Config{JWTSecret: "test-secret"}
	authService := NewAuthService(mockUserRepo, mockRefreshTokenRepo, cfg)

	existingUser := &models.User{Username: "testuser"}
	mockUserRepo.On("FindByUsername", "testuser").Return(existingUser, nil)

	user, err := authService.Register("testuser", "password123", "test@example.com")

	assert.Error(t, err)
	assert.Equal(t, ErrNameInUse, err)
	assert.Nil(t, user)
	mockUserRepo.AssertExpectations(t)
}

func TestRegister_EmailExists(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	cfg := &config.Config{JWTSecret: "test-secret"}
	authService := NewAuthService(mockUserRepo, mockRefreshTokenRepo, cfg)

	existingUser := &models.User{Email: "test@example.com"}
	mockUserRepo.On("FindByUsername", "testuser").Return(nil, gorm.ErrRecordNotFound)
	mockUserRepo.On("FindByEmail", "test@example.com").Return(existingUser, nil)

	user, err := authService.Register("testuser", "password123", "test@example.com")

	assert.Error(t, err)
	assert.Equal(t, ErrEmailInUse, err)
	assert.Nil(t, user)
	mockUserRepo.AssertExpectations(t)
}

func TestLogin_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	cfg := &config.Config{
		JWTSecret:       "test-secret",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	}
	authService := NewAuthService(mockUserRepo, mockRefreshTokenRepo, cfg)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := &models.User{
		ID:       "user-id",
		Username: "testuser",
		Email:    "test@example.com",
		Password: string(hashedPassword),
		Role:     "user",
	}

	mockUserRepo.On("FindByUsername", "testuser").Return(user, nil)
	mockRefreshTokenRepo.On("Create", mock.AnythingOfType("*models.RefreshToken")).Return(nil)

	accessToken, refreshToken, returnedUser, err := authService.Login("testuser", "password123", "")

	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.Equal(t, user.Username, returnedUser.Username)
	mockUserRepo.AssertExpectations(t)
	mockRefreshTokenRepo.AssertExpectations(t)
}

func TestLogin_InvalidPassword(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	cfg := &config.Config{JWTSecret: "test-secret"}
	authService := NewAuthService(mockUserRepo, mockRefreshTokenRepo, cfg)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := &models.User{
		ID:       "user-id",
		Username: "testuser",
		Password: string(hashedPassword),
	}

	mockUserRepo.On("FindByUsername", "testuser").Return(user, nil)

	accessToken, refreshToken, returnedUser, err := authService.Login("testuser", "wrongpassword", "")

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
	assert.Nil(t, returnedUser)
	mockUserRepo.AssertExpectations(t)
}

func TestLogin_UserNotFound(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	cfg := &config.Config{JWTSecret: "test-secret"}
	authService := NewAuthService(mockUserRepo, mockRefreshTokenRepo, cfg)

	mockUserRepo.On("FindByUsername", "nonexistent").Return(nil, gorm.ErrRecordNotFound)

	accessToken, refreshToken, user, err := authService.Login("nonexistent", "password123", "")

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
	assert.Nil(t, user)
	mockUserRepo.AssertExpectations(t)
}

func TestValidateToken_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	cfg := &config.Config{
		JWTSecret:      "test-secret",
		AccessTokenTTL: 15 * time.Minute,
	}
	authService := NewAuthService(mockUserRepo, mockRefreshTokenRepo, cfg)

	claims := Claims{
		UserID:   "user-id",
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "user",
		Scopes:   []string{"read:manga"},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "mangahub",
			Subject:   "user-id",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(cfg.JWTSecret))

	validatedClaims, err := authService.ValidateToken(tokenString)

	assert.NoError(t, err)
	assert.NotNil(t, validatedClaims)
	assert.Equal(t, "testuser", validatedClaims.Username)
}

func TestValidateToken_Expired(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	cfg := &config.Config{JWTSecret: "test-secret"}
	authService := NewAuthService(mockUserRepo, mockRefreshTokenRepo, cfg)

	claims := Claims{
		UserID:   "user-id",
		Username: "testuser",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			Issuer:    "mangahub",
			Subject:   "user-id",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(cfg.JWTSecret))

	validatedClaims, err := authService.ValidateToken(tokenString)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
	assert.Nil(t, validatedClaims)
}

func TestValidateToken_InvalidSigningMethod(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	cfg := &config.Config{JWTSecret: "test-secret"}
	authService := NewAuthService(mockUserRepo, mockRefreshTokenRepo, cfg)

	// This would require RS256 key pair, so we'll just test with invalid token
	invalidToken := "invalid.token.here"

	validatedClaims, err := authService.ValidateToken(invalidToken)

	assert.Error(t, err)
	assert.Nil(t, validatedClaims)
}

func TestValidateToken_InvalidIssuer(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	cfg := &config.Config{JWTSecret: "test-secret"}
	authService := NewAuthService(mockUserRepo, mockRefreshTokenRepo, cfg)

	claims := Claims{
		UserID:   "user-id",
		Username: "testuser",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			Issuer:    "wrong-issuer",
			Subject:   "user-id",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(cfg.JWTSecret))

	validatedClaims, err := authService.ValidateToken(tokenString)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidToken, err)
	assert.Nil(t, validatedClaims)
}

func TestValidateToken_EmptySubject(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	cfg := &config.Config{JWTSecret: "test-secret"}
	authService := NewAuthService(mockUserRepo, mockRefreshTokenRepo, cfg)

	claims := Claims{
		UserID:   "user-id",
		Username: "testuser",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			Issuer:    "mangahub",
			Subject:   "",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(cfg.JWTSecret))

	validatedClaims, err := authService.ValidateToken(tokenString)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidToken, err)
	assert.Nil(t, validatedClaims)
}

func TestRefreshAccessToken_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	cfg := &config.Config{
		JWTSecret:       "test-secret",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	}
	authService := NewAuthService(mockUserRepo, mockRefreshTokenRepo, cfg)

	refreshToken := &models.RefreshToken{
		ID:        "token-id",
		UserID:    "user-id",
		Token:     "refresh-token",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		Revoked:   false,
	}
	user := &models.User{
		ID:       "user-id",
		Username: "testuser",
		Role:     "user",
	}

	mockRefreshTokenRepo.On("FindByToken", "refresh-token").Return(refreshToken, nil)
	mockUserRepo.On("FindByID", "user-id").Return(user, nil)
	mockRefreshTokenRepo.On("Revoke", "token-id").Return(nil)
	mockRefreshTokenRepo.On("Create", mock.AnythingOfType("*models.RefreshToken")).Return(nil)

	newAccessToken, newRefreshToken, err := authService.RefreshAccessToken("refresh-token")

	assert.NoError(t, err)
	assert.NotEmpty(t, newAccessToken)
	assert.NotEmpty(t, newRefreshToken)
	mockRefreshTokenRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestRefreshAccessToken_TokenExpired(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	cfg := &config.Config{JWTSecret: "test-secret"}
	authService := NewAuthService(mockUserRepo, mockRefreshTokenRepo, cfg)

	refreshToken := &models.RefreshToken{
		ID:        "token-id",
		Token:     "expired-token",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}

	mockRefreshTokenRepo.On("FindByToken", "expired-token").Return(refreshToken, nil)
	mockRefreshTokenRepo.On("Delete", "token-id").Return(nil)

	newAccessToken, newRefreshToken, err := authService.RefreshAccessToken("expired-token")

	assert.Error(t, err)
	assert.Empty(t, newAccessToken)
	assert.Empty(t, newRefreshToken)
	mockRefreshTokenRepo.AssertExpectations(t)
}

func TestRefreshAccessToken_TokenRevoked(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	cfg := &config.Config{JWTSecret: "test-secret"}
	authService := NewAuthService(mockUserRepo, mockRefreshTokenRepo, cfg)

	refreshToken := &models.RefreshToken{
		ID:        "token-id",
		Token:     "revoked-token",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		Revoked:   true,
	}

	mockRefreshTokenRepo.On("FindByToken", "revoked-token").Return(refreshToken, nil)
	mockRefreshTokenRepo.On("Delete", "token-id").Return(nil)

	newAccessToken, newRefreshToken, err := authService.RefreshAccessToken("revoked-token")

	assert.Error(t, err)
	assert.Empty(t, newAccessToken)
	assert.Empty(t, newRefreshToken)
	mockRefreshTokenRepo.AssertExpectations(t)
}

func TestRefreshAccessToken_UserNotFound(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	cfg := &config.Config{JWTSecret: "test-secret"}
	authService := NewAuthService(mockUserRepo, mockRefreshTokenRepo, cfg)

	refreshToken := &models.RefreshToken{
		ID:        "token-id",
		UserID:    "nonexistent-user",
		Token:     "refresh-token",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		Revoked:   false,
	}

	mockRefreshTokenRepo.On("FindByToken", "refresh-token").Return(refreshToken, nil)
	mockUserRepo.On("FindByID", "nonexistent-user").Return(nil, errors.New("user not found"))

	newAccessToken, newRefreshToken, err := authService.RefreshAccessToken("refresh-token")

	assert.Error(t, err)
	assert.Empty(t, newAccessToken)
	assert.Empty(t, newRefreshToken)
	mockRefreshTokenRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestScopeString(t *testing.T) {
	claims := Claims{
		Scopes: []string{"read:manga", "write:comment", "delete:library"},
	}

	scopeString := claims.ScopeString()
	assert.Equal(t, "read:manga write:comment delete:library", scopeString)
}

func TestScopeString_Empty(t *testing.T) {
	claims := Claims{
		Scopes: []string{},
	}

	scopeString := claims.ScopeString()
	assert.Equal(t, "", scopeString)
}

func TestMatchesWildcard(t *testing.T) {
	tests := []struct {
		pattern  string
		item     string
		expected bool
	}{
		{"read:*", "read:manga", true},
		{"read:*", "read:library", true},
		{"read:*", "write:manga", false},
		{"write:comment", "write:comment", true},
		{"write:comment", "write:profile", false},
	}

	for _, tt := range tests {
		result := matchesWildcard(tt.pattern, tt.item)
		assert.Equal(t, tt.expected, result)
	}
}

func TestContains(t *testing.T) {
	slice := []string{"read:manga", "write:*", "delete:library"}

	assert.True(t, contains(slice, "read:manga"))
	assert.True(t, contains(slice, "write:comment"))
	assert.False(t, contains(slice, "admin:users"))
}

func TestRevokeToken_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	cfg := &config.Config{JWTSecret: "test-secret"}
	authService := NewAuthService(mockUserRepo, mockRefreshTokenRepo, cfg)

	refreshToken := &models.RefreshToken{
		ID:    "token-id",
		Token: "refresh-token",
	}

	mockRefreshTokenRepo.On("FindByToken", "refresh-token").Return(refreshToken, nil)
	mockRefreshTokenRepo.On("Revoke", "token-id").Return(nil)

	err := authService.RevokeToken("refresh-token")

	assert.NoError(t, err)
	mockRefreshTokenRepo.AssertExpectations(t)
}

func TestRevokeToken_NotFound(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	cfg := &config.Config{JWTSecret: "test-secret"}
	authService := NewAuthService(mockUserRepo, mockRefreshTokenRepo, cfg)

	mockRefreshTokenRepo.On("FindByToken", "invalid-token").Return(nil, errors.New("not found"))

	err := authService.RevokeToken("invalid-token")

	assert.NoError(t, err) // Should return nil to avoid leaking token validity
	mockRefreshTokenRepo.AssertExpectations(t)
}
