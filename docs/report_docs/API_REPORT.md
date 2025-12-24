# HTTP API Use Cases Documentation

## 1. Overview

The MangaHub HTTP API is a RESTful service built with **Go (Golang)** and the **Gin web framework**. It provides comprehensive endpoints for managing manga content, user interactions, and real-time features. The API follows modern best practices including JWT authentication, role-based access control, and pagination.

**Technology Stack**:
- **Language**: Go 1.21+
- **Web Framework**: Gin
- **ORM**: GORM
- **Database**: PostgreSQL
- **Authentication**: JWT (Access + Refresh Tokens)
- **Real-time**: WebSocket (Gorilla WebSocket)

---

## 2. Architecture Overview

### 2.1 Layered Architecture

```
┌─────────────────────────────────────┐
│         HTTP Layer (Gin)            │
│     Routes, Middleware, CORS        │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│       Handler Layer                 │
│  Request Validation, Response       │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│       Service Layer                 │
│  Business Logic, Validation         │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│      Repository Layer               │
│    Database Queries (GORM)          │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│         Database (PostgreSQL)       │
└─────────────────────────────────────┘
```

**Benefits**:
- ✅ **Separation of Concerns**: Each layer has a single responsibility
- ✅ **Testability**: Easy to mock dependencies for unit testing
- ✅ **Maintainability**: Changes in one layer don't affect others
- ✅ **Scalability**: Layers can be optimized independently

---

## 3. Authentication System

### 3.1 JWT Token-Based Authentication

#### Architecture

```go
type AuthService interface {
    Register(username, password, email string) (*models.User, error)
    Login(username, password, email string) (accessToken, refreshToken string, user *models.User, error)
    RefreshAccessToken(refreshToken string) (newAccessToken, newRefreshToken string, error)
    RevokeToken(refreshToken string) error
}
```

#### Token Strategy

**Access Token**:
- **Lifetime**: 15 minutes
- **Purpose**: Authorize API requests
- **Claims**: `user_id`, `username`, `role`, `scopes`
- **Storage**: Client-side (memory/localStorage)

**Refresh Token**:
- **Lifetime**: 7 days
- **Purpose**: Generate new access tokens
- **Storage**: Database with revocation support
- **Rotation**: New refresh token issued on each refresh

#### Implementation Example

```go
// filepath: internal/microservices/http-api/service/authService.go (excerpt)

func (s *authService) Login(username, password, email string) (string, string, *models.User, error) {
    // 1. Find user by username or email
    user, err := s.userRepo.FindByUsernameOrEmail(username, email)
    if err != nil {
        return "", "", nil, ErrInvalidCredentials
    }

    // 2. Verify password hash (bcrypt)
    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
        return "", "", nil, ErrInvalidCredentials
    }

    // 3. Generate access token (15 min)
    accessToken, err := s.generateAccessToken(user)
    if err != nil {
        return "", "", nil, err
    }

    // 4. Generate refresh token (7 days) and store in DB
    refreshToken, err := s.generateAndStoreRefreshToken(user.ID)
    if err != nil {
        return "", "", nil, err
    }

    // 5. Update last login timestamp
    user.LastLogin = time.Now()
    s.userRepo.Update(user)

    return accessToken, refreshToken, user, nil
}
```

#### Security Features

**Password Security**:
- **Hashing Algorithm**: bcrypt (cost factor: 10)
- **Salt**: Automatically generated per password
- **Protection**: Against rainbow table attacks

**Token Revocation**:
```go
func (s *authService) RevokeToken(tokenString string) error {
    // Mark token as revoked in database
    return s.refreshTokenRepo.RevokeToken(tokenString)
}
```

**Refresh Token Rotation**:
```go
func (s *authService) RefreshAccessToken(oldRefreshToken string) (string, string, error) {
    // 1. Validate old refresh token
    claims, err := s.validateRefreshToken(oldRefreshToken)
    if err != nil {
        return "", "", err
    }

    // 2. Check if token is revoked
    isRevoked, err := s.refreshTokenRepo.IsRevoked(oldRefreshToken)
    if err != nil || isRevoked {
        return "", "", ErrInvalidToken
    }

    // 3. Revoke old refresh token
    s.refreshTokenRepo.RevokeToken(oldRefreshToken)

    // 4. Generate new tokens
    newAccessToken, _ := s.generateAccessToken(claims.UserID)
    newRefreshToken, _ := s.generateAndStoreRefreshToken(claims.UserID)

    return newAccessToken, newRefreshToken, nil
}
```

---

### 3.2 Middleware Authentication

```go
// filepath: internal/microservices/http-api/middleware/auth.go

func AuthMiddleware(authService service.AuthService) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Extract token from Authorization header
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(401, gin.H{"error": "authorization header required"})
            c.Abort()
            return
        }

        // 2. Parse Bearer token
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        
        // 3. Validate JWT signature and expiration
        claims, err := authService.ValidateAccessToken(tokenString)
        if err != nil {
            c.JSON(401, gin.H{"error": "invalid token"})
            c.Abort()
            return
        }

        // 4. Store user info in context for handlers
        c.Set("userID", claims.UserID)
        c.Set("username", claims.Username)
        c.Set("role", claims.Role)
        c.Set("scopes", claims.Scopes)

        c.Next()
    }
}
```

#### Scope-Based Authorization

```go
func RequireScopes(requiredScopes ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userScopes, exists := c.Get("scopes")
        if !exists {
            c.JSON(403, gin.H{"error": "insufficient permissions"})
            c.Abort()
            return
        }

        scopes := userScopes.([]string)
        for _, required := range requiredScopes {
            if !contains(scopes, required) {
                c.JSON(403, gin.H{"error": "missing required scope: " + required})
                c.Abort()
                return
            }
        }

        c.Next()
    }
}
```

#### Role-Based Authorization

```go
func RequireAdmin() gin.HandlerFunc {
    return func(c *gin.Context) {
        role, exists := c.Get("role")
        if !exists || role != "admin" {
            c.JSON(403, gin.H{"error": "admin access required"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

---

## 4. Core Use Cases

### 4.1 Manga Management

#### Use Case 1: Create Manga (Admin Only)

**Endpoint**: `POST /api/manga`

**Authentication**: Required (Admin role)

**Request**:
```json
{
    "title": "One Piece",
    "author": "Eiichiro Oda",
    "status": "ongoing",
    "total_chapters": 1100,
    "description": "Follow Monkey D. Luffy's quest to become Pirate King",
    "cover_url": "https://cdn.example.com/one-piece-cover.jpg",
    "genre_ids": [1, 5, 8]
}
```

**Handler Implementation**:
```go
func (h *MangaHandler) Create(c *gin.Context) {
    var in dto.CreateMangaDTO
    if err := c.ShouldBindJSON(&in); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // Convert DTO to model
    model := in.ToModel()
    
    ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
    defer cancel()

    // Create manga (generates slug automatically)
    if err := h.svc.Create(ctx, &model); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // Assign genres if provided
    if len(in.GenreIDs) > 0 {
        if err := h.svc.ReplaceGenresForManga(ctx, model.ID, in.GenreIDs); err != nil {
            c.JSON(201, gin.H{
                "manga": dto.FromModelToResponse(model),
                "warning": "Failed to assign some genres: " + err.Error(),
            })
            return
        }
    }

    // Return created manga with genres
    created, _ := h.svc.GetByID(ctx, model.ID)
    c.JSON(201, dto.FromModelToResponse(*created))
}
```

**Service Layer Logic**:
```go
func (s *mangaService) Create(ctx context.Context, m *models.Manga) error {
    // 1. Validate title
    if strings.TrimSpace(m.Title) == "" {
        return errors.New("title is required")
    }

    // 2. Generate slug from title if not provided
    if m.Slug == nil || strings.TrimSpace(*m.Slug) == "" {
        slug := generateSlug(m.Title)
        // Add UUID suffix to avoid collisions
        slug = fmt.Sprintf("%s-%s", slug, uuid.New().String()[:8])
        m.Slug = &slug
    }

    // 3. Normalize author name
    if m.Author != nil {
        author := strings.TrimSpace(*m.Author)
        m.Author = &author
    }

    // 4. Persist to database
    if err := s.repo.Create(ctx, m); err != nil {
        return err
    }

    // 5. Send notification to UDP server (non-blocking)
    go notifyNewManga(m.ID, m.Title)

    return nil
}
```

**Response**:
```json
{
    "id": 1,
    "slug": "one-piece-a1b2c3d4",
    "title": "One Piece",
    "author": "Eiichiro Oda",
    "status": "ongoing",
    "total_chapters": 1100,
    "description": "Follow Monkey D. Luffy's quest to become Pirate King",
    "cover_url": "https://cdn.example.com/one-piece-cover.jpg",
    "average_rating": null,
    "created_at": "2025-12-23T10:30:00Z",
    "genres": [
        {"id": 1, "name": "Action"},
        {"id": 5, "name": "Adventure"},
        {"id": 8, "name": "Fantasy"}
    ]
}
```

**Business Rules**:
- ✅ Slug auto-generation with collision prevention
- ✅ Real-time notification via UDP for new manga
- ✅ Genre assignment with transaction safety
- ✅ Admin-only access control

---

#### Use Case 2: Search Manga (Advanced)

**Endpoint**: `GET /api/manga/advanced-search`

**Query Parameters**:
- `q` - Full-text search query
- `status` - Filter by status (ongoing/completed/hiatus)
- `genres` - Comma-separated genre IDs or names
- `min_rating` - Minimum average rating (0-10)
- `sort_by` - Sort order (popularity/rating/recent/title)
- `page` - Page number (default: 1)
- `page_size` - Items per page (default: 20, max: 100)

**Example Request**:
```
GET /api/manga/advanced-search?q=pirates&genres=action,adventure&status=ongoing&min_rating=8.0&sort_by=popularity&page=1&page_size=20
```

**Handler Implementation**:
```go
func (h *MangaHandler) AdvancedSearch(c *gin.Context) {
    var filters dto.SearchFilters

    // Parse query parameters with sanitization
    filters.Query = strings.TrimSpace(c.Query("q"))
    filters.Status = strings.TrimSpace(c.Query("status"))
    filters.SortBy = strings.TrimSpace(c.Query("sort_by"))

    // Parse genres (comma-separated)
    if genresStr := c.Query("genres"); genresStr != "" {
        filters.Genres = strings.Split(genresStr, ",")
    }

    // Parse min_rating
    if minRatingStr := c.Query("min_rating"); minRatingStr != "" {
        if minRating, err := strconv.ParseFloat(minRatingStr, 64); err == nil {
            filters.MinRating = &minRating
        }
    }

    // Parse pagination
    filters.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
    filters.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))

    // Execute search
    list, total, err := h.svc.AdvancedSearch(ctx, filters)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    // Return paginated response
    c.JSON(200, gin.H{
        "data": list,
        "pagination": gin.H{
            "page": filters.Page,
            "page_size": filters.PageSize,
            "total": total,
            "total_pages": (total + int64(filters.PageSize) - 1) / int64(filters.PageSize),
        },
    })
}
```

**Repository Query Logic**:
```go
func (r *MangaRepo) AdvancedSearch(ctx context.Context, filters dto.SearchFilters) ([]models.Manga, int64, error) {
    db := r.db.WithContext(ctx).Model(&models.Manga{})

    // 1. Full-text search on title, author, description, slug
    if filters.Query != "" {
        tokens := strings.Fields(filters.Query)
        clauses := make([]string, 0)
        args := make([]interface{}, 0)
        
        for _, token := range tokens {
            pattern := "%" + token + "%"
            clauses = append(clauses, 
                "(title ILIKE ? OR author ILIKE ? OR description ILIKE ? OR slug ILIKE ?)")
            args = append(args, pattern, pattern, pattern, pattern)
        }
        
        db = db.Where(strings.Join(clauses, " AND "), args...)
    }

    // 2. Filter by status
    if filters.Status != "" {
        db = db.Where("LOWER(status) = LOWER(?)", filters.Status)
    }

    // 3. Filter by minimum rating
    if filters.MinRating != nil {
        db = db.Where("average_rating >= ?", *filters.MinRating)
    }

    // 4. Filter by genres (many-to-many)
    if len(filters.Genres) > 0 {
        db = db.Joins("JOIN manga_genres ON manga_genres.manga_id = manga.id").
            Joins("JOIN genres ON genres.id = manga_genres.genre_id").
            Where("LOWER(genres.name) IN (?)", filters.Genres).
            Group("manga.id").
            Having("COUNT(DISTINCT genres.id) >= ?", len(filters.Genres))
    }

    // 5. Count total results
    var total int64
    db.Count(&total)

    // 6. Apply sorting
    switch filters.SortBy {
    case "popularity", "rating":
        db = db.Order("average_rating DESC NULLS LAST")
    case "recent":
        db = db.Order("created_at DESC")
    case "title":
        db = db.Order("title ASC")
    default:
        db = db.Order("created_at DESC")
    }

    // 7. Apply pagination
    offset := (filters.Page - 1) * filters.PageSize
    var results []models.Manga
    db.Limit(filters.PageSize).Offset(offset).Find(&results)

    return results, total, nil
}
```

**Response**:
```json
{
    "data": [
        {
            "id": 1,
            "slug": "one-piece-a1b2c3d4",
            "title": "One Piece",
            "author": "Eiichiro Oda",
            "status": "ongoing",
            "total_chapters": 1100,
            "cover_url": "https://cdn.example.com/one-piece-cover.jpg",
            "average_rating": 9.2
        }
    ],
    "pagination": {
        "page": 1,
        "page_size": 20,
        "total": 1,
        "total_pages": 1,
        "has_next": false,
        "has_previous": false
    },
    "filters": {
        "query": "pirates",
        "genres": ["action", "adventure"],
        "status": "ongoing",
        "min_rating": 8.0,
        "sort_by": "popularity"
    }
}
```

**Performance Optimizations**:
- ✅ **Indexed columns**: title, author, status, average_rating
- ✅ **Composite index**: (manga_id, genre_id) for genre joins
- ✅ **Pagination**: Limits database load
- ✅ **ILIKE optimization**: PostgreSQL full-text search ready

---

### 4.2 User Library Management

#### Use Case 3: Add Manga to Library

**Endpoint**: `POST /api/library`

**Authentication**: Required

**Request**:
```json
{
    "manga_id": 1
}
```

**Service Implementation**:
```go
func (s *libraryService) Add(ctx context.Context, userID string, mangaID int64) error {
    // 1. Verify manga exists
    if _, err := s.mangaRepo.GetByID(ctx, mangaID); err != nil {
        return errors.New("manga not found")
    }
    
    // 2. Check if already in library (prevent duplicates)
    exists, err := s.repo.Exists(ctx, userID, mangaID)
    if err != nil {
        return err
    }
    if exists {
        return ErrAlreadyInLibrary // HTTP 409 Conflict
    }
    
    // 3. Add to library
    return s.repo.Add(ctx, userID, mangaID)
}
```

**Response** (201 Created):
```json
{
    "message": "manga added to library"
}
```

---

#### Use Case 4: Get User's Library

**Endpoint**: `GET /api/library`

**Response**:
```json
{
    "items": [
        {
            "id": 1,
            "manga_id": 1,
            "added_at": "2025-12-23T10:30:00Z",
            "manga": {
                "id": 1,
                "title": "One Piece",
                "author": "Eiichiro Oda",
                "cover_url": "https://cdn.example.com/one-piece-cover.jpg",
                "average_rating": 9.2
            }
        }
    ],
    "total": 1
}
```

**Repository Query**:
```go
func (r *libraryRepository) List(ctx context.Context, userID string) ([]models.UserLibrary, error) {
    var library []models.UserLibrary
    
    err := r.db.WithContext(ctx).
        Preload("Manga").               // Eager load manga details
        Where("user_id = ?", userID).
        Order("added_at DESC").         // Most recent first
        Find(&library).Error
    
    return library, err
}
```

---

### 4.3 Rating System

#### Use Case 5: Rate Manga

**Endpoint**: `POST /api/manga/:manga_id/ratings`

**Request**:
```json
{
    "rating": 9
}
```

**Service Logic with Average Calculation**:
```go
func (s *ratingService) CreateOrUpdateRating(userID string, mangaID int64, ratingValue int) (*dto.RatingResponse, error) {
    // 1. Verify manga exists
    _, err := s.mangaRepo.GetByID(ctx, mangaID)
    if err != nil {
        return nil, errors.New("manga not found")
    }

    // 2. Check if user already rated this manga
    existingRating, err := s.ratingRepo.GetByUserAndManga(userID, mangaID)

    if existingRating != nil {
        // Update existing rating
        existingRating.Rating = ratingValue
        s.ratingRepo.Update(existingRating)
    } else {
        // Create new rating
        newRating := &models.Rating{
            UserID: userID,
            MangaID: mangaID,
            Rating: ratingValue,
        }
        s.ratingRepo.Create(newRating)
    }

    // 3. Recalculate manga's average rating
    avg, err := s.ratingRepo.CalculateAverageRating(mangaID)
    if err == nil {
        // Update manga table
        manga, _ := s.mangaRepo.GetByID(ctx, mangaID)
        manga.AverageRating = &avg
        s.mangaRepo.Update(ctx, mangaID, manga)
    }

    return rating, nil
}
```

**Average Rating Calculation**:
```go
func (r *ratingRepository) CalculateAverageRating(mangaID int64) (float64, error) {
    var result struct {
        Average float64
    }

    err := r.db.Model(&models.Rating{}).
        Select("COALESCE(AVG(rating), 0) as average").
        Where("manga_id = ?", mangaID).
        Scan(&result).Error

    return result.Average, err
}
```

**Response**:
```json
{
    "username": "john_doe",
    "rating": 9,
    "created_at": "2025-12-23T10:30:00Z",
    "updated_at": "2025-12-23T10:30:00Z"
}
```

---

### 4.4 Comment System

#### Use Case 6: Post Comment

**Endpoint**: `POST /api/manga/:manga_id/comments`

**Request**:
```json
{
    "content": "This manga is amazing! Best story I've read in years."
}
```

**Handler**:
```go
func (h *CommentHandler) Create(c *gin.Context) {
    mangaID, _ := strconv.ParseInt(c.Param("manga_id"), 10, 64)
    userID, _ := c.Get("userID")

    var req dto.CreateCommentDTO
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    comment, err := h.commentService.CreateComment(
        userID.(string), 
        mangaID, 
        req.Content,
    )
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(201, comment)
}
```

**Response**:
```json
{
    "username": "john_doe",
    "content": "This manga is amazing! Best story I've read in years.",
    "created_at": "2025-12-23T10:30:00Z",
    "updated_at": "2025-12-23T10:30:00Z"
}
```

---

#### Use Case 7: Get Comments (Paginated)

**Endpoint**: `GET /api/manga/:manga_id/comments?page=1&page_size=20`

**Service Logic**:
```go
func (s *commentService) GetMangaComments(mangaID int64, page, pageSize int) (*dto.PaginatedCommentResponse, error) {
    // 1. Verify manga exists
    _, err := s.mangaRepo.GetByID(ctx, mangaID)
    if err != nil {
        return nil, errors.New("manga not found")
    }

    // 2. Get paginated comments
    comments, total, err := s.commentRepo.GetByManga(mangaID, page, pageSize)
    if err != nil {
        return nil, err
    }

    // 3. Convert to response DTOs
    commentResponses := make([]dto.CommentResponse, len(comments))
    for i, comment := range comments {
        commentResponses[i] = *dto.FromModelToCommentResponse(&comment)
    }

    return dto.NewPaginatedCommentResponse(commentResponses, int(total), page, pageSize), nil
}
```

**Response**:
```json
{
    "data": [
        {
            "username": "john_doe",
            "content": "This manga is amazing!",
            "created_at": "2025-12-23T10:30:00Z",
            "updated_at": "2025-12-23T10:30:00Z"
        }
    ],
    "page": 1,
    "page_size": 20,
    "total": 1,
    "total_pages": 1
}
```

---

### 4.5 Reading Progress Tracking

#### Use Case 8: Update Reading Progress

**Endpoint**: `POST /api/progress/:manga_id`

**Request**:
```json
{
    "manga_id": 1,
    "current_chapter": 105,
    "status": "reading",
    "page": 15
}
```

**Service Implementation**:
```go
func (s *progressService) UpdateProgress(req *dto.UpdateProgressRequest) (*dto.ProgressResponse, error) {
    // Upsert: Create if not exists, update if exists
    progress := &models.UserProgress{
        UserID:         req.UserID,
        MangaID:        req.MangaID,
        CurrentChapter: req.CurrentChapter,
        Status:         req.Status,
        Page:           req.Page,
    }

    return s.repo.Upsert(progress)
}
```

**Response**:
```json
{
    "manga_id": 1,
    "current_chapter": 105,
    "status": "reading",
    "page": 15,
    "updated_at": "2025-12-23T10:30:00Z"
}
```

---

## 5. CLI Client Implementation

### HTTP Client Architecture

The CLI uses a dedicated HTTP client to interact with the API:

```go
type HTTPClient struct {
    baseURL    string
    httpClient *http.Client
    token      string
}

func NewHTTPClient(apiURL string) *HTTPClient {
    return &HTTPClient{
        baseURL: apiURL,
        httpClient: &http.Client{
            Timeout: 10 * time.Second,
        },
    }
}
```

### Example CLI Command: Search Manga

```go
func (c *HTTPClient) SearchManga(query string) ([]MangaResponse, error) {
    // URL-encode query
    params := url.Values{}
    params.Add("q", query)
    fullURL := fmt.Sprintf("%s/api/manga/search?%s", c.baseURL, params.Encode())

    // Prepare request
    req, err := http.NewRequest("GET", fullURL, nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("Authorization", "Bearer "+c.token)

    // Execute request
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    // Parse response
    var result struct {
        Data  []MangaResponse `json:"data"`
        Total int             `json:"total"`
    }
    json.NewDecoder(resp.Body).Decode(&result)

    return result.Data, nil
}
```

---

## 6. Security Best Practices

### 6.1 Input Validation

```go
// DTO with validation tags
type CreateMangaDTO struct {
    Title         string   `json:"title" binding:"required,min=1,max=200"`
    Author        *string  `json:"author,omitempty" binding:"omitempty,max=100"`
    Status        *string  `json:"status,omitempty" binding:"omitempty,oneof=ongoing completed hiatus"`
    TotalChapters *int     `json:"total_chapters,omitempty" binding:"omitempty,min=0"`
    Description   *string  `json:"description,omitempty" binding:"omitempty,max=5000"`
    GenreIDs      []int64  `json:"genre_ids,omitempty"`
}
```

### 6.2 SQL Injection Prevention

GORM automatically uses parameterized queries:

```go
// Safe: GORM prevents SQL injection
db.Where("title ILIKE ?", "%"+userInput+"%").Find(&mangas)

// Unsafe (never do this):
// db.Raw("SELECT * FROM manga WHERE title LIKE '%" + userInput + "%'")
```

### 6.3 Rate Limiting (Future Enhancement)

```go
import "github.com/gin-contrib/ratelimit"

func setupRateLimiter() gin.HandlerFunc {
    return ratelimit.RateLimiter(
        store.New(10*time.Minute, 100), // 100 requests per 10 minutes
        ratelimit.WithKeyFunc(func(c *gin.Context) string {
            return c.ClientIP()
        }),
    )
}
```

---

## 7. Error Handling

### Structured Error Responses

```go
type ErrorResponse struct {
    Error   string                 `json:"error"`
    Code    string                 `json:"code,omitempty"`
    Details map[string]interface{} `json:"details,omitempty"`
}

// Example usage
c.JSON(400, ErrorResponse{
    Error: "Invalid input",
    Code: "VALIDATION_ERROR",
    Details: map[string]interface{}{
        "field": "title",
        "reason": "Title cannot be empty",
    },
})
```

### Error Types

| HTTP Code | Error Type | Description |
|-----------|-----------|-------------|
| 400 | Bad Request | Invalid input/malformed JSON |
| 401 | Unauthorized | Missing/invalid access token |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | Resource doesn't exist |
| 409 | Conflict | Duplicate entry (library, rating) |
| 500 | Internal Server Error | Unexpected server error |

---

## 8. Performance Optimizations

### 8.1 Database Query Optimization

**Before (N+1 Problem)**:
```go
// Bad: Queries genres for each manga individually
mangas := getMangaList()
for _, manga := range mangas {
    manga.Genres = getGenresForManga(manga.ID) // N queries
}
```

**After (Eager Loading)**:
```go
// Good: Single JOIN query
db.Preload("Genres").Find(&mangas)
```

### 8.2 Pagination Best Practices

```go
// Always use LIMIT/OFFSET
db.Limit(pageSize).Offset((page - 1) * pageSize).Find(&results)

// Add total count for UI pagination
var total int64
db.Model(&models.Manga{}).Count(&total)
```

### 8.3 Connection Pooling

```go
// GORM automatically manages connection pool
sqlDB, _ := db.DB()
sqlDB.SetMaxOpenConns(25)
sqlDB.SetMaxIdleConns(5)
sqlDB.SetConnMaxLifetime(5 * time.Minute)
```

---

## 9. Testing Strategy

### 9.1 Unit Tests (Service Layer)

```go
func TestCreateManga(t *testing.T) {
    // Mock repository
    mockRepo := &MockMangaRepo{
        CreateFunc: func(m *models.Manga) error {
            m.ID = 1
            return nil
        },
    }

    svc := NewMangaService(mockRepo)

    manga := &models.Manga{Title: "Test Manga"}
    err := svc.Create(context.Background(), manga)

    assert.NoError(t, err)
    assert.Equal(t, int64(1), manga.ID)
}
```

### 9.2 Integration Tests (Handler Layer)

```go
func TestMangaHandler_Create(t *testing.T) {
    router := setupTestRouter()
    
    reqBody := `{"title":"Test Manga","author":"Test Author"}`
    req, _ := http.NewRequest("POST", "/api/manga", strings.NewReader(reqBody))
    req.Header.Set("Authorization", "Bearer "+testToken)
    req.Header.Set("Content-Type", "application/json")

    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    assert.Equal(t, 201, w.Code)
}
```

---

## 10. Deployment Considerations

### 10.1 Docker Compose Configuration

```yaml
services:
  api-server:
    build: .
    ports:
      - "8084:8084"
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - JWT_SECRET=${JWT_SECRET}
      - TLS_ENABLED=false
    depends_on:
      - postgres
```

### 10.2 Health Checks

```go
// Liveness probe
r.GET("/check-conn", func(c *gin.Context) {
    c.JSON(200, gin.H{"ok": true, "message": "api server running"})
})

// Readiness probe (with DB ping)
r.GET("/db-ping", func(c *gin.Context) {
    sqlDB, _ := gdb.DB()
    if err := sqlDB.Ping(); err != nil {
        c.JSON(500, gin.H{"ok": false, "error": err.Error()})
        return
    }
    c.JSON(200, gin.H{"ok": true})
})
```

---

## 11. Future Enhancements

### Planned Features

1. **GraphQL API**: Alternative to REST for complex queries
2. **Redis Caching**: Cache popular manga/search results
3. **Elasticsearch**: Full-text search improvement
4. **API Versioning**: `/api/v1/`, `/api/v2/` support
5. **Audit Logging**: Track all admin actions
6. **File Upload**: Direct manga cover upload to S3
7. **WebSocket Events**: Real-time notifications for new chapters

---

## 12. Conclusion

The MangaHub HTTP API provides a **robust, scalable, and secure** foundation for manga management. Key strengths include:

✅ **Clean Architecture**: Layered design with clear separation of concerns  
✅ **Security**: JWT authentication with refresh token rotation  
✅ **Performance**: Optimized queries with pagination and eager loading  
✅ **Developer Experience**: CLI client for easy testing and automation  
✅ **Extensibility**: Modular design allows easy feature additions  

The API follows **RESTful principles** and modern Go best practices, making it production-ready while remaining maintainable for future development.