package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"mangahub/internal/microservices/http-api/dto"
	"mangahub/internal/microservices/http-api/middleware"
	"mangahub/internal/microservices/http-api/models"
	"mangahub/internal/microservices/http-api/service"

	"github.com/gin-gonic/gin"
)

type GenreHandler struct {
	svc service.GenreService
}

func NewGenreHandler(svc service.GenreService) *GenreHandler {
	return &GenreHandler{svc: svc}
}

func (h *GenreHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/", middleware.RequireScopes("read:genre"), h.List)
	rg.POST("/", middleware.RequireScopes("write:genre"), h.Create)

	// new route: GET /api/genres/:id/mangas
	rg.GET("/:id/mangas", middleware.RequireScopes("read:manga"), h.GetMangasByGenre)
}

func (h *GenreHandler) List(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	list, err := h.svc.GetAll(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	resp := make([]dto.GenreResponse, 0, len(list))
	for _, g := range list {
		resp = append(resp, dto.GenreFromModel(g))
	}
	c.JSON(http.StatusOK, resp)
}

func (h *GenreHandler) Create(c *gin.Context) {
	var in dto.CreateGenreDTO
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	model := models.Genre{Name: in.Name}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := h.svc.Create(ctx, &model); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dto.GenreFromModel(model))
}

// GetMangasByGenre handles GET /api/genres/:id/mangas
func (h *GenreHandler) GetMangasByGenre(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid genre id"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	list, err := h.svc.GetMangasByGenre(ctx, id)
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
