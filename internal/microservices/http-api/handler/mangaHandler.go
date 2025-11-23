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
	rg.GET("/:id", middleware.RequireScopes("read:manga"), h.Get)

	// Admin-only routes
	rg.POST("/", middleware.RequireScopes("read:manga", "write:manga"), middleware.RequireAdmin(), h.Create)
	rg.PUT("/:id", middleware.RequireScopes("read:manga", "write:manga"), middleware.RequireAdmin(), h.Update)
	rg.DELETE("/:id", middleware.RequireScopes("delete:manga"), middleware.RequireAdmin(), h.Delete)
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

	c.JSON(http.StatusOK, dto.NewPaginatedMangaBasicResponse(resp, page, pageSize, total))
}

func (h *MangaHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
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
			// Log the error but don't fail the request since manga was created
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
		// Manga was created but we couldn't fetch it back
		c.JSON(http.StatusCreated, dto.FromModelToResponse(model))
		return
	}

	c.JSON(http.StatusCreated, dto.FromModelToResponse(*created))
}

func (h *MangaHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
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
	idStr := c.Param("id")
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

	resp := make([]dto.MangaResponse, 0, len(list))
	for _, m := range list {
		resp = append(resp, dto.FromModelToResponse(m))
	}
	c.JSON(http.StatusOK, resp)
}
