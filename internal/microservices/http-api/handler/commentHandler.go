package handler

import (
	"net/http"
	"strconv"

	"mangahub/internal/microservices/http-api/dto"
	"mangahub/internal/microservices/http-api/service"

	"github.com/gin-gonic/gin"
)

type CommentHandler struct {
	commentService service.CommentService
}

func NewCommentHandler(commentService service.CommentService) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
	}
}

// RegisterRoutes registers comment-related routes
func (h *CommentHandler) RegisterRoutes(router *gin.RouterGroup) {
	// Manga comments
	mangaComments := router.Group("/:manga_id/comments")
	{
		// Public routes
		mangaComments.GET("", h.ListByManga) // Get all comments for a manga

		// Write routes (already authenticated by parent middleware)
		mangaComments.POST("", h.Create) // Create a comment
	}

	// Comment operations (already authenticated by parent middleware)
	comments := router.Group("/comments")
	{
		comments.GET("/:id", h.GetByID)          // Get a specific comment
		comments.PUT("/:id", h.Update)           // Update a comment (user's own)
		comments.DELETE("/:id", h.Delete)        // Delete a comment (user's own)
		comments.GET("/me", h.ListByCurrentUser) // Get current user's comments
	}
}

// Create creates a new comment for a manga
// POST /api/manga/:manga_id/comments
func (h *CommentHandler) Create(c *gin.Context) {
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

	var req dto.CreateCommentDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment, err := h.commentService.CreateComment(userID.(string), mangaID, req.Content)
	if err != nil {
		if err.Error() == "manga not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, comment)
}

// Update updates an existing comment
// PUT /api/comments/:id
func (h *CommentHandler) Update(c *gin.Context) {
	commentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req dto.UpdateCommentDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment, err := h.commentService.UpdateComment(commentID, userID.(string), req.Content)
	if err != nil {
		if err.Error() == "comment not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "you don't have permission to update this comment" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, comment)
}

// Delete deletes a comment
// DELETE /api/comments/:id
func (h *CommentHandler) Delete(c *gin.Context) {
	commentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if err := h.commentService.DeleteComment(commentID, userID.(string)); err != nil {
		if err.Error() == "comment not found or you don't have permission to delete it" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment deleted successfully"})
}

// GetByID retrieves a comment by ID
// GET /api/comments/:id
func (h *CommentHandler) GetByID(c *gin.Context) {
	commentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}

	comment, err := h.commentService.GetCommentByID(commentID)
	if err != nil {
		if err.Error() == "comment not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, comment)
}

// ListByManga retrieves all comments for a manga with pagination
// GET /api/manga/:manga_id/comments?page=1&page_size=20
func (h *CommentHandler) ListByManga(c *gin.Context) {
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

	comments, err := h.commentService.GetMangaComments(mangaID, page, pageSize)
	if err != nil {
		if err.Error() == "manga not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, comments)
}

// ListByCurrentUser retrieves all comments by the current user with pagination
// GET /api/comments/me?page=1&page_size=20
func (h *CommentHandler) ListByCurrentUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
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

	comments, err := h.commentService.GetUserComments(userID.(string), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, comments)
}
