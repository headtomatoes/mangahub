package handler

import (
	"context"
	"mangahub/internal/microservices/http-api/dto"
	"mangahub/internal/microservices/http-api/middleware"
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
	rg.GET("", middleware.RequireScopes("read:progress"), h.GetAllProgress)
	rg.GET("/:manga_id", middleware.RequireScopes("read:progress"), h.GetProgressByMangaID)
	rg.POST("/:manga_id", middleware.RequireScopes("write:progress"), h.UpdateProgress)
	rg.DELETE("/:manga_id", middleware.RequireScopes("write:progress"), h.DeleteProgress)
}

func (h *ProgressHandler) GetAllProgress(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	userHistory, err := h.progressService.GetAllProgress(ctx, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// map to DTO

	var progressList []dto.ProgressResponse
	for _, progress := range *userHistory {
		progressList = append(progressList, dto.ProgressResponse{
			UserID:  progress.UserID,
			MangaID: progress.MangaID,
			// MangaTitle: progress.MangaTitle, // Temporarily disabled
			Chapter:   progress.CurrentChapter,
			Status:    progress.Status,
			UpdatedAt: progress.UpdatedAt.Format(time.RFC3339),
		})
	}

	progressHistory := dto.ProgressHistoryResponse{
		History: progressList,
		Total:   len(progressList),
	}
	c.JSON(http.StatusOK, progressHistory)
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
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	var req dto.UpdateProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	progress := &models.UserProgress{
		UserID:  userID.(string),
		MangaID: req.MangaID,
		// MangaTitle:     req.MangaTitle, // Temporarily disabled until migration runs
		CurrentChapter: req.Chapter,
		Status:         req.Status,
		UpdatedAt:      time.Now(),
	}
	if err := h.progressService.UpdateProgress(ctx, progress); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	res := dto.ProgressResponse{
		UserID:  progress.UserID,
		MangaID: progress.MangaID,
		// MangaTitle: progress.MangaTitle, // Temporarily disabled
		Chapter:   progress.CurrentChapter,
		Status:    progress.Status,
		UpdatedAt: progress.UpdatedAt.Format(time.RFC3339),
	}
	c.JSON(http.StatusOK, res)
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
