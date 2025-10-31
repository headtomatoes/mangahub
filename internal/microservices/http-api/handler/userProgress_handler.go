package handler

import (
    "context"
    "net/http"
    "strconv"
    "time"

    "mangahub/internal/microservices/http-api/dto"
    "mangahub/internal/microservices/http-api/service"

    "github.com/gin-gonic/gin"
)

type ProgressHandler struct {
    svc service.ProgressService
}

func NewProgressHandler(svc service.ProgressService) *ProgressHandler {
    return &ProgressHandler{svc: svc}
}

func (h *ProgressHandler) RegisterRoutes(rg *gin.RouterGroup) {
    rg.PUT("/:manga_id", h.Update)
    rg.GET("/:manga_id", h.Get)
    rg.GET("/", h.GetAll)
}

// Update reading progress
func (h *ProgressHandler) Update(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
        return
    }
    
    mangaIDStr := c.Param("manga_id")
    mangaID, err := strconv.ParseInt(mangaIDStr, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manga_id"})
        return
    }
    
    var req dto.UpdateProgressRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
    defer cancel()
    
    if err := h.svc.Update(ctx, userID.(string), mangaID, req.CurrentChapter, req.Page, req.Status); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"message": "progress updated"})
}

// Get progress for a specific manga
func (h *ProgressHandler) Get(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
        return
    }
    
    mangaIDStr := c.Param("manga_id")
    mangaID, err := strconv.ParseInt(mangaIDStr, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manga_id"})
        return
    }
    
    ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
    defer cancel()
    
    progress, err := h.svc.Get(ctx, userID.(string), mangaID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    if progress == nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "no progress found"})
        return
    }
    
    c.JSON(http.StatusOK, dto.ProgressResponse{
        MangaID:        progress.MangaID,
        CurrentChapter: progress.CurrentChapter,
        Page:           progress.Page,
        Status:         progress.Status,
        UpdatedAt:      progress.UpdatedAt,
    })
}

// GetAll gets all progress for user
func (h *ProgressHandler) GetAll(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
        return
    }
    
    ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
    defer cancel()
    
    progress, err := h.svc.GetByUser(ctx, userID.(string))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    resp := make([]dto.ProgressResponse, 0, len(progress))
    for _, p := range progress {
        resp = append(resp, dto.ProgressResponse{
            MangaID:        p.MangaID,
            CurrentChapter: p.CurrentChapter,
            Page:           p.Page,
            Status:         p.Status,
            UpdatedAt:      p.UpdatedAt,
        })
    }
    
    c.JSON(http.StatusOK, resp)
}