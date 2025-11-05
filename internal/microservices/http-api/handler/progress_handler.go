package handler

import (
	"context"
	"mangahub/internal/microservices/http-api/dto"
	"mangahub/internal/microservices/http-api/models"
	"mangahub/internal/microservices/http-api/service"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ProgressHandler struct {
	progressService service.ProgressService
}

func NewProgressHandler(progressService service.ProgressService) *ProgressHandler {
	return &ProgressHandler{progressService: progressService}
}

// RegisterRoutes registers the progress-related routes
func (h *ProgressHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.GetAllProgress)
	rg.GET("/:manga_id", h.GetProgressByMangaID)
	rg.POST("/:manga_id", h.UpdateProgress)
	rg.DELETE("/:manga_id", h.DeleteProgress)
}

func (h *ProgressHandler) GetAllProgress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	progressList, err := h.progressService.GetAllProgress(ctx, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, progressList)
}
func (h *ProgressHandler) GetProgressByMangaID(c *gin.Context) {
	var req dto.GetProgressByMangaIDRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	progress, err := h.progressService.GetProgressByMangaID(ctx, req.UserID, req.MangaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, progress)
}

func (h *ProgressHandler) UpdateProgress(c *gin.Context) {
	var req dto.UpdateProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	progress := &models.UserProgress{
		UserID:         req.UserID,
		MangaID:        req.MangaID,
		CurrentChapter: req.Chapter,
		Status:         req.Status,
		UpdatedAt:      time.Now(),
	}
	if err := h.progressService.UpdateProgress(ctx, progress); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "progress updated"})
}

func (h *ProgressHandler) DeleteProgress(c *gin.Context) {
	var req dto.DeleteProgressRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := h.progressService.DeleteProgress(ctx, req.UserID, req.MangaID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "progress deleted"})
}
