package handler

import (
	"fmt"
	"mangahub/internal/microservices/http-api/dto"
	"mangahub/internal/microservices/http-api/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.Register(req.Username, req.Password, req.Email)
	if err == service.ErrNameInUse || err == service.ErrEmailInUse {
		c.JSON(http.StatusConflict, gin.H{"error": "Account creation failed"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, refreshToken, user, err := h.authService.Login(req.Username, req.Password, req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserID:       user.ID,
		Username:     user.Username,
		ExpiresIn:    900, // 15 minutes in seconds
	})
}

// TODO: what is the best practice for refresh token response structure?
// always return new RT along with new AT
// rotate both tokens to enhance security
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newAccessToken, newRefreshToken, err := h.authService.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.RefreshResponse{
		RefreshToken: newRefreshToken,
		AccessToken:  newAccessToken,
		TokenType:    "Bearer",
		ExpiresIn:    900, // 15 minutes in seconds
	})
}

func (h *AuthHandler) RevokeToken(c *gin.Context) {
	var req dto.RevokeTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.authService.RevokeToken(req.RefreshToken)
	if err != nil {
		fmt.Println("some error occurred:", err)
	}

	// always return success response to avoid token fishing
	c.JSON(http.StatusOK, dto.RevokeTokenResponse{
		Message: "Refresh token revoked successfully",
	})
}
