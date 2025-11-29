package mangadex

import (
	"context"
	"fmt"
	"log"
	"time"
)

// formatMangaDexDate formats time to MangaDex API format (YYYY-MM-DDTHH:MM:SS)
func formatMangaDexDate(t time.Time) string {
	// MangaDex expects format without timezone suffix
	return t.Format("2006-01-02T15:04:05")
}

// parseMangaDexDate parses MangaDex date format back to time.Time
func parseMangaDexDate(dateStr string) (time.Time, error) {
	// Try with timezone first
	t, err := time.Parse(time.RFC3339, dateStr)
	if err == nil {
		return t, nil
	}

	// Try without timezone
	return time.Parse("2006-01-02T15:04:05", dateStr)
}

// RunInitialSync performs one-time bulk import of manga
// Fetches configured number of manga (default: 150) with complete metadata
// Does NOT fetch historical chapters - only stores baseline total_chapters
func (s *SyncService) RunInitialSync(ctx context.Context) error {
	log.Println("[InitialSync] Starting initial sync...")

	// Check if already completed
	state, err := s.getSyncState("initial_sync")
	if err == nil && state.Status == "completed" {
		log.Println("[InitialSync] Already completed, skipping")
		return nil
	}

	// Update status to running
	if err := s.updateSyncState("initial_sync", "running", "", nil); err != nil {
		return fmt.Errorf("failed to update sync state: %w", err)
	}

	limit := s.initialSyncLimit
	if limit == 0 {
		limit = getInitialSyncLimit()
	}

	log.Printf("[InitialSync] Fetching %d manga with concurrent processing (%d workers)...", limit, s.workerCount)

	// Create worker pool
	pool := NewWorkerPool(s.workerCount)
	pool.Start()
	defer pool.Wait()

	offset := 0
	batchSize := 100
	totalSynced := 0

	for offset < limit {
		// Fetch batch
		params := BuildMangaQueryParams(batchSize, offset, "")
		resp, err := s.client.GetManga(ctx, params)
		if err != nil {
			s.updateSyncState("initial_sync", "error", "", err)
			return fmt.Errorf("failed to fetch manga batch: %w", err)
		}

		log.Printf("[InitialSync] Fetched batch: %d manga (offset %d)", len(resp.Data), offset)

		// Process each manga concurrently
		for _, apiManga := range resp.Data {
			apiManga := apiManga // Capture loop variable

			pool.Submit(func(ctx context.Context) error {
				if err := s.processManga(ctx, apiManga); err != nil {
					log.Printf("[InitialSync] Failed to process manga: %v", err)
					return err
				}
				totalSynced++
				return nil
			})
		}

		offset += batchSize

		// Break if we've fetched all available
		if len(resp.Data) < batchSize {
			break
		}
	}

	// Wait for all workers to complete
	pool.Wait()

	// Update sync state
	now := time.Now()
	if err := s.updateSyncState("initial_sync", "completed", formatMangaDexDate(now), nil); err != nil {
		return fmt.Errorf("failed to update sync state: %w", err)
	}

	log.Printf("[InitialSync] ✅ Completed! Synced %d manga", totalSynced)
	return nil
}

// PollNewManga checks for newly published manga on MangaDex
// Runs every 24 hours, detects manga created since last poll
func (s *SyncService) PollNewManga(ctx context.Context) error {
	log.Println("[NewMangaPoll] Starting new manga detection...")

	// Update status to running
	if err := s.updateSyncState("new_manga_poll", "running", "", nil); err != nil {
		return fmt.Errorf("failed to update sync state: %w", err)
	}

	// Get last cursor (timestamp)
	state, err := s.getSyncState("new_manga_poll")
	cursor := ""
	if err == nil && state.LastCursor != "" {
		cursor = state.LastCursor
	} else {
		// First run: use 24 hours ago
		cursor = formatMangaDexDate(time.Now().Add(-24 * time.Hour))
	}

	log.Printf("[NewMangaPoll] Checking for manga created since: %s", cursor)

	// Fetch new manga
	params := BuildMangaQueryParams(100, 0, cursor)
	params.Set("order[createdAt]", "asc") // Chronological order

	resp, err := s.client.GetManga(ctx, params)
	if err != nil {
		s.updateSyncState("new_manga_poll", "error", cursor, err)
		return fmt.Errorf("failed to fetch new manga: %w", err)
	}

	if len(resp.Data) == 0 {
		log.Println("[NewMangaPoll] No new manga found")
		s.updateSyncState("new_manga_poll", "completed", cursor, nil)
		return nil
	}

	log.Printf("[NewMangaPoll] Found %d new manga", len(resp.Data))

	// Create worker pool
	pool := NewWorkerPool(s.workerCount)
	pool.Start()
	defer pool.Wait()

	newCount := 0
	lastCreatedAt := cursor

	for _, apiManga := range resp.Data {
		apiManga := apiManga // Capture loop variable

		pool.Submit(func(ctx context.Context) error {
			// Check if already exists
			var existing Manga
			err := s.db.Where("mangadex_id = ?", apiManga.ID).First(&existing).Error

			if err == nil {
				// Already exists, skip
				return nil
			}

			// Process new manga
			if err := s.processManga(ctx, apiManga); err != nil {
				return err
			}

			// Extract metadata for notification
			extracted, err := ExtractMangaMetadata(apiManga)
			if err != nil {
				return err
			}

			// Get manga ID from database
			var manga Manga
			if err := s.db.Where("mangadex_id = ?", apiManga.ID).First(&manga).Error; err != nil {
				return err
			}

			// Send notification (async)
			s.notifier.NotifyNewManga(manga.ID, extracted.Title)

			newCount++
			return nil
		})

		// Update cursor to last manga's createdAt
		if apiManga.Attributes.CreatedAt != "" {
			lastCreatedAt = apiManga.Attributes.CreatedAt
		}
	}

	pool.Wait()

	// Update sync state
	if err := s.updateSyncState("new_manga_poll", "completed", lastCreatedAt, nil); err != nil {
		return fmt.Errorf("failed to update sync state: %w", err)
	}

	log.Printf("[NewMangaPoll] ✅ Completed! Found %d new manga", newCount)
	return nil
}

// CheckChapterUpdates checks for new chapters for tracked manga
// Runs every 48 hours (2 days), only stores chapters > baseline
func (s *SyncService) CheckChapterUpdates(ctx context.Context) error {
	log.Println("[ChapterCheck] Starting chapter update detection...")

	// Update status to running
	if err := s.updateSyncState("chapter_check", "running", "", nil); err != nil {
		return fmt.Errorf("failed to update sync state: %w", err)
	}

	// Get manga that haven't been checked in 48 hours
	var mangaList []Manga
	err := s.db.Where("mangadex_id IS NOT NULL").
		Where("last_chapter_check IS NULL OR last_chapter_check < ?", time.Now().Add(-48*time.Hour)).
		Order("last_chapter_check ASC NULLS FIRST").
		Limit(50). // Check 50 manga per run
		Find(&mangaList).Error

	if err != nil {
		s.updateSyncState("chapter_check", "error", "", err)
		return fmt.Errorf("failed to fetch manga list: %w", err)
	}

	if len(mangaList) == 0 {
		log.Println("[ChapterCheck] No manga to check")
		s.updateSyncState("chapter_check", "completed", "", nil)
		return nil
	}

	log.Printf("[ChapterCheck] Checking %d manga for chapter updates...", len(mangaList))

	// Create worker pool
	pool := NewWorkerPool(s.workerCount)
	pool.Start()
	defer pool.Wait()

	updateCount := 0

	for _, manga := range mangaList {
		manga := manga // Capture loop variable

		pool.Submit(func(ctx context.Context) error {
			return s.checkMangaChapters(ctx, &manga, &updateCount)
		})
	}

	pool.Wait()

	// Update sync state
	if err := s.updateSyncState("chapter_check", "completed", "", nil); err != nil {
		return fmt.Errorf("failed to update sync state: %w", err)
	}

	log.Printf("[ChapterCheck] ✅ Completed! Found updates for %d manga", updateCount)
	return nil
}

// checkMangaChapters checks a single manga for new chapters
func (s *SyncService) checkMangaChapters(ctx context.Context, manga *Manga, updateCount *int) error {
	// Acquire rate semaphore
	s.rateSemaphore <- struct{}{}
	defer func() { <-s.rateSemaphore }()

	// Fetch latest chapters
	params := BuildChapterQueryParams(100, "desc")
	resp, err := s.client.GetMangaFeed(ctx, *manga.MangaDexID, params)
	if err != nil {
		log.Printf("[ChapterCheck] Failed to fetch chapters for %s: %v", manga.Title, err)
		return err
	}

	baseline := 0
	if manga.TotalChapters != nil {
		baseline = *manga.TotalChapters
	}

	// Filter for new chapters (chapter_number > baseline)
	var newChapters []ChapterData
	highestChapter := baseline

	for _, apiChapter := range resp.Data {
		extracted, err := ExtractChapterMetadata(apiChapter)
		if err != nil {
			continue
		}

		chapterNum := int(extracted.ChapterNumber)

		// Only process chapters after baseline
		if chapterNum > baseline {
			newChapters = append(newChapters, apiChapter)

			if chapterNum > highestChapter {
				highestChapter = chapterNum
			}
		}
	}

	// Update last_chapter_check timestamp
	now := time.Now()
	s.db.Model(&manga).Update("last_chapter_check", &now)

	if len(newChapters) == 0 {
		return nil
	}

	log.Printf("[ChapterCheck] Found %d new chapters for %s (baseline: %d)", len(newChapters), manga.Title, baseline)

	// Store new chapters and send notifications
	for _, apiChapter := range newChapters {
		extracted, err := ExtractChapterMetadata(apiChapter)
		if err != nil {
			continue
		}

		// Store chapter
		if err := s.storeChapter(ctx, manga.ID, extracted); err != nil {
			log.Printf("[ChapterCheck] Failed to store chapter: %v", err)
			continue
		}

		// Send notification (async)
		s.notifier.NotifyNewChapter(manga.ID, manga.Title, int(extracted.ChapterNumber))
	}

	// Update manga's total_chapters to highest found
	s.db.Model(&manga).Update("total_chapters", highestChapter)

	*updateCount++
	return nil
}

// StartPollers starts all scheduled pollers in goroutines
func (s *SyncService) StartPollers(ctx context.Context) {
	// New manga poller: every 24 hours
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		log.Println("[Pollers] New manga poller started (interval: 24 hours)")

		// Run immediately on start
		if err := s.PollNewManga(ctx); err != nil {
			log.Printf("[Pollers] New manga poll error: %v", err)
		}

		for {
			select {
			case <-ticker.C:
				log.Println("[Pollers] Running new manga poll...")
				if err := s.PollNewManga(ctx); err != nil {
					log.Printf("[Pollers] New manga poll error: %v", err)
				}
			case <-ctx.Done():
				log.Println("[Pollers] New manga poller stopped")
				return
			}
		}
	}()

	// Chapter update poller: every 48 hours (2 days)
	go func() {
		ticker := time.NewTicker(48 * time.Hour)
		defer ticker.Stop()

		log.Println("[Pollers] Chapter update poller started (interval: 48 hours)")

		// Wait 1 hour before first run (let initial sync complete)
		time.Sleep(1 * time.Hour)

		if err := s.CheckChapterUpdates(ctx); err != nil {
			log.Printf("[Pollers] Chapter check error: %v", err)
		}

		for {
			select {
			case <-ticker.C:
				log.Println("[Pollers] Running chapter update check...")
				if err := s.CheckChapterUpdates(ctx); err != nil {
					log.Printf("[Pollers] Chapter check error: %v", err)
				}
			case <-ctx.Done():
				log.Println("[Pollers] Chapter update poller stopped")
				return
			}
		}
	}()
}
