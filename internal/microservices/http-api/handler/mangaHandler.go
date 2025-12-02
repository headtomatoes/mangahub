package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"mangahub/internal/microservices/http-api/dto"
	"mangahub/internal/microservices/http-api/middleware"
	"mangahub/internal/microservices/http-api/models"
	"mangahub/internal/microservices/http-api/service"

	"github.com/gin-gonic/gin"
)

type MangaHandler struct {
	svc service.MangaService
}

func NewMangaHandler(svc service.MangaService) *MangaHandler {
	return &MangaHandler{svc: svc}
}

func (h *MangaHandler) RegisterRoutes(rg *gin.RouterGroup) {
	// Public routes (any authenticated user)
	rg.GET("/", middleware.RequireScopes("read:manga"), h.List)
	rg.GET("/search", middleware.RequireScopes("read:manga"), h.SearchByTitle)
	rg.GET("/advanced-search", middleware.RequireScopes("read:manga"), h.AdvancedSearch)
	rg.GET("/:manga_id", middleware.RequireScopes("read:manga"), h.Get)

	// Admin-only routes
	rg.POST("/", middleware.RequireScopes("read:manga", "write:manga"), middleware.RequireAdmin(), h.Create)
	rg.PUT("/:manga_id", middleware.RequireScopes("read:manga", "write:manga"), middleware.RequireAdmin(), h.Update)
	rg.DELETE("/:manga_id", middleware.RequireScopes("delete:manga"), middleware.RequireAdmin(), h.Delete)
}

func (h *MangaHandler) List(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Parse pagination parameters
	page := 1
	pageSize := 20

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	list, total, err := h.svc.GetAll(ctx, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Use basic response with only essential fields
	resp := make([]dto.MangaBasicResponse, 0, len(list))
	for _, m := range list {
		resp = append(resp, dto.FromModelToBasicResponse(m))
	}

	c.JSON(http.StatusOK, gin.H{
		"data": resp,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

func (h *MangaHandler) Get(c *gin.Context) {
	idStr := c.Param("manga_id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	m, err := h.svc.GetByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "manga not found"})
		return
	}
	c.JSON(http.StatusOK, dto.FromModelToResponse(*m))
}

func (h *MangaHandler) Create(c *gin.Context) {
	var in dto.CreateMangaDTO
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	model := in.ToModel()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Create manga
	if err := h.svc.Create(ctx, &model); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Assign genres if provided
	if len(in.GenreIDs) > 0 {
		if err := h.svc.ReplaceGenresForManga(ctx, model.ID, in.GenreIDs); err != nil {
			c.JSON(http.StatusCreated, gin.H{
				"manga":   dto.FromModelToResponse(model),
				"warning": "Manga created but failed to assign some genres: " + err.Error(),
			})
			return
		}
	}

	// Fetch the manga with genres to return complete data
	created, err := h.svc.GetByID(ctx, model.ID)
	if err != nil {
		c.JSON(http.StatusCreated, dto.FromModelToResponse(model))
		return
	}

	c.JSON(http.StatusCreated, dto.FromModelToResponse(*created))
}

func (h *MangaHandler) Update(c *gin.Context) {
	idStr := c.Param("manga_id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var in dto.UpdateMangaDTO
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// prepare model with provided fields only
	var m models.Manga
	in.ApplyTo(&m)

	// Update manga basic info
	if err := h.svc.Update(ctx, id, &m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Replace genres if provided
	if in.GenreIDs != nil {
		if err := h.svc.ReplaceGenresForManga(ctx, id, in.GenreIDs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Manga updated but failed to update genres: " + err.Error(),
				"manga": id,
			})
			return
		}
	}

	// Fetch updated manga with genres
	updated, err := h.svc.GetByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.FromModelToResponse(*updated))
}

func (h *MangaHandler) Delete(c *gin.Context) {
	idStr := c.Param("manga_id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := h.svc.Delete(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *MangaHandler) SearchByTitle(c *gin.Context) {
	// accept either ?q=... or ?title=... for compatibility
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		q = strings.TrimSpace(c.Query("title"))
	}
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "q or title query parameter is required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	list, err := h.svc.SearchByTitle(ctx, q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := make([]dto.MangaBasicResponse, 0, len(list))
	for _, m := range list {
		resp = append(resp, dto.FromModelToBasicResponse(m))
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  resp,
		"total": len(resp),
	})
}

// AdvancedSearch handles GET /api/manga/advanced-search with multiple filter parameters
func (h *MangaHandler) AdvancedSearch(c *gin.Context) {
	var filters dto.SearchFilters

	// Manual parsing with sanitization
	filters.Query = strings.TrimSpace(c.Query("q"))
	filters.Status = strings.TrimSpace(c.Query("status"))
	filters.SortBy = strings.TrimSpace(c.Query("sort_by"))

	// Parse genres (comma-separated)
	if genresStr := strings.TrimSpace(c.Query("genres")); genresStr != "" {
		genresList := strings.Split(genresStr, ",")
		filters.Genres = make([]string, 0, len(genresList))
		for _, g := range genresList {
			if trimmed := strings.TrimSpace(g); trimmed != "" {
				filters.Genres = append(filters.Genres, trimmed)
			}
		}
	}

	// Parse min_rating
	if minRatingStr := strings.TrimSpace(c.Query("min_rating")); minRatingStr != "" {
		if minRating, err := strconv.ParseFloat(minRatingStr, 64); err == nil && minRating >= 0 && minRating <= 10 {
			filters.MinRating = &minRating
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid min_rating parameter, must be between 0 and 10"})
			return
		}
	}

	// Parse page
	filters.Page = 1
	if pageStr := strings.TrimSpace(c.Query("page")); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page >= 1 {
			filters.Page = page
		}
	}

	// Parse page_size
	filters.PageSize = 20
	if pageSizeStr := strings.TrimSpace(c.Query("page_size")); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize >= 1 && pageSize <= 100 {
			filters.PageSize = pageSize
		}
	}

	// Validate status
	if filters.Status != "" {
		validStatuses := map[string]bool{"ongoing": true, "completed": true, "hiatus": true}
		if !validStatuses[strings.ToLower(filters.Status)] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status, must be one of: ongoing, completed, hiatus"})
			return
		}
	}

	// Validate sort_by
	if filters.SortBy != "" {
		validSortBy := map[string]bool{"popularity": true, "rating": true, "recent": true, "title": true}
		if !validSortBy[strings.ToLower(filters.SortBy)] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sort_by, must be one of: popularity, rating, recent, title"})
			return
		}
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	list, total, err := h.svc.AdvancedSearch(ctx, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Use MangaBasicResponse for list results
	resp := make([]dto.MangaBasicResponse, 0, len(list))
	for _, m := range list {
		resp = append(resp, dto.FromModelToBasicResponse(m))
	}

	totalPages := int64(0)
	if filters.PageSize > 0 {
		totalPages = (total + int64(filters.PageSize) - 1) / int64(filters.PageSize)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": resp,
		"pagination": gin.H{
			"page":         filters.Page,
			"page_size":    filters.PageSize,
			"total":        total,
			"total_pages":  totalPages,
			"has_next":     filters.Page < int(totalPages),
			"has_previous": filters.Page > 1,
		},
		"filters": gin.H{
			"query":      filters.Query,
			"genres":     filters.Genres,
			"status":     filters.Status,
			"min_rating": filters.MinRating,
			"sort_by":    filters.SortBy,
		},
	})
}
