package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"mangahub/internal/microservices/http-api/dto"
	"mangahub/internal/microservices/http-api/models"
	"mangahub/internal/microservices/http-api/service"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService mocks the AuthService interface
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(username, password, email string) (*models.User, error) {
	args := m.Called(username, password, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthService) Login(username, password, email string) (string, string, *models.User, error) {
	args := m.Called(username, password, email)
	return args.String(0), args.String(1), args.Get(2).(*models.User), args.Error(3)
}

func (m *MockAuthService) RefreshAccessToken(refreshToken string) (string, string, error) {
	args := m.Called(refreshToken)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockAuthService) ValidateToken(tokenString string) (*service.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.Claims), args.Error(1)
}

func (m *MockAuthService) RevokeToken(refreshToken string) error {
	args := m.Called(refreshToken)
	return args.Error(0)
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestRegister_Success(t *testing.T) {
	mockAuthService := new(MockAuthService)
	handler := NewAuthHandler(mockAuthService)
	router := setupRouter()
	router.POST("/register", handler.Register)

	user := &models.User{
		ID:       "user-123",
		Username: "testuser",
		Email:    "test@example.com",
	}

	mockAuthService.On("Register", "testuser", "password123", "test@example.com").Return(user, nil)

	reqBody := dto.RegisterRequest{
		Username: "testuser",
		Password: "password123",
		Email:    "test@example.com",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "user-123", response["user_id"])
	assert.Equal(t, "testuser", response["username"])
	assert.Equal(t, "test@example.com", response["email"])

	mockAuthService.AssertExpectations(t)
}

func TestRegister_UsernameInUse(t *testing.T) {
	mockAuthService := new(MockAuthService)
	handler := NewAuthHandler(mockAuthService)
	router := setupRouter()
	router.POST("/register", handler.Register)

	mockAuthService.On("Register", "testuser", "password123", "test@example.com").
		Return(nil, service.ErrNameInUse)

	reqBody := dto.RegisterRequest{
		Username: "testuser",
		Password: "password123",
		Email:    "test@example.com",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Account creation failed", response["error"])

	mockAuthService.AssertExpectations(t)
}

func TestRegister_EmailInUse(t *testing.T) {
	mockAuthService := new(MockAuthService)
	handler := NewAuthHandler(mockAuthService)
	router := setupRouter()
	router.POST("/register", handler.Register)

	mockAuthService.On("Register", "testuser", "password123", "test@example.com").
		Return(nil, service.ErrEmailInUse)

	reqBody := dto.RegisterRequest{
		Username: "testuser",
		Password: "password123",
		Email:    "test@example.com",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	mockAuthService.AssertExpectations(t)
}

func TestRegister_InvalidJSON(t *testing.T) {
	mockAuthService := new(MockAuthService)
	handler := NewAuthHandler(mockAuthService)
	router := setupRouter()
	router.POST("/register", handler.Register)

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogin_Success(t *testing.T) {
	mockAuthService := new(MockAuthService)
	handler := NewAuthHandler(mockAuthService)
	router := setupRouter()
	router.POST("/login", handler.Login)

	user := &models.User{ // use an already account
		ID:       "68f3b8be-5bd8-4c6c-9919-a4614b2731b3",
		Username: "manCity",
		Email:    "johndoe@example.com",
	}

	mockAuthService.On("Login", "manCity", "mcfc1213", "").
		Return("access-token", "refresh-token", user, nil)

	reqBody := dto.LoginRequest{
		Username: "manCity",
		Password: "mcfc1213",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.AuthResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "access-token", response.AccessToken)
	assert.Equal(t, "refresh-token", response.RefreshToken)
	assert.Equal(t, "68f3b8be-5bd8-4c6c-9919-a4614b2731b3", response.UserID)
	assert.Equal(t, "manCity", response.Username)
	assert.Equal(t, int64(9000), response.ExpiresIn)

	mockAuthService.AssertExpectations(t)
}

func TestLogin_InvalidCredentials(t *testing.T) {
	mockAuthService := new(MockAuthService)
	handler := NewAuthHandler(mockAuthService)
	router := setupRouter()
	router.POST("/login", handler.Login)

	mockAuthService.On("Login", "testuser", "wrongpassword", "").
		Return("", "", (*models.User)(nil), service.ErrInvalidCredentials)

	reqBody := dto.LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockAuthService.AssertExpectations(t)
}

func TestLogin_InvalidJSON(t *testing.T) {
	mockAuthService := new(MockAuthService)
	handler := NewAuthHandler(mockAuthService)
	router := setupRouter()
	router.POST("/login", handler.Login)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRefreshToken_Success(t *testing.T) {
	mockAuthService := new(MockAuthService)
	handler := NewAuthHandler(mockAuthService)
	router := setupRouter()
	router.POST("/refresh", handler.RefreshToken)

	mockAuthService.On("RefreshAccessToken", "old-refresh-token").
		Return("new-access-token", "new-refresh-token", nil)

	reqBody := dto.RefreshTokenRequest{
		RefreshToken: "old-refresh-token",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.RefreshResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "new-access-token", response.AccessToken)
	assert.Equal(t, "new-refresh-token", response.RefreshToken)
	assert.Equal(t, "Bearer", response.TokenType)
	assert.EqualValues(t, 900, response.ExpiresIn)

	mockAuthService.AssertExpectations(t)
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	mockAuthService := new(MockAuthService)
	handler := NewAuthHandler(mockAuthService)
	router := setupRouter()
	router.POST("/refresh", handler.RefreshToken)

	mockAuthService.On("RefreshAccessToken", "invalid-token").
		Return("", "", errors.New("invalid refresh token"))

	reqBody := dto.RefreshTokenRequest{
		RefreshToken: "invalid-token",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockAuthService.AssertExpectations(t)
}

func TestRefreshToken_InvalidJSON(t *testing.T) {
	mockAuthService := new(MockAuthService)
	handler := NewAuthHandler(mockAuthService)
	router := setupRouter()
	router.POST("/refresh", handler.RefreshToken)

	req, _ := http.NewRequest("POST", "/refresh", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRevokeToken_Success(t *testing.T) {
	mockAuthService := new(MockAuthService)
	handler := NewAuthHandler(mockAuthService)
	router := setupRouter()
	router.POST("/revoke", handler.RevokeToken)

	mockAuthService.On("RevokeToken", "refresh-token").Return(nil)

	reqBody := dto.RevokeTokenRequest{
		RefreshToken: "refresh-token",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/revoke", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.RevokeTokenResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Refresh token revoked successfully", response.Message)

	mockAuthService.AssertExpectations(t)
}

func TestRevokeToken_WithError(t *testing.T) {
	mockAuthService := new(MockAuthService)
	handler := NewAuthHandler(mockAuthService)
	router := setupRouter()
	router.POST("/revoke", handler.RevokeToken)

	mockAuthService.On("RevokeToken", "invalid-token").Return(errors.New("some error"))

	reqBody := dto.RevokeTokenRequest{
		RefreshToken: "invalid-token",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/revoke", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should still return success to avoid token fishing
	assert.Equal(t, http.StatusOK, w.Code)
	mockAuthService.AssertExpectations(t)
}

func TestRevokeToken_InvalidJSON(t *testing.T) {
	mockAuthService := new(MockAuthService)
	handler := NewAuthHandler(mockAuthService)
	router := setupRouter()
	router.POST("/revoke", handler.RevokeToken)

	req, _ := http.NewRequest("POST", "/revoke", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
