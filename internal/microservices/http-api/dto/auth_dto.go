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

// RegisterResponse: response payload after successful registration
type RegisterResponse struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

// OAuth2.1 DTOs
// OAuthTokenRequest: payload for OAuth2.1 token request
type OAuthTokenRequest struct {
	GrantType    string `json:"grant_type" binding:"required"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Code         string `json:"code,omitempty"`
	RedirectURI  string `json:"redirect_uri,omitempty"`
	ClientID     string `json:"client_id,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
}

// OAuthTokenResponse: response payload for OAuth2.1 token response
type OAuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type"` // e.g., "Bearer"
	ExpiresIn    int64  `json:"expires_in"` // seconds
	Scope        string `json:"scope,omitempty"`
}

// OAuthAuthorizeRequest: payload for OAuth2.1 authorization request
type OAuthAuthorizeRequest struct {
	ResponseType string `json:"response_type" binding:"required"` // e.g., "code"
	ClientID     string `json:"client_id" binding:"required"`
	RedirectURI  string `json:"redirect_uri" binding:"required"`
	Scope        string `json:"scope,omitempty"`
	State        string `json:"state,omitempty"`
}

// OAuthAuthorizeResponse: response payload for OAuth2.1 authorization response
type OAuthAuthorizeResponse struct {
	Code  string `json:"code"`
	State string `json:"state,omitempty"`
}

// OAuthClientRegistrationRequest: payload for OAuth2.1 client registration
type OAuthClientRegistrationRequest struct {
	ClientName   string   `json:"client_name" binding:"required"`
	RedirectURIs []string `json:"redirect_uris" binding:"required"`
	Scopes       []string `json:"scopes,omitempty"`
}

// OAuthClientRegistrationResponse: response payload for OAuth2.1 client registration
type OAuthClientRegistrationResponse struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURIs []string `json:"redirect_uris"`
	Scopes       []string `json:"scopes,omitempty"`
}
