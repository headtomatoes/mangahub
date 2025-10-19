package dto

// Data Transfer Objects for authentication requests and responses

// RegisterRequest: payload for user registration
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=8"`
	Email    string `json:"email" binding:"required,email"`
}

// LoginRequest: payload for user login
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse: response payload after successful authentication
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
	ExpiresIn    int64  `json:"expires_in"` // seconds
}

// RefreshTokenRequest: payload for refreshing access token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshResponse: response payload after refreshing access token
type RefreshResponse struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
