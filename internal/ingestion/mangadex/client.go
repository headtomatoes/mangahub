package mangadex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/time/rate"
)

const (
	baseURL = "https://api.mangadex.org"

	// Rate limiting: MangaDex allows 5 requests per second
	rateLimit = 5
	rateBurst = 10

	// Retry configuration
	maxRetries   = 5
	initialDelay = 1 * time.Second
	maxDelay     = 32 * time.Second
)

// MangaDexClient handles API requests with rate limiting and retry logic
type MangaDexClient struct {
	baseURL     string
	apiKey      string
	httpClient  *http.Client
	rateLimiter *rate.Limiter
}

// NewClient creates a new MangaDex API client
func NewClient(apiKey string) *MangaDexClient {
	return &MangaDexClient{
		baseURL:     baseURL,
		apiKey:      apiKey,
		rateLimiter: rate.NewLimiter(rate.Limit(rateLimit), rateBurst),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

// GetManga fetches a list of manga with pagination and filters
func (c *MangaDexClient) GetManga(ctx context.Context, params url.Values) (*MangaListResponse, error) {
	endpoint := "/manga"

	// Add required includes for complete metadata
	if !params.Has("includes[]") {
		params.Add("includes[]", "author")
		params.Add("includes[]", "cover_art")
		params.Add("includes[]", "artist")
	}

	var response MangaListResponse
	err := c.doRequest(ctx, "GET", endpoint, params, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manga: %w", err)
	}

	return &response, nil
}

// GetMangaFeed fetches chapters for a specific manga
func (c *MangaDexClient) GetMangaFeed(ctx context.Context, mangaID string, params url.Values) (*ChapterListResponse, error) {
	endpoint := fmt.Sprintf("/manga/%s/feed", mangaID)

	// Default to English chapters
	if !params.Has("translatedLanguage[]") {
		params.Add("translatedLanguage[]", "en")
	}

	// Default ordering
	if !params.Has("order[chapter]") {
		params.Add("order[chapter]", "desc")
	}

	var response ChapterListResponse
	err := c.doRequest(ctx, "GET", endpoint, params, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manga feed: %w", err)
	}

	return &response, nil
}

// doRequest performs an HTTP request with rate limiting and retry logic
func (c *MangaDexClient) doRequest(ctx context.Context, method, endpoint string, params url.Values, result interface{}) error {
	// Construct full URL
	fullURL := c.baseURL + endpoint
	if params != nil {
		fullURL += "?" + params.Encode()
	}

	var lastErr error
	delay := initialDelay

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Rate limit
		if err := c.rateLimiter.Wait(ctx); err != nil {
			return fmt.Errorf("rate limiter error: %w", err)
		}

		// Create request
		req, err := http.NewRequestWithContext(ctx, method, fullURL, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		// Add headers
		req.Header.Set("User-Agent", "MangaHub/1.0")
		if c.apiKey != "" {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
		}

		// Execute request
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if attempt < maxRetries {
				log.Printf("[MangaDex] Request failed (attempt %d/%d): %v, retrying in %v...",
					attempt+1, maxRetries, err, delay)
				time.Sleep(delay)
				delay = minDuration(delay*2, maxDelay)
				continue
			}
			return fmt.Errorf("request failed after %d attempts: %w", maxRetries, err)
		}

		defer resp.Body.Close()

		// Handle HTTP errors
		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			bodyStr := string(bodyBytes)

			// Retry on rate limit or server errors
			if shouldRetry(resp.StatusCode) && attempt < maxRetries {
				lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, bodyStr)

				// Check for Retry-After header
				if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
					if retryDuration, err := time.ParseDuration(retryAfter + "s"); err == nil {
						delay = retryDuration
					}
				}

				log.Printf("[MangaDex] HTTP %d (attempt %d/%d), retrying in %v...",
					resp.StatusCode, attempt+1, maxRetries, delay)
				time.Sleep(delay)
				delay = minDuration(delay*2, maxDelay)
				continue
			}

			return fmt.Errorf("HTTP %d: %s", resp.StatusCode, bodyStr)
		}

		// Parse response
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		return nil
	}

	return fmt.Errorf("request failed after %d attempts: %w", maxRetries, lastErr)
}

// shouldRetry determines if an HTTP status code warrants a retry
func shouldRetry(statusCode int) bool {
	return statusCode == http.StatusTooManyRequests || // 429
		statusCode >= 500 // 500-504
}

// minDuration returns the smaller of two durations
func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

// BuildMangaQueryParams creates common query parameters for manga list
func BuildMangaQueryParams(limit, offset int, createdAtSince string) url.Values {
	params := url.Values{}
	params.Add("limit", fmt.Sprintf("%d", limit))
	params.Add("offset", fmt.Sprintf("%d", offset))
	params.Add("includes[]", "author")
	params.Add("includes[]", "cover_art")
	params.Add("includes[]", "artist")
	params.Add("order[createdAt]", "desc")

	if createdAtSince != "" {
		params.Add("createdAtSince", createdAtSince)
	}

	return params
}

// BuildChapterQueryParams creates common query parameters for chapter feed
func BuildChapterQueryParams(limit int, order string) url.Values {
	params := url.Values{}
	params.Add("limit", fmt.Sprintf("%d", limit))
	params.Add("translatedLanguage[]", "en")
	params.Add("order[chapter]", order) // "desc" or "asc"

	return params
}
