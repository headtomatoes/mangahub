package mangadex

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"gorm.io/gorm"
)

// SyncService manages MangaDex data synchronization
type SyncService struct {
	client   *MangaDexClient
	db       *gorm.DB
	notifier *Notifier

	// Configuration
	initialSyncLimit int
	workerCount      int
	rateSemaphore    chan struct{} // Limits concurrent API calls
}

// SyncConfig holds configuration for the sync service
type SyncConfig struct {
	APIKey           string
	DatabaseURL      string
	UDPServerURL     string
	InitialSyncLimit int
	WorkerCount      int
	RateConcurrency  int // Max concurrent API calls (default: 5)
}

// NewSyncService creates a new sync service instance
func NewSyncService(config SyncConfig, db *gorm.DB) *SyncService {
	client := NewClient(config.APIKey)
	notifier := NewNotifier(config.UDPServerURL)

	workerCount := config.WorkerCount
	if workerCount == 0 {
		workerCount = 10 // Default
	}

	rateConcurrency := config.RateConcurrency
	if rateConcurrency == 0 {
		rateConcurrency = 5 // Default (MangaDex limit)
	}

	return &SyncService{
		client:           client,
		db:               db,
		notifier:         notifier,
		initialSyncLimit: config.InitialSyncLimit,
		workerCount:      workerCount,
		rateSemaphore:    make(chan struct{}, rateConcurrency),
	}
}

// ============================================
// SYNC STATE MANAGEMENT
// ============================================

// SyncState represents a sync operation state in the database
type SyncState struct {
	ID            int    `gorm:"primaryKey"`
	SyncType      string `gorm:"unique;not null"`
	LastRunAt     *time.Time
	LastSuccessAt *time.Time
	LastCursor    string
	Status        string
	ErrorMessage  string
	Metadata      string `gorm:"type:jsonb"`
	UpdatedAt     time.Time
}

// TableName specifies the table name for SyncState
func (SyncState) TableName() string {
	return "sync_state"
}

// updateSyncState updates the sync state in database
func (s *SyncService) updateSyncState(syncType, status string, cursor string, err error) error {
	update := map[string]interface{}{
		"last_run_at": time.Now(),
		"status":      status,
		"last_cursor": cursor,
	}

	if status == "completed" {
		now := time.Now()
		update["last_success_at"] = &now
		update["error_message"] = ""
	}

	if err != nil {
		update["error_message"] = err.Error()
	}

	return s.db.Model(&SyncState{}).
		Where("sync_type = ?", syncType).
		Updates(update).Error
}

// getSyncState retrieves sync state from database
func (s *SyncService) getSyncState(syncType string) (*SyncState, error) {
	var state SyncState
	err := s.db.Where("sync_type = ?", syncType).First(&state).Error
	if err != nil {
		return nil, err
	}
	return &state, nil
}

// ============================================
// DATABASE MODELS
// ============================================

// Manga represents a manga entry in database
type Manga struct {
	ID               int64   `gorm:"primaryKey;autoIncrement"`
	MangaDexID       *string `gorm:"column:mangadex_id;type:uuid;unique"`
	Slug             *string `gorm:"unique"`
	Title            string  `gorm:"not null"`
	Author           *string
	Status           *string
	TotalChapters    *int
	Description      *string
	AverageRating    *float64
	CoverURL         *string
	LastSyncedAt     *time.Time
	LastChapterCheck *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// TableName specifies the table name for Manga
func (Manga) TableName() string {
	return "manga"
}

// Chapter represents a chapter entry in database
type Chapter struct {
	ID                int64   `gorm:"primaryKey;autoIncrement"`
	MangaID           int64   `gorm:"not null;index"`
	MangaDexChapterID *string `gorm:"column:mangadex_chapter_id;type:uuid;unique"`
	ChapterNumber     float64 `gorm:"not null"`
	Title             string
	Volume            string
	Pages             int
	PublishedAt       *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// TableName specifies the table name for Chapter
func (Chapter) TableName() string {
	return "chapters"
}

// Genre represents a genre in database
type Genre struct {
	ID   int    `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"unique;not null"`
}

// TableName specifies the table name for Genre
func (Genre) TableName() string {
	return "genres"
}

// MangaGenre represents the many-to-many relationship
type MangaGenre struct {
	ID      int64 `gorm:"primaryKey;autoIncrement"`
	MangaID int64 `gorm:"not null;index"`
	GenreID int   `gorm:"not null;index"`
}

// TableName specifies the table name for MangaGenre
func (MangaGenre) TableName() string {
	return "manga_genres"
}

// ============================================
// HELPER FUNCTIONS
// ============================================

// storeManga stores extracted manga metadata in database
func (s *SyncService) storeManga(ctx context.Context, extracted *ExtractedManga) (int64, error) {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check if manga already exists
	var existingManga Manga
	err := tx.Where("mangadex_id = ?", extracted.MangaDexID).First(&existingManga).Error

	now := time.Now()
	manga := Manga{
		MangaDexID:    &extracted.MangaDexID,
		Slug:          &extracted.Slug,
		Title:         extracted.Title,
		Author:        &extracted.Author,
		Status:        &extracted.Status,
		TotalChapters: &extracted.TotalChapters,
		Description:   &extracted.Description,
		CoverURL:      &extracted.CoverURL,
		LastSyncedAt:  &now,
	}

	if err == gorm.ErrRecordNotFound {
		// Create new manga
		if err := tx.Create(&manga).Error; err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("failed to create manga: %w", err)
		}
	} else if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("database error: %w", err)
	} else {
		// Update existing manga
		manga.ID = existingManga.ID
		if err := tx.Model(&manga).Updates(manga).Error; err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("failed to update manga: %w", err)
		}
	}

	// Store genres
	for _, genreName := range extracted.Genres {
		// Upsert genre
		var genre Genre
		err := tx.Where("name = ?", genreName).FirstOrCreate(&genre, Genre{Name: genreName}).Error
		if err != nil {
			continue
		}

		// Link manga to genre
		var link MangaGenre
		err = tx.Where("manga_id = ? AND genre_id = ?", manga.ID, genre.ID).FirstOrCreate(&link, MangaGenre{
			MangaID: manga.ID,
			GenreID: genre.ID,
		}).Error
		if err != nil {
			log.Printf("[SyncService] Failed to link genre '%s' to manga: %v", genreName, err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return manga.ID, nil
}

// storeChapter stores extracted chapter metadata in database
func (s *SyncService) storeChapter(ctx context.Context, mangaID int64, extracted *ExtractedChapter) error {
	chapter := Chapter{
		MangaID:           mangaID,
		MangaDexChapterID: &extracted.MangaDexChapterID,
		ChapterNumber:     extracted.ChapterNumber,
		Title:             extracted.Title,
		Volume:            extracted.Volume,
		Pages:             extracted.Pages,
		PublishedAt:       &extracted.PublishedAt,
	}

	// Upsert chapter (insert or update on conflict)
	err := s.db.Where("manga_id = ? AND chapter_number = ?", mangaID, extracted.ChapterNumber).
		FirstOrCreate(&chapter).Error

	if err != nil {
		return fmt.Errorf("failed to store chapter: %w", err)
	}

	return nil
}

// getInitialSyncLimit returns the initial sync limit from env or default
func getInitialSyncLimit() int {
	limitStr := os.Getenv("MANGA_SYNC_INITIAL_COUNT")
	if limitStr == "" {
		return 150 // Default
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return 150
	}

	return limit
}

// processManga is a helper to extract and store a single manga
func (s *SyncService) processManga(ctx context.Context, apiManga MangaData) error {
	// Acquire rate semaphore
	s.rateSemaphore <- struct{}{}
	defer func() { <-s.rateSemaphore }()

	// Extract metadata
	extracted, err := ExtractMangaMetadata(apiManga)
	if err != nil {
		return fmt.Errorf("failed to extract metadata: %w", err)
	}

	// Store in database
	mangaID, err := s.storeManga(ctx, extracted)
	if err != nil {
		return fmt.Errorf("failed to store manga: %w", err)
	}

	log.Printf("[SyncService] ✅ Synced: %s (ID: %d, MangaDex: %s)", extracted.Title, mangaID, extracted.MangaDexID)

	return nil
}
