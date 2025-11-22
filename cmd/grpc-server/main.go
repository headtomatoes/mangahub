package main

import (
	"log"
	"mangahub/database"
	"mangahub/internal/config"
	"mangahub/internal/microservices/grpc"
	"mangahub/internal/microservices/http-api/models"
	rb "mangahub/internal/microservices/http-api/repository"
	"strconv"
)

func main() {
	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
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
	port := cfg.GRPCPort
	log.Printf("gRPC server starting on port %d", port)
	portStr := ":" + strconv.Itoa(port)
	// Create repositories
	mangaRepo := rb.NewMangaRepo(gdb)
	progressRepo := rb.NewProgressRepository(gdb)

	// Start gRPC server
	if err := grpc.StartGRPCServer(portStr, mangaRepo, progressRepo); err != nil {
		log.Fatalf("gRPC server failed: %v", err)
	}
}
