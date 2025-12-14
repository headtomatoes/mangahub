package anilist

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "time"

    "golang.org/x/time/rate"
)

const (
    apiURL = "https://graphql.anilist.co"

    // Rate limiting: AniList allows ~90 requests per minute
    rateLimit = 1 // 1 requests per second = 60/min
    rateBurst = 5

    // Retry configuration
    maxRetries   = 5
    initialDelay = 1 * time.Second
    maxDelay     = 32 * time.Second
)

// AniListClient handles GraphQL API requests with rate limiting
type AniListClient struct {
    apiURL      string
    httpClient  *http.Client
    rateLimiter *rate.Limiter
}

// NewClient creates a new AniList API client
func NewClient() *AniListClient {
    return &AniListClient{
        apiURL:      apiURL,
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

// GraphQLRequest represents a GraphQL query request
type GraphQLRequest struct {
    Query     string                 `json:"query"`
    Variables map[string]interface{} `json:"variables,omitempty"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
    Data   json.RawMessage `json:"data"`
    Errors []GraphQLError  `json:"errors,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
    Message   string                 `json:"message"`
    Locations []GraphQLErrorLocation `json:"locations,omitempty"`
}

// GraphQLErrorLocation represents error location
type GraphQLErrorLocation struct {
    Line   int `json:"line"`
    Column int `json:"column"`
}

// GetManga fetches manga list with pagination
func (c *AniListClient) GetManga(ctx context.Context, page, perPage int) (*MangaPageResponse, error) {
    query := `
    query ($page: Int, $perPage: Int) {
        Page(page: $page, perPage: $perPage) {
            pageInfo {
                total
                currentPage
                lastPage
                hasNextPage
                perPage
            }
            media(type: MANGA, sort: POPULARITY_DESC) {
                id
                idMal
                title {
                    english
                    romaji
                    native
                }
                description
                status
                chapters
                volumes
                coverImage {
                    large
                    medium
                }
                bannerImage
                genres
                tags {
                    name
                    rank
                }
                averageScore
                staff {
                    edges {
                        role
                        node {
                            name {
                                full
                            }
                        }
                    }
                }
                startDate {
                    year
                    month
                    day
                }
                endDate {
                    year
                    month
                    day
                }
                updatedAt
            }
        }
    }
    `

    variables := map[string]interface{}{
        "page":    page,
        "perPage": perPage,
    }

    var result MangaPageResponse
    err := c.doRequest(ctx, query, variables, &result)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch manga: %w", err)
    }

    return &result, nil
}

// GetMangaByID fetches a specific manga by ID
func (c *AniListClient) GetMangaByID(ctx context.Context, id int) (*MediaResponse, error) {
    query := `
    query ($id: Int) {
        Media(id: $id, type: MANGA) {
            id
            idMal
            title {
                english
                romaji
                native
            }
            description
            status
            chapters
            volumes
            coverImage {
                large
                medium
            }
            bannerImage
            genres
            tags {
                name
                rank
            }
            averageScore
            staff {
                edges {
                    role
                    node {
                        name {
                            full
                        }
                    }
                }
            }
            startDate {
                year
                month
                day
            }
            updatedAt
        }
    }
    `

    variables := map[string]interface{}{
        "id": id,
    }

    var result struct {
        Media MediaData `json:"Media"`
    }

    err := c.doRequest(ctx, query, variables, &result)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch manga by ID: %w", err)
    }

    return &MediaResponse{Media: result.Media}, nil
}

// GetRecentlyUpdated fetches manga updated after a specific timestamp
func (c *AniListClient) GetRecentlyUpdated(ctx context.Context, updatedAfter int64, page, perPage int) (*MangaPageResponse, error) {
    query := `
    query ($page: Int, $perPage: Int, $updatedAfter: Int) {
        Page(page: $page, perPage: $perPage) {
            pageInfo {
                total
                currentPage
                lastPage
                hasNextPage
                perPage
            }
            media(type: MANGA, sort: UPDATED_AT_DESC) {
                id
                idMal
                title {
                    english
                    romaji
                    native
                }
                description
                status
                chapters
                volumes
                coverImage {
                    large
                    medium
                }
                bannerImage
                genres
                tags {
                    name
                    rank
                }
                averageScore
                staff {
                    edges {
                        role
                        node {
                            name {
                                full
                            }
                        }
                    }
                }
                startDate {
                    year
                    month
                    day
                }
                updatedAt
            }
        }
    }
    `

    variables := map[string]interface{}{
        "page":         page,
        "perPage":      perPage,
        "updatedAfter": updatedAfter,
    }

    var result MangaPageResponse
    err := c.doRequest(ctx, query, variables, &result)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch recently updated manga: %w", err)
    }

    return &result, nil
}

// doRequest performs a GraphQL request with rate limiting and retry logic
func (c *AniListClient) doRequest(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
    // Prepare request body
    reqBody := GraphQLRequest{
        Query:     query,
        Variables: variables,
    }

    bodyJSON, err := json.Marshal(reqBody)
    if err != nil {
        return fmt.Errorf("failed to marshal request: %w", err)
    }

    var lastErr error
    delay := initialDelay

    for attempt := 0; attempt <= maxRetries; attempt++ {
        // Rate limit
        if err := c.rateLimiter.Wait(ctx); err != nil {
            return fmt.Errorf("rate limiter error: %w", err)
        }

        // Create request
        req, err := http.NewRequestWithContext(ctx, "POST", c.apiURL, bytes.NewReader(bodyJSON))
        if err != nil {
            return fmt.Errorf("failed to create request: %w", err)
        }

        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("Accept", "application/json")

        // Execute request
        resp, err := c.httpClient.Do(req)
        if err != nil {
            lastErr = err
            if attempt < maxRetries {
                log.Printf("[AniList] Request failed (attempt %d/%d): %v, retrying in %v...",
                    attempt+1, maxRetries, err, delay)
                time.Sleep(delay)
                delay = minDuration(delay*2, maxDelay)
                continue
            }
            return fmt.Errorf("request failed after %d attempts: %w", maxRetries, err)
        }

        defer resp.Body.Close()

        // Read response body
        respBody, err := io.ReadAll(resp.Body)
        if err != nil {
            return fmt.Errorf("failed to read response: %w", err)
        }

        // Handle HTTP errors
        if resp.StatusCode != http.StatusOK {
            // Retry on rate limit or server errors
            if shouldRetry(resp.StatusCode) && attempt < maxRetries {
                lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))

                // Check for Retry-After header
                if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
                    if retryDuration, err := time.ParseDuration(retryAfter + "s"); err == nil {
                        delay = retryDuration
                    }
                }

                log.Printf("[AniList] HTTP %d (attempt %d/%d), retrying in %v...",
                    resp.StatusCode, attempt+1, maxRetries, delay)
                time.Sleep(delay)
                delay = minDuration(delay*2, maxDelay)
                continue
            }

            return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
        }

        // Parse GraphQL response
        var gqlResp GraphQLResponse
        if err := json.Unmarshal(respBody, &gqlResp); err != nil {
            return fmt.Errorf("failed to parse GraphQL response: %w", err)
        }

        // Check for GraphQL errors
        if len(gqlResp.Errors) > 0 {
            errMsgs := make([]string, len(gqlResp.Errors))
            for i, e := range gqlResp.Errors {
                errMsgs[i] = e.Message
            }
            return fmt.Errorf("GraphQL errors: %v", errMsgs)
        }

        // Parse data into result
        if err := json.Unmarshal(gqlResp.Data, result); err != nil {
            return fmt.Errorf("failed to parse data: %w", err)
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