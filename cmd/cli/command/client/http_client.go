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
	Username string `json:"username"`
	Password string `json:"password"`
}
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

// response structure for HTTP client
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
	ExpiresIn    int64  `json:"expires_in"`
}
type RegisterResponse struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type RevokeTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RevokeTokenResponse struct {
	Message string `json:"message"`
}

// Struct for mangas and genres:
// Manga-related request/response structures
type CreateMangaRequest struct {
	Title         string  `json:"title"`
	Author        *string `json:"author,omitempty"`
	Status        *string `json:"status,omitempty"`
	TotalChapters *int    `json:"total_chapters,omitempty"`
	Description   *string `json:"description,omitempty"`
	CoverURL      *string `json:"cover_url,omitempty"`
	Slug          *string `json:"slug,omitempty"`
}

type UpdateMangaRequest struct {
	Title         *string `json:"title,omitempty"`
	Author        *string `json:"author,omitempty"`
	Status        *string `json:"status,omitempty"`
	TotalChapters *int    `json:"total_chapters,omitempty"`
	Description   *string `json:"description,omitempty"`
	CoverURL      *string `json:"cover_url,omitempty"`
	Slug          *string `json:"slug,omitempty"`
}

type MangaResponse struct {
	ID            int64      `json:"id"`
	Slug          *string    `json:"slug,omitempty"`
	Title         string     `json:"title"`
	Author        *string    `json:"author,omitempty"`
	Status        *string    `json:"status,omitempty"`
	TotalChapters *int       `json:"total_chapters,omitempty"`
	Description   *string    `json:"description,omitempty"`
	CoverURL      *string    `json:"cover_url,omitempty"`
	CreatedAt     *time.Time `json:"created_at,omitempty"`
}

type GenreResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type GenreIDsRequest struct {
	GenreIDs []int64 `json:"genre_ids"`
}

// Library-related request/response structures
type AddToLibraryRequest struct {
	MangaID int64 `json:"manga_id"`
}

type LibraryItemResponse struct {
	ID      int64         `json:"id"`
	MangaID int64         `json:"manga_id"`
	Manga   MangaResponse `json:"manga"`
	AddedAt time.Time     `json:"added_at"`
}

type LibraryListResponse struct {
	Items []LibraryItemResponse `json:"items"`
	Total int                   `json:"total"`
}

type PaginatedMangaResponse struct {
	Data       []MangaResponse `json:"data"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	Total      int64           `json:"total"`
	TotalPages int             `json:"total_pages"`
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
func (c *HTTPClient) Register(request *RegisterRequest) (*RegisterResponse, error) {
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

	var result RegisterResponse
	// if decoding the response fails, return error
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// refresh token method for HTTP client
func (c *HTTPClient) RefreshToken(request *RefreshTokenRequest) (*RefreshResponse, error) {
	body := map[string]string{"refresh_token": request.RefreshToken} // prepare request body
	jsonData, err := json.Marshal(body)                              // marshal to JSON
	if err != nil {
		fmt.Println("Failed to marshal refresh token request:", err)
		return nil, err
	}

	resp, err := c.httpClient.Post(c.baseURL+"/auth/refresh", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Failed to send refresh token request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Refresh token request failed with status:", resp.Status)
		return nil, fmt.Errorf("refresh failed: %s", resp.Status)
	}

	var result RefreshResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("Failed to decode refresh token response:", err)
		return nil, err
	}

	return &result, nil
}

// revoke token method for HTTP client
func (c *HTTPClient) RevokeToken(request *RevokeTokenRequest) (*RevokeTokenResponse, error) {
	body := map[string]string{"refresh_token": request.RefreshToken}
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Post(c.baseURL+"/auth/revoke", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("revoke failed: %s", resp.Status)
	}
	var result RevokeTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// add more methods as needed and link to actual API endpoints
// Manga CURD
func (c *HTTPClient) GetAllManga() ([]MangaResponse, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/manga", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get manga list: %s", resp.Status)
	}

	var paginatedResult PaginatedMangaResponse
	if err := json.NewDecoder(resp.Body).Decode(&paginatedResult); err != nil {
		return nil, err
	}
	return paginatedResult.Data, nil
}

func (c *HTTPClient) GetMangaByID(id int64) (*MangaResponse, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/manga/%d", c.baseURL, id), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("manga not found: %s", resp.Status)
	}

	var result MangaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *HTTPClient) SearchManga(query string) ([]MangaResponse, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/manga/search?q=%s", c.baseURL, query), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed: %s", resp.Status)
	}

	var result []MangaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *HTTPClient) CreateManga(request *CreateMangaRequest) (*MangaResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/manga", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create manga: %s", resp.Status)
	}

	var result MangaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *HTTPClient) UpdateManga(id int64, request *UpdateMangaRequest) (*MangaResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/api/manga/%d", c.baseURL, id), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to update manga: %s", resp.Status)
	}

	var result MangaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *HTTPClient) DeleteManga(id int64) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/manga/%d", c.baseURL, id), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to delete manga: %s", resp.Status)
	}
	return nil
}

func (c *HTTPClient) GetMangaGenres(mangaID int64) ([]GenreResponse, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/manga/%d/genres", c.baseURL, mangaID), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get genres: %s", resp.Status)
	}

	var result []GenreResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *HTTPClient) AddMangaGenres(mangaID int64, genreIDs []int64) error {
	request := GenreIDsRequest{GenreIDs: genreIDs}
	jsonData, err := json.Marshal(request)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/manga/%d/genres", c.baseURL, mangaID), bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to add genres: %s", resp.Status)
	}
	return nil
}

func (c *HTTPClient) RemoveMangaGenres(mangaID int64, genreIDs []int64) error {
	request := GenreIDsRequest{GenreIDs: genreIDs}
	jsonData, err := json.Marshal(request)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/manga/%d/genres", c.baseURL, mangaID), bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to remove genres: %s", resp.Status)
	}
	return nil
}

// Library methods
func (c *HTTPClient) AddToLibrary(mangaID int64) error {
	request := AddToLibraryRequest{MangaID: mangaID}
	jsonData, err := json.Marshal(request)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/library", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return fmt.Errorf("manga already in library")
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to add to library: %s", resp.Status)
	}

	return nil
}

func (c *HTTPClient) GetLibrary() (*LibraryListResponse, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/api/library", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get library: %s", resp.Status)
	}

	var result LibraryListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *HTTPClient) RemoveFromLibrary(mangaID int64) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/library/%d", c.baseURL, mangaID), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to remove from library: %s", resp.Status)
	}

	return nil
}

// Progress-related structures
type UpdateProgressRequest struct {
	UserID  string `json:"user_id"`
	MangaID int64  `json:"manga_id"`
	Chapter int    `json:"chapter"`
	Status  string `json:"status"`
	Volume  *int   `json:"volume,omitempty"`
	Notes   string `json:"notes,omitempty"`
}

type ProgressResponse struct {
	UserID          string    `json:"user_id"`
	MangaID         int64     `json:"manga_id"`
	MangaTitle      string    `json:"manga_title,omitempty"`
	Chapter         int       `json:"chapter"`
	Volume          *int      `json:"volume,omitempty"`
	Status          string    `json:"status"`
	Notes           string    `json:"notes,omitempty"`
	PreviousChapter int       `json:"previous_chapter,omitempty"`
	UpdatedAt       time.Time `json:"updated_at"`
	TotalChapters   *int      `json:"total_chapters,omitempty"`
	ReadingStreak   int       `json:"reading_streak,omitempty"`
}

type ProgressHistoryResponse struct {
	History []ProgressResponse `json:"history"`
	Total   int                `json:"total"`
}

// Progress methods
func (c *HTTPClient) GetProgress(mangaID int64) (*ProgressResponse, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/progress/%d", c.baseURL, mangaID), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("progress not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get progress: %s", resp.Status)
	}

	var result ProgressResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *HTTPClient) UpdateProgress(request *UpdateProgressRequest) (*ProgressResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/progress/%d", c.baseURL, request.MangaID), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to update progress: %s", resp.Status)
	}

	var result ProgressResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *HTTPClient) GetProgressHistory(mangaID *int64) (*ProgressHistoryResponse, error) {
	url := c.baseURL + "/api/progress/history"
	if mangaID != nil {
		url = fmt.Sprintf("%s?manga_id=%d", url, *mangaID)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get progress history: %s", resp.Status)
	}

	var result ProgressHistoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}
