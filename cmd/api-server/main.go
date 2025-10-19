package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"mangahub/database"
	// gormdb "mangahub/internal/db" // removed — use database.OpenGorm()
	"mangahub/internal/config"
	h "mangahub/internal/microservices/http-api/handler"
	"mangahub/internal/microservices/http-api/models"
	repo "mangahub/internal/microservices/http-api/repository"
	svc "mangahub/internal/microservices/http-api/service"
)

func main() {
	// Load config (fallback to env/default)
	cfg, err := config.LoadConfig()
	if err != nil {
		// fallback to env or default port
		p := 8084
		if v := os.Getenv("HTTP_PORT"); v != "" {
			if pi, err := strconv.Atoi(v); err == nil {
				p = pi
			}
		}
		cfg = &config.Config{HTTPPort: p}
	}

	// Try to initialize optional pgx pool (used by some packages). Non-fatal.
	if err := database.Connect(); err != nil {
		log.Printf("warning: pgx connect failed (continuing): %v", err)
	} else {
		log.Println("pgx pool connected")
		defer database.Close()
	}

	// Open GORM DB (used by repository)
	gdb, err := database.OpenGorm()
	if err != nil {
		log.Fatalf("failed to open gorm DB: %v", err)
	}

	// Auto-migrate manga model (safe for dev; careful in prod)
	if err := gdb.AutoMigrate(&models.Manga{}); err != nil {
		log.Fatalf("auto-migrate failed: %v", err)
	}

	// Wire repository, service, handler
	mangaRepo := repo.NewMangaRepo(gdb)
	mangaSvc := svc.NewMangaService(mangaRepo)
	mangaHandler := h.NewMangaHandler(mangaSvc)

	// Gin setup
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// API routes
	api := r.Group("/api")
	mangaHandler.RegisterRoutes(api.Group("/manga"))

	// Health/readiness
	r.GET("/check-conn", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true, "message": "api server running"})
	})

	// Active DB ping using GORM SQL DB
	r.GET("/db-ping", func(c *gin.Context) {
		sqlDB, err := gdb.DB()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": "db handle unavailable"})
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := sqlDB.PingContext(ctx); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// HTTP server with graceful shutdown
	addr := fmt.Sprintf("0.0.0.0:%d", cfg.HTTPPort)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		log.Printf("🚀 server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server ListenAndServe error: %v", err)
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
	log.Println("server stopped")
}
