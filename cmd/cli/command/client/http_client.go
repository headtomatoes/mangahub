package client

// http_client.go = handles HTTP client functionality for the mangahubCLI application.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// defines the HTTP client structure and methods
type HTTPClient struct {
	// fields for HTTP client configuration
	baseURL    string
	httpClient *http.Client
	token      string
}

// request structure for HTTP client
type LoginRequest struct {
	Username string
	Password string
}
type RegisterRequest struct {
	Username string
	Password string
	Email    string
}

// response structure for HTTP client
type AuthResponse struct {
	AccessToken  string
	RefreshToken string
	UserID       string
	Username     string
	ExpiresIn    int64
}

// constructor for HTTP client
func NewHTTPClient(apiURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: apiURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// set token for HTTP client
func (c *HTTPClient) SetToken(token string) {
	c.token = token
}

// login method for HTTP client
func (c *HTTPClient) Login(request *LoginRequest) (*AuthResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	response, err := c.httpClient.Post(c.baseURL+"/auth/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close() // Ensure the response body is closed

	// check for non-200 status code => login failed
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login failed with status: %s", response.Status)
	}

	var result AuthResponse

	// if decoding the response fails, return error
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// register method for HTTP client
func (c *HTTPClient) Register(request *RegisterRequest) (*AuthResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	response, err := c.httpClient.Post(c.baseURL+"/auth/register", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close() // Ensure the response body is closed

	// check for non-201 status code => registration failed
	if response.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("registration failed with status: %s", response.Status)
	}

	var result AuthResponse

	// if decoding the response fails, return error
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// add more methods as needed and link to actual API endpoints
