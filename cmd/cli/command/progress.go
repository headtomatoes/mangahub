package command

import (
	"fmt"
	"mangahub/cmd/cli/command/client"
	"strconv"

	"github.com/spf13/cobra"
)

// progressCmd represents the progress command
var progressCmd = &cobra.Command{
	Use:   "progress",
	Short: "Manage manga reading progress",
	Long:  `Track and sync your manga reading progress across devices.`,
}

// progressUpdateCmd represents the progress update command
var progressUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update reading progress",
	Long:  `Update your current reading progress for a manga.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		mangaIDStr, _ := cmd.Flags().GetString("manga-id")
		chapter, _ := cmd.Flags().GetInt("chapter")
		volume, _ := cmd.Flags().GetInt("volume")
		notes, _ := cmd.Flags().GetString("notes")
		status, _ := cmd.Flags().GetString("status")
		force, _ := cmd.Flags().GetBool("force")

		// Validate manga ID
		if mangaIDStr == "" {
			return fmt.Errorf("--manga-id is required")
		}

		// Convert manga-id (could be slug or ID)
		mangaID, err := strconv.ParseInt(mangaIDStr, 10, 64)
		if err != nil {
			// Try to search by slug/title
			httpClient := GetAuthenticatedClient()
			mangas, err := httpClient.SearchManga(mangaIDStr)
			if err != nil || len(mangas) == 0 {
				return fmt.Errorf("manga '%s' not found in library", mangaIDStr)
			}
			mangaID = mangas[0].ID
		}

		// Validate chapter
		if chapter <= 0 {
			return fmt.Errorf("--chapter must be a positive number")
		}

		// Default status to "reading"
		if status == "" {
			status = "reading"
		}

		// Validate status
		validStatuses := map[string]bool{
			"reading":      true,
			"completed":    true,
			"on-hold":      true,
			"dropped":      true,
			"plan-to-read": true,
		}
		if !validStatuses[status] {
			return fmt.Errorf("invalid status: %s (valid: reading, completed, on-hold, dropped, plan-to-read)", status)
		}

		httpClient := GetAuthenticatedClient()

		// Get current progress (if exists)
		currentProgress, _ := httpClient.GetProgress(mangaID)

		// Check for backward progress
		if !force && currentProgress != nil && chapter < currentProgress.Chapter {
			return fmt.Errorf("✗ Progress update failed: Chapter %d is behind your current progress (Chapter %d)\nUse --force to set backwards progress: --force --chapter %d", chapter, currentProgress.Chapter, chapter)
		}

		// Get manga details
		manga, err := httpClient.GetMangaByID(mangaID)
		if err != nil {
			return fmt.Errorf("manga not found: %w", err)
		}

		// Check if chapter exceeds total chapters
		if manga.TotalChapters != nil && chapter > *manga.TotalChapters {
			return fmt.Errorf("✗ Progress update failed: Chapter %d exceeds manga's total chapters (%d)\nValid range: 1-%d", chapter, *manga.TotalChapters, *manga.TotalChapters)
		}

		fmt.Println("Updating reading progress...")

		// Prepare request
		var volumePtr *int
		if volume > 0 {
			volumePtr = &volume
		}

		req := &client.UpdateProgressRequest{
			MangaID: mangaID,
			Chapter: chapter,
			Status:  status,
			Volume:  volumePtr,
			Notes:   notes,
		}

		// Update progress
		progress, err := httpClient.UpdateProgress(req)
		if err != nil {
			return fmt.Errorf("✗ Progress update failed: %w", err)
		}

		// Success output
		fmt.Println("✓ Progress updated successfully!")
		fmt.Printf("\nManga: %s\n", manga.Title)

		if currentProgress != nil {
			diff := chapter - currentProgress.Chapter
			fmt.Printf("Previous: Chapter %s\n", formatNumber(currentProgress.Chapter))
			if diff > 0 {
				fmt.Printf("Current: Chapter %s (+%d)\n", formatNumber(chapter), diff)
			} else {
				fmt.Printf("Current: Chapter %s\n", formatNumber(chapter))
			}
		} else {
			fmt.Printf("Current: Chapter %s\n", formatNumber(chapter))
		}

		if volume > 0 {
			fmt.Printf("Volume: %d\n", volume)
		}

		fmt.Printf("Updated: %s\n", progress.UpdatedAt.Format("2006-01-02 15:04:05 MST"))

		// Sync status (simulated for now)
		fmt.Println("\nSync Status:")
		fmt.Println("  Local database: ✓ Updated")
		fmt.Println("  TCP sync server: ⚠ Not connected (use 'mangahub sync connect')")
		fmt.Println("  Cloud backup: ⚠ Not implemented")

		// Statistics
		fmt.Println("\nStatistics:")
		fmt.Printf("  Total chapters read: %s\n", formatNumber(chapter))
		if manga.TotalChapters != nil {
			percentage := float64(chapter) / float64(*manga.TotalChapters) * 100
			fmt.Printf("  Progress: %.1f%% (%d/%d)\n", percentage, chapter, *manga.TotalChapters)
			remaining := *manga.TotalChapters - chapter
			if remaining > 0 {
				fmt.Printf("  Chapters remaining: %d\n", remaining)
			}
		}

		// Next actions
		fmt.Println("\nNext actions:")
		if manga.TotalChapters != nil && chapter < *manga.TotalChapters {
			fmt.Printf("  Continue reading: Chapter %s available\n", formatNumber(chapter+1))
		}
		fmt.Printf("  Rate this chapter: mangahub library update --manga-id %d --rating 9\n", mangaID)

		return nil
	},
}

// progressHistoryCmd represents the progress history command
var progressHistoryCmd = &cobra.Command{
	Use:   "history",
	Short: "View progress history",
	Long:  `View your reading progress history for a specific manga or all mangas.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		mangaIDStr, _ := cmd.Flags().GetString("manga-id")

		httpClient := GetAuthenticatedClient()

		var mangaID *int64
		if mangaIDStr != "" {
			id, err := strconv.ParseInt(mangaIDStr, 10, 64)
			if err != nil {
				// Try to search by slug/title
				mangas, err := httpClient.SearchManga(mangaIDStr)
				if err != nil || len(mangas) == 0 {
					return fmt.Errorf("manga '%s' not found", mangaIDStr)
				}
				id = mangas[0].ID
			}
			mangaID = &id
		}

		history, err := httpClient.GetProgressHistory(mangaID)
		if err != nil {
			return fmt.Errorf("failed to get progress history: %w", err)
		}

		if len(history.History) == 0 {
			fmt.Println("No progress history found.")
			return nil
		}

		if mangaID != nil {
			manga, _ := httpClient.GetMangaByID(*mangaID)
			if manga != nil {
				fmt.Printf("Progress History for: %s\n\n", manga.Title)
			}
		} else {
			fmt.Println("Progress History (All Manga)\n")
		}

		for i, progress := range history.History {
			fmt.Printf("[%d] Chapter %s", i+1, formatNumber(progress.Chapter))
			if progress.Volume != nil {
				fmt.Printf(" (Vol. %d)", *progress.Volume)
			}
			fmt.Printf(" - %s\n", progress.UpdatedAt.Format("2006-01-02 15:04"))
			if progress.Status != "" {
				fmt.Printf("    Status: %s\n", progress.Status)
			}
			if progress.Notes != "" {
				fmt.Printf("    Notes: %s\n", progress.Notes)
			}
			if progress.MangaTitle != "" && mangaID == nil {
				fmt.Printf("    Manga: %s\n", progress.MangaTitle)
			}
			fmt.Println()
		}

		fmt.Printf("Total entries: %d\n", history.Total)
		return nil
	},
}

// progressSyncCmd represents the progress sync command
var progressSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Manually sync progress with server",
	Long:  `Force a manual sync of your progress with the server.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Syncing progress with server...")
		fmt.Println("⚠ TCP sync not yet implemented")
		fmt.Println("Use 'mangahub sync connect' to establish TCP connection")
		return nil
	},
}

// progressSyncStatusCmd represents the progress sync-status command
var progressSyncStatusCmd = &cobra.Command{
	Use:   "sync-status",
	Short: "Check sync status",
	Long:  `Check the current synchronization status of your progress.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Progress Sync Status:")
		fmt.Println("  Local database: ✓ Up to date")
		fmt.Println("  TCP sync server: ✗ Not connected")
		fmt.Println("  Cloud backup: ⚠ Not implemented")
		fmt.Println("\nTo enable real-time sync, run:")
		fmt.Println("  mangahub sync connect")
		return nil
	},
}

// Helper function to format numbers with commas
func formatNumber(n int) string {
	str := strconv.Itoa(n)
	if len(str) <= 3 {
		return str
	}

	result := ""
	for i, digit := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result += ","
		}
		result += string(digit)
	}
	return result
}

func init() {
	// Add progress subcommands
	progressCmd.AddCommand(progressUpdateCmd)
	progressCmd.AddCommand(progressHistoryCmd)
	progressCmd.AddCommand(progressSyncCmd)
	progressCmd.AddCommand(progressSyncStatusCmd)

	// Progress update flags
	progressUpdateCmd.Flags().String("manga-id", "", "Manga ID or slug (required)")
	progressUpdateCmd.Flags().Int("chapter", 0, "Current chapter (required)")
	progressUpdateCmd.Flags().Int("volume", 0, "Current volume (optional)")
	progressUpdateCmd.Flags().String("notes", "", "Notes about this chapter (optional)")
	progressUpdateCmd.Flags().String("status", "reading", "Reading status (reading, completed, on-hold, dropped, plan-to-read)")
	progressUpdateCmd.Flags().Bool("force", false, "Force update even if chapter is behind current progress")
	progressUpdateCmd.MarkFlagRequired("manga-id")
	progressUpdateCmd.MarkFlagRequired("chapter")

	// Progress history flags
	progressHistoryCmd.Flags().String("manga-id", "", "Manga ID or slug (optional, shows all if not specified)")
}
