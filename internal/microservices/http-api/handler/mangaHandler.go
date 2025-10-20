package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"mangahub/internal/microservices/http-api/dto"
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
	rg.GET("/", h.List)
	rg.GET("/search", h.SearchByTitle) // new route (supports ?q= or ?title=)
	rg.GET("/:id", h.Get)
	rg.POST("/", h.Create)
	rg.PUT("/:id", h.Update)
	rg.DELETE("/:id", h.Delete)
}

func (h *MangaHandler) List(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	list, err := h.svc.GetAll(ctx)
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

	if err := h.svc.Create(ctx, &model); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dto.FromModelToResponse(model))
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

	if err := h.svc.Update(ctx, id, &m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

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
