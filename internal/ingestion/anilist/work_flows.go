package anilist

import (
    "context"
    "fmt"
    "log"
    "time"
)

// RunInitialSync performs one-time bulk import of manga
// Fetches configured number of manga (default: 50) with complete metadata
func (s *SyncService) RunInitialSync(ctx context.Context) error {
    log.Println("[AniListSync] Starting initial sync...")

    // Check if already completed
    state, err := s.getSyncState("anilist_initial_sync")
    if err == nil && state.Status == "completed" {
        log.Println("[AniListSync] Initial sync already completed. Skipping.")
        return nil
    }

    // Update status to running
    if err := s.updateSyncState("anilist_initial_sync", "running", "", nil); err != nil {
        return fmt.Errorf("failed to update sync state: %w", err)
    }

    // Calculate pages needed
    perPage := 50 // AniList max per page
    totalToFetch := s.initialSyncLimit
    if totalToFetch == 0 {
        totalToFetch = getInitialSyncLimit()
    }

    totalPages := (totalToFetch + perPage - 1) / perPage
    log.Printf("[AniListSync] Fetching %d manga across %d pages...", totalToFetch, totalPages)

    // Create worker pool
    pool := NewWorkerPool(s.workerCount)
    pool.Start()

    successCount := 0
    errorCount := 0

    // Fetch and process pages
    for page := 1; page <= totalPages; page++ {
        select {
        case <-ctx.Done():
            log.Println("[AniListSync] Context cancelled, stopping sync...")
            pool.Shutdown()
            return ctx.Err()
        default:
        }

        log.Printf("[AniListSync] Fetching page %d/%d...", page, totalPages)

        // Fetch manga list
        response, err := s.client.GetManga(ctx, page, perPage)
        if err != nil {
            log.Printf("[AniListSync] ❌ Failed to fetch page %d: %v", page, err)
            errorCount++
            continue
        }

        log.Printf("[AniListSync] Processing %d manga from page %d...", len(response.Page.Media), page)

        // Submit tasks to worker pool
        for _, apiManga := range response.Page.Media {
            manga := apiManga // Capture for closure
            pool.Submit(func(ctx context.Context) error {
                if err := s.processManga(ctx, manga); err != nil {
                    log.Printf("[AniListSync] ❌ Failed to process manga %d: %v", manga.ID, err)
                    errorCount++
                    return err
                }

                // Notify about new manga
                s.notifier.NotifyNewManga(int64(manga.ID), manga.Title.English, manga.Title.Romaji)
                successCount++
                return nil
            })
        }

        // Don't fetch more than needed
        if successCount >= totalToFetch {
            break
        }

        // Respect rate limits between pages
        time.Sleep(time.Second)
    }

    // Wait for all workers to complete
    pool.Wait()

    log.Printf("[AniListSync] Initial sync completed: %d success, %d errors", successCount, errorCount)

    // Update sync state
    if err := s.updateSyncState("anilist_initial_sync", "completed", "", nil); err != nil {
        return fmt.Errorf("failed to update sync state: %w", err)
    }

    return nil
}

// PollNewManga checks for newly published manga on AniList
// Runs every 24 hours, detects manga updated since last poll
func (s *SyncService) PollNewManga(ctx context.Context) error {
    log.Println("[AniListSync] Polling for new manga...")

    // Update status to running
    if err := s.updateSyncState("anilist_new_manga_poll", "running", "", nil); err != nil {
        return fmt.Errorf("failed to update sync state: %w", err)
    }

    // Get last cursor (timestamp)
    state, err := s.getSyncState("anilist_new_manga_poll")
    var lastUpdate int64
    if err == nil && state.LastCursor != "" {
        fmt.Sscanf(state.LastCursor, "%d", &lastUpdate)
    } else {
        // Default to 24 hours ago
        lastUpdate = time.Now().Add(-24 * time.Hour).Unix()
    }

    log.Printf("[AniListSync] Checking for manga updated after %s", time.Unix(lastUpdate, 0).Format(time.RFC3339))

    // Create worker pool
    pool := NewWorkerPool(s.workerCount)
    pool.Start()

    successCount := 0
    errorCount := 0
    page := 1
    perPage := 50

    for {
        select {
        case <-ctx.Done():
            pool.Shutdown()
            return ctx.Err()
        default:
        }

        // Fetch recently updated manga
        response, err := s.client.GetRecentlyUpdated(ctx, lastUpdate, page, perPage)
        if err != nil {
            log.Printf("[AniListSync] ❌ Failed to fetch page %d: %v", page, err)
            break
        }

        if len(response.Page.Media) == 0 {
            log.Println("[AniListSync] No more new manga found")
            break
        }

        log.Printf("[AniListSync] Processing %d manga from page %d...", len(response.Page.Media), page)

        // Process manga
        for _, apiManga := range response.Page.Media {
            manga := apiManga
            pool.Submit(func(ctx context.Context) error {
                if err := s.processManga(ctx, manga); err != nil {
                    log.Printf("[AniListSync] ❌ Failed to process manga %d: %v", manga.ID, err)
                    errorCount++
                    return err
                }

                // Notify about new manga
                s.notifier.NotifyNewManga(int64(manga.ID), manga.Title.English, manga.Title.Romaji)
                successCount++
                return nil
            })
        }

        // Check if there are more pages
        if !response.Page.PageInfo.HasNextPage {
            break
        }

        page++
        time.Sleep(time.Second) // Rate limit between pages
    }

    pool.Wait()

    log.Printf("[AniListSync] Poll completed: %d new/updated manga, %d errors", successCount, errorCount)

    // Update sync state with current timestamp as cursor
    newCursor := fmt.Sprintf("%d", time.Now().Unix())
    if err := s.updateSyncState("anilist_new_manga_poll", "completed", newCursor, nil); err != nil {
        return fmt.Errorf("failed to update sync state: %w", err)
    }

    return nil
}

// CheckChapterUpdates checks for chapter count updates for tracked manga
// Runs every 48 hours, checks manga that haven't been checked recently
func (s *SyncService) CheckChapterUpdates(ctx context.Context) error {
    log.Println("[AniListSync] Checking for chapter updates...")

    // Update status to running
    if err := s.updateSyncState("anilist_chapter_check", "running", "", nil); err != nil {
        return fmt.Errorf("failed to update sync state: %w", err)
    }

    // Get manga that haven't been checked in 48 hours
    var mangaList []Manga
    checkThreshold := time.Now().Add(-48 * time.Hour)

    err := s.db.Where("anilist_id IS NOT NULL").
        Where("anilist_last_chapter_check IS NULL OR anilist_last_chapter_check < ?", checkThreshold).
        Limit(100). // Limit to avoid overwhelming the API
        Find(&mangaList).Error

    if err != nil {
        return fmt.Errorf("failed to fetch manga for update check: %w", err)
    }

    if len(mangaList) == 0 {
        log.Println("[AniListSync] No manga need chapter updates")
        if err := s.updateSyncState("anilist_chapter_check", "completed", "", nil); err != nil {
            return err
        }
        return nil
    }

    log.Printf("[AniListSync] Checking chapter updates for %d manga...", len(mangaList))

    // Create worker pool
    pool := NewWorkerPool(s.workerCount)
    pool.Start()

    successCount := 0
    errorCount := 0

    // Process each manga
    for _, manga := range mangaList {
        m := manga
        pool.Submit(func(ctx context.Context) error {
            if err := s.checkMangaChapters(ctx, &m); err != nil {
                log.Printf("[AniListSync] ❌ Failed to check chapters for manga %d: %v", m.ID, err)
                errorCount++
                return err
            }
            successCount++
            return nil
        })
    }

    pool.Wait()

    log.Printf("[AniListSync] Chapter check completed: %d checked, %d errors", successCount, errorCount)

    // Update sync state
    if err := s.updateSyncState("anilist_chapter_check", "completed", "", nil); err != nil {
        return fmt.Errorf("failed to update sync state: %w", err)
    }

    return nil
}

// checkMangaChapters checks a single manga for chapter count updates
func (s *SyncService) checkMangaChapters(ctx context.Context, manga *Manga) error {
    // Acquire rate semaphore
    s.rateSemaphore <- struct{}{}
    defer func() { <-s.rateSemaphore }()

    if manga.AniListID == nil {
        return fmt.Errorf("manga %d has no AniList ID", manga.ID)
    }

    // Fetch current data from AniList
    response, err := s.client.GetMangaByID(ctx, *manga.AniListID)
    if err != nil {
        return fmt.Errorf("failed to fetch manga from AniList: %w", err)
    }

    oldChapters := 0
    if manga.TotalChapters != nil {
        oldChapters = *manga.TotalChapters
    }

    newChapters := 0
    if response.Media.Chapters != nil {
        newChapters = *response.Media.Chapters
    }

    // Update chapter count if changed
    if newChapters != oldChapters && newChapters > 0 {
        now := time.Now()
        updates := map[string]interface{}{
            "total_chapters":               newChapters,
            "anilist_last_chapter_check":   &now,
            "anilist_last_synced_at":       &now,
        }

        // Update other metadata if changed
        if response.Media.Status != "" {
            status := mapAniListStatus(response.Media.Status)
            updates["status"] = &status
        }

        if response.Media.AverageScore != nil {
            rating := float64(*response.Media.AverageScore) / 10.0
            updates["average_rating"] = &rating
        }

        if err := s.db.Model(manga).Updates(updates).Error; err != nil {
            return fmt.Errorf("failed to update manga: %w", err)
        }

        log.Printf("[AniListSync] 📖 Chapter update: %s (%d → %d chapters)", manga.Title, oldChapters, newChapters)

        // Notify about chapter update
        s.notifier.NotifyChapterUpdate(manga.ID, manga.Title, oldChapters, newChapters)
    } else {
        // Just update the check timestamp
        now := time.Now()
        s.db.Model(manga).Update("anilist_last_chapter_check", &now)
    }

    return nil
}

// StartPollers starts all scheduled pollers in goroutines
func (s *SyncService) StartPollers(ctx context.Context) {
    log.Println("[AniListSync] Starting scheduled pollers...")

    // Poll for new manga every 24 hours
    go func() {
        ticker := time.NewTicker(24 * time.Hour)
        defer ticker.Stop()

        for {
            select {
            case <-ctx.Done():
                log.Println("[AniListSync] New manga poller stopped")
                return
            case <-ticker.C:
                if err := s.PollNewManga(ctx); err != nil {
                    log.Printf("[AniListSync] ❌ New manga poll failed: %v", err)
                }
            }
        }
    }()

    // Check for chapter updates every 48 hours
    go func() {
        ticker := time.NewTicker(48 * time.Hour)
        defer ticker.Stop()

        for {
            select {
            case <-ctx.Done():
                log.Println("[AniListSync] Chapter check poller stopped")
                return
            case <-ticker.C:
                if err := s.CheckChapterUpdates(ctx); err != nil {
                    log.Printf("[AniListSync] ❌ Chapter check failed: %v", err)
                }
            }
        }
    }()

    log.Println("[AniListSync] All pollers started")
}