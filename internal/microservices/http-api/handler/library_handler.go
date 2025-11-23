package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"mangahub/internal/microservices/http-api/dto"
	"mangahub/internal/microservices/http-api/middleware"
	"mangahub/internal/microservices/http-api/service"

	"github.com/gin-gonic/gin"
)

type LibraryHandler struct {
	svc service.LibraryService
}

func NewLibraryHandler(svc service.LibraryService) *LibraryHandler {
	return &LibraryHandler{svc: svc}
}

func (h *LibraryHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/", middleware.RequireScopes("write:library"), h.Add)
	rg.GET("/", middleware.RequireScopes("read:library"), h.List)
	rg.DELETE("/:manga_id", middleware.RequireScopes("write:library"), h.Remove)
}

// Add manga to user's library
func (h *LibraryHandler) Add(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		fmt.Println("userID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var req dto.AddToLibraryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := h.svc.Add(ctx, userID.(string), req.MangaID); err != nil {
		if err == service.ErrAlreadyInLibrary {
			c.JSON(http.StatusConflict, gin.H{"error": "manga already in library"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "manga added to library"})
}

// List user's library
func (h *LibraryHandler) List(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		fmt.Println("user_id not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	library, err := h.svc.List(ctx, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to response DTOs
	items := make([]dto.LibraryResponse, 0, len(library))
	for _, item := range library {
		resp := dto.LibraryResponse{
			ID:      item.ID,
			MangaID: item.MangaID,
			AddedAt: item.AddedAt,
		}
		if item.Manga != nil {
			resp.Manga = dto.FromModelToResponse(*item.Manga)
		}
		items = append(items, resp)
	}

	c.JSON(http.StatusOK, dto.LibraryListResponse{
		Items: items,
		Total: len(items),
	})
}

// Remove manga from library
func (h *LibraryHandler) Remove(c *gin.Context) {
	userID, exists := c.Get("userID")
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

	if err := h.svc.Remove(ctx, userID.(string), mangaID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
