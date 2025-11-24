package handler

import (
	"net/http"
	"strconv"

	"mangahub/internal/microservices/http-api/dto"
	"mangahub/internal/microservices/http-api/service"

	"github.com/gin-gonic/gin"
)

type RatingHandler struct {
	ratingService service.RatingService
}

func NewRatingHandler(ratingService service.RatingService) *RatingHandler {
	return &RatingHandler{
		ratingService: ratingService,
	}
}

// RegisterRoutes registers rating-related routes
func (h *RatingHandler) RegisterRoutes(router *gin.RouterGroup) {
	ratings := router.Group("/:manga_id/ratings")
	{
		// Public routes (no additional middleware needed - read access already through parent middleware)
		ratings.GET("", h.List)               // Get all ratings for a manga
		ratings.GET("/average", h.GetAverage) // Get average rating and count

		// Write routes (already authenticated by parent middleware)
		ratings.POST("", h.CreateOrUpdate)  // Create or update user's rating
		ratings.GET("/me", h.GetUserRating) // Get current user's rating
		ratings.DELETE("", h.Delete)        // Delete user's rating
	}
}

// CreateOrUpdate creates or updates a rating for a manga
// POST /api/manga/:manga_id/ratings
func (h *RatingHandler) CreateOrUpdate(c *gin.Context) {
	mangaID, err := strconv.ParseInt(c.Param("manga_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid manga ID"})
		return
	}

	// Get user ID from context (set by AuthMiddleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req dto.CreateRatingDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rating, err := h.ratingService.CreateOrUpdateRating(userID.(string), mangaID, req.Rating)
	if err != nil {
		if err.Error() == "manga not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rating)
}

// GetUserRating retrieves the current user's rating for a manga
// GET /api/manga/:manga_id/ratings/me
func (h *RatingHandler) GetUserRating(c *gin.Context) {
	mangaID, err := strconv.ParseInt(c.Param("manga_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid manga ID"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	rating, err := h.ratingService.GetUserRating(userID.(string), mangaID)
	if err != nil {
		if err.Error() == "rating not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rating)
}

// Delete removes a user's rating for a manga
// DELETE /api/manga/:manga_id/ratings
func (h *RatingHandler) Delete(c *gin.Context) {
	mangaID, err := strconv.ParseInt(c.Param("manga_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid manga ID"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if err := h.ratingService.DeleteRating(userID.(string), mangaID); err != nil {
		if err.Error() == "manga not found" || err.Error() == "rating not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rating deleted successfully"})
}

// List retrieves all ratings for a manga with pagination
// GET /api/manga/:manga_id/ratings?page=1&page_size=20
func (h *RatingHandler) List(c *gin.Context) {
	mangaID, err := strconv.ParseInt(c.Param("manga_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid manga ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	ratings, err := h.ratingService.GetMangaRatings(mangaID, page, pageSize)
	if err != nil {
		if err.Error() == "manga not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ratings)
}

// GetAverage retrieves the average rating and count for a manga
// GET /api/manga/:manga_id/ratings/average
func (h *RatingHandler) GetAverage(c *gin.Context) {
	mangaID, err := strconv.ParseInt(c.Param("manga_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid manga ID"})
		return
	}

	avg, count, err := h.ratingService.GetMangaAverageRating(mangaID)
	if err != nil {
		if err.Error() == "manga not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"average_rating": avg,
		"total_ratings":  count,
	})
}
