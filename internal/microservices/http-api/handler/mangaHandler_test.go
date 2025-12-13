package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"

	//"strconv"
	"testing"
	"time"

	"mangahub/internal/microservices/http-api/dto"
	"mangahub/internal/microservices/http-api/handler"
	"mangahub/internal/microservices/http-api/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- HELPER FUNCTIONS FOR POINTERS ---
func stringPtr(s string) *string { return &s }

// func intPtr(i int) *int          { return &i }
func floatPtr(f float64) *float64 { return &f }

// --- MOCK SERVICE ---

type MockMangaService struct {
	mock.Mock
}

func (m *MockMangaService) GetAll(ctx context.Context, page, pageSize int) ([]models.Manga, int64, error) {
	args := m.Called(ctx, page, pageSize)
	return args.Get(0).([]models.Manga), args.Get(1).(int64), args.Error(2)
}

func (m *MockMangaService) GetByID(ctx context.Context, id int64) (*models.Manga, error) {
	args := m.Called(ctx, id)
	// Handle nil return safely
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Manga), args.Error(1)
}

func (m *MockMangaService) Create(ctx context.Context, manga *models.Manga) error {
	args := m.Called(ctx, manga)
	return args.Error(0)
}

func (m *MockMangaService) Update(ctx context.Context, id int64, manga *models.Manga) error {
	args := m.Called(ctx, id, manga)
	return args.Error(0)
}

func (m *MockMangaService) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMangaService) SearchByTitle(ctx context.Context, title string) ([]models.Manga, error) {
	args := m.Called(ctx, title)
	return args.Get(0).([]models.Manga), args.Error(1)
}

func (m *MockMangaService) AdvancedSearch(ctx context.Context, filters dto.SearchFilters) ([]models.Manga, int64, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]models.Manga), args.Get(1).(int64), args.Error(2)
}

func (m *MockMangaService) ReplaceGenresForManga(ctx context.Context, mangaID int64, genreIDs []int64) error {
	args := m.Called(ctx, mangaID, genreIDs)
	return args.Error(0)
}

// --- SETUP ---

func setupRouter(mockService *MockMangaService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	h := handler.NewMangaHandler(mockService)

	rg := r.Group("/api/manga")
	{
		rg.GET("", h.List) // Changed from "/" to ""
		rg.GET("/:manga_id", h.Get)
		rg.GET("/search", h.SearchByTitle)
		rg.GET("/advanced-search", h.AdvancedSearch)
		rg.POST("", h.Create) // Changed from "/" to ""
		rg.PUT("/:manga_id", h.Update)
		rg.DELETE("/:manga_id", h.Delete)
	}
	return r
}

// Add this helper function
func mockAuthMiddleware(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", "test-user-id")
		c.Set("username", "testuser")
		c.Set("role", role)
		c.Next()
	}
}

// Update setupRouter to accept role parameter
func setupRouterWithAuth(mockService *MockMangaService, role string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	h := handler.NewMangaHandler(mockService)

	rg := r.Group("/api/manga")

	// Apply auth middleware if role is provided
	if role != "" {
		rg.Use(mockAuthMiddleware(role))
	}
	{
		rg.GET("", h.List)
		rg.GET("/:manga_id", h.Get)
		rg.GET("/search", h.SearchByTitle)
		rg.GET("/advanced-search", h.AdvancedSearch)
		rg.POST("", h.Create)
		rg.PUT("/:manga_id", h.Update)
		rg.DELETE("/:manga_id", h.Delete)
	}
	return r
}

// --- TESTS ---

func TestMangaHandler_List(t *testing.T) {
	mockService := new(MockMangaService)
	r := setupRouter(mockService)

	// Prepare data suitable for MangaBasicResponse
	expectedManga := []models.Manga{
		{ID: 1, Title: "Manga 1", Author: stringPtr("Author A"), AverageRating: floatPtr(9.5)},
		{ID: 2, Title: "Manga 2", Status: stringPtr("completed")},
	}
	expectedTotal := int64(50)

	t.Run("Success", func(t *testing.T) {
		mockService.On("GetAll", mock.Anything, 1, 20).Return(expectedManga, expectedTotal, nil).Once()

		req, _ := http.NewRequest(http.MethodGet, "/api/manga", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		data := response["data"].([]interface{})
		assert.Len(t, data, 2)

		// Verify fields mapped by FromModelToBasicResponse
		item1 := data[0].(map[string]interface{})
		assert.Equal(t, "Manga 1", item1["title"])
		assert.Equal(t, "Author A", item1["author"])
		assert.Equal(t, 9.5, item1["average_rating"])
	})
}

func TestMangaHandler_Get(t *testing.T) {
	mockService := new(MockMangaService)
	r := setupRouter(mockService)

	// Prepare detailed data for MangaResponse
	now := time.Now()
	expectedManga := &models.Manga{
		ID:            101,
		Title:         "Test Manga",
		Genres:        []models.Genre{{Name: "Action"}, {Name: "Adventure"}}, // Assuming models.Genre exists
		Slug:          stringPtr("test-manga"),
		CreatedAt:     &now,
		AverageRating: floatPtr(8.8),
	}

	t.Run("Success", func(t *testing.T) {
		mockService.On("GetByID", mock.Anything, int64(101)).Return(expectedManga, nil).Once()

		req, _ := http.NewRequest(http.MethodGet, "/api/manga/101", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.MangaResponse
		json.Unmarshal(w.Body.Bytes(), &response)

		// Assert fields mapped by FromModelToResponse
		assert.Equal(t, int64(101), response.ID)
		assert.Equal(t, "Test Manga", response.Title)
		assert.Equal(t, "test-manga", *response.Slug)
		assert.Equal(t, 8.8, *response.AverageRating)

		// Verify Genre conversion (slice of structs to slice of strings)
		assert.Len(t, response.Genres, 2)
		assert.Contains(t, response.Genres, "Action")
		assert.Contains(t, response.Genres, "Adventure")
	})

	t.Run("NotFound", func(t *testing.T) {
		mockService.On("GetByID", mock.Anything, int64(999)).Return(nil, errors.New("not found")).Once()
		req, _ := http.NewRequest(http.MethodGet, "/api/manga/999", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestMangaHandler_Create(t *testing.T) {
	mockService := new(MockMangaService)
	r := setupRouterWithAuth(mockService, "admin") // Pass "admin" role

	createDTO := dto.CreateMangaDTO{
		Title:       "New Manga",
		Author:      stringPtr("New Author"),
		Description: stringPtr("Desc"),
		GenreIDs:    []int64{1, 2},
	}

	t.Run("Success", func(t *testing.T) {
		// Mock logic: The handler converts DTO to Model then calls Service.Create
		// We use mock.MatchedBy to verify the model passed to the service has correct data
		mockService.On("Create", mock.Anything, mock.MatchedBy(func(m *models.Manga) bool {
			return m.Title == "New Manga" && *m.Author == "New Author"
		})).Return(nil).Once()

		// Mock the ReplaceGenresForManga call
		mockService.On("ReplaceGenresForManga", mock.Anything, mock.Anything, []int64{1, 2}).Return(nil).Once()

		// Mock GetByID call (handler fetches the created manga to return it)
		createdManga := &models.Manga{
			ID:     1,
			Title:  "New Manga",
			Author: stringPtr("New Author"),
		}
		mockService.On("GetByID", mock.Anything, mock.Anything).Return(createdManga, nil).Once()

		body, _ := json.Marshal(createDTO)
		req, _ := http.NewRequest(http.MethodPost, "/api/manga", bytes.NewBuffer(body)) // Removed trailing slash
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("ValidationError", func(t *testing.T) {
		// Title is required in CreateMangaDTO
		invalidDTO := dto.CreateMangaDTO{
			Author: stringPtr("Author Only"),
		}
		body, _ := json.Marshal(invalidDTO)
		req, _ := http.NewRequest(http.MethodPost, "/api/manga", bytes.NewBuffer(body)) // Removed trailing slash
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestMangaHandler_Update(t *testing.T) {
	mockService := new(MockMangaService)
	r := setupRouterWithAuth(mockService, "admin") // Use admin auth

	updateDTO := dto.UpdateMangaDTO{
		Title:  stringPtr("Updated Title"),
		Status: stringPtr("completed"),
	}

	t.Run("Success", func(t *testing.T) {
		mangaID := int64(10)

		// Mock GetByID to return existing manga (handler may fetch before updating)
		existingManga := &models.Manga{
			ID:     mangaID,
			Title:  "Old Title",
			Status: stringPtr("ongoing"),
		}
		mockService.On("GetByID", mock.Anything, mangaID).Return(existingManga, nil).Once()

		// The Handler usually:
		// 1. Fetches existing (optional, depends on implementation)
		// 2. Applies changes
		// 3. Calls Update

		// Assuming Handler calls Update with the modified model
		mockService.On("Update", mock.Anything, mangaID, mock.MatchedBy(func(m *models.Manga) bool {
			return m.Title == "Updated Title" && *m.Status == "completed"
		})).Return(nil).Once()

		body, _ := json.Marshal(updateDTO)
		req, _ := http.NewRequest(http.MethodPut, "/api/manga/10", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestMangaHandler_Delete(t *testing.T) {
	mockService := new(MockMangaService)
	r := setupRouterWithAuth(mockService, "admin") // Use admin auth

	t.Run("Success", func(t *testing.T) {
		mockService.On("Delete", mock.Anything, int64(55)).Return(nil).Once()

		req, _ := http.NewRequest(http.MethodDelete, "/api/manga/55", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestMangaHandler_SearchByTitle(t *testing.T) {
	mockService := new(MockMangaService)
	r := setupRouter(mockService)

	t.Run("Success", func(t *testing.T) {
		mockService.On("SearchByTitle", mock.Anything, "naruto").Return([]models.Manga{}, nil).Once()

		req, _ := http.NewRequest(http.MethodGet, "/api/manga/search?q=naruto", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestMangaHandler_AdvancedSearch(t *testing.T) {
	mockService := new(MockMangaService)
	r := setupRouter(mockService)

	expectedManga := []models.Manga{{ID: 1, Title: "A"}}

	t.Run("Success_ComplexFilters", func(t *testing.T) {
		// Define expected filter structure based on dto.SearchFilters
		// Note: Pointers must value-match
		expectedFilters := dto.SearchFilters{
			Query:    "adventure",
			Status:   "completed",
			SortBy:   "popularity",
			Genres:   []string{"isekai", "fantasy"},
			Page:     1,
			PageSize: 10,
			// MinRating is handled via matcher because comparing pointers directly is tricky if instances differ
		}

		mockService.On("AdvancedSearch", mock.Anything, mock.MatchedBy(func(f dto.SearchFilters) bool {
			if f.Query != expectedFilters.Query {
				return false
			}
			if f.Status != expectedFilters.Status {
				return false
			}
			if len(f.Genres) != 2 {
				return false
			}
			if *f.MinRating != 7.5 {
				return false
			} // Check pointer value
			return true
		})).Return(expectedManga, int64(1), nil).Once()

		// URL construction must match form tags in DTO (`form:"..."`)
		url := "/api/manga/advanced-search?q=adventure&status=completed&sort_by=popularity&genres=isekai,fantasy&min_rating=7.5&page=1&page_size=10"
		req, _ := http.NewRequest(http.MethodGet, url, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify Response Structure
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		// Check that filters are echoed back correctly
		filtersResp := response["filters"].(map[string]interface{})
		assert.Equal(t, "adventure", filtersResp["query"]) // Handler returns "query" not "q"
		assert.Equal(t, 7.5, filtersResp["min_rating"])
	})

	t.Run("Invalid_Enum_Status", func(t *testing.T) {
		// DTO tag: binding:"omitempty,oneof=ongoing completed hiatus"
		req, _ := http.NewRequest(http.MethodGet, "/api/manga/advanced-search?status=dropped", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
