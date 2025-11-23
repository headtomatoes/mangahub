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

	"mangahub/database"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	// gormdb "mangahub/internal/db" // removed — use database.OpenGorm()
	"mangahub/internal/config"
	h "mangahub/internal/microservices/http-api/handler"

	mid "mangahub/internal/microservices/http-api/middleware"
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

	// Auto-migrate models
	if err := gdb.AutoMigrate(
		&models.Manga{},
		&models.Genre{},
		&models.MangaGenre{},
		&models.User{},
		&models.RefreshToken{},
		&models.UserLibrary{},
		&models.UserProgress{},
		&models.Notification{},
	); err != nil {
		log.Printf("warning: auto-migrate failed (continuing): %v", err)
	}

	// Wire repository, service, handler
	mangaRepo := repo.NewMangaRepo(gdb)
	mangaSvc := svc.NewMangaService(mangaRepo)
	mangaHandler := h.NewMangaHandler(mangaSvc)

	// genres repo/service/handler
	genreRepo := repo.NewGenreRepo(gdb)
	genreSvc := svc.NewGenreService(genreRepo)
	genreHandler := h.NewGenreHandler(genreSvc)

	// auth and user setup
	userRepo := repo.NewUserRepository(gdb)
	refreshToken := repo.NewRefreshTokenRepository(gdb)
	authSvc := svc.NewAuthService(userRepo, refreshToken, cfg)
	authHandler := h.NewAuthHandler(authSvc)

	// library setup
	libraryRepo := repo.NewLibraryRepository(gdb)
	librarySvc := svc.NewLibraryService(libraryRepo, mangaRepo)
	libraryHandler := h.NewLibraryHandler(librarySvc)

	// notification setup
	notificationRepo := repo.NewNotificationRepository(gdb)
	notificationSvc := svc.NewNotificationService(notificationRepo)
	notificationHandler := h.NewNotificationHandler(notificationSvc)

	// ---progress repo/service/handler---
	progressRepo := repo.NewProgressRepository(gdb)
	progressSvc := svc.NewProgressService(progressRepo)
	progressHandler := h.NewProgressHandler(progressSvc)

	// Gin setup
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOrigins,                                     //["http://localhost:3000"] //frontend origin
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, //allowed methods
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"}, //allowed headers
		ExposeHeaders:    []string{"Content-Length"},                          //exposed headers
		AllowCredentials: true,                                                //allow cookies, authorization headers with CORS requests
		MaxAge:           12 * time.Hour,                                      //preflight request cache duration
	}))

	// Public routes
	auth := r.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.POST("/revoke", authHandler.RevokeToken)
	}

	// Protected routes
	api := r.Group("/api")
	api.Use(mid.AuthMiddleware(authSvc))
	{
		mangaHandler.RegisterRoutes(api.Group("/manga")) // newly added the scopes based middleware in handler | other handlers follow same pattern later
		genreHandler.RegisterRoutes(api.Group("/genres"))
		libraryHandler.RegisterRoutes(api.Group("/library"))
		progressHandler.RegisterRoutes(api.Group("/progress"))
		notificationHandler.RegisterRoutes(api.Group("/notifications")) // Add this

	}

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

	addr := fmt.Sprintf("0.0.0.0:%d", cfg.HTTPPort)

	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start HTTPS server in a goroutine
	go func() {
		if cfg.TLSEnabled {
			log.Printf("🚀 server listening on https://%s", addr)
			if err := srv.ListenAndServeTLS(cfg.TLSCertPath, cfg.TLSKeyPath); err != nil && err != http.ErrServerClosed {
				log.Fatalf("server ListenAndServeTLS error: %v", err)
			}
		} else {
			log.Printf("🚀 server listening on http://%s", addr)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("server ListenAndServe error: %v", err)
			}
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
