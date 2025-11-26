package command

import (
	"context"
	"fmt"
	"mangahub/cmd/cli/command/client"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

var (
	grpcAddress string
)

// grpcCmd represents the grpc command
var grpcCmd = &cobra.Command{
	Use:   "grpc",
	Short: "Interact with MangaHub via gRPC",
	Long:  `Use gRPC protocol to interact with MangaHub services for manga queries and progress updates.`,
}

// grpcMangaCmd represents the grpc manga command
var grpcMangaCmd = &cobra.Command{
	Use:   "manga",
	Short: "Manga operations via gRPC",
	Long:  `Query manga information using gRPC protocol.`,
}

// grpcMangaGetCmd retrieves manga by ID
var grpcMangaGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get manga details by ID",
	Long:  `Retrieve detailed information about a manga using its ID via gRPC.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		mangaIDStr, _ := cmd.Flags().GetString("id")
		if mangaIDStr == "" {
			return fmt.Errorf("manga ID is required (--id)")
		}

		mangaID, err := strconv.ParseInt(mangaIDStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid manga ID: %w", err)
		}

		// Create gRPC client
		grpcClient, err := client.NewGRPCClient(grpcAddress)
		if err != nil {
			return fmt.Errorf("failed to create gRPC client: %w", err)
		}
		defer grpcClient.Close()

		// Get manga
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		manga, err := grpcClient.GetManga(ctx, mangaID)
		if err != nil {
			return fmt.Errorf("failed to get manga: %w", err)
		}

		// Display manga details
		fmt.Printf("\n📖 Manga Details (via gRPC)\n")
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("ID:          %d\n", manga.Id)
		fmt.Printf("Title:       %s\n", manga.Title)
		fmt.Printf("Description: %s\n", manga.Description)
		if len(manga.Authors) > 0 {
			fmt.Printf("Authors:     %v\n", manga.Authors)
		}
		if len(manga.Genres) > 0 {
			fmt.Printf("Genres:      %v\n", manga.Genres)
		}
		if manga.CoverUrl != "" {
			fmt.Printf("Cover URL:   %s\n", manga.CoverUrl)
		}
		fmt.Printf("Chapters:    %d\n", manga.ChaptersCount)
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

		return nil
	},
}

// grpcMangaSearchCmd searches for manga
var grpcMangaSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search manga by query",
	Long:  `Search for manga using a search term via gRPC.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		query, _ := cmd.Flags().GetString("query")
		if query == "" {
			return fmt.Errorf("search query is required (--query)")
		}

		limit, _ := cmd.Flags().GetInt32("limit")
		offset, _ := cmd.Flags().GetInt32("offset")

		// Create gRPC client
		grpcClient, err := client.NewGRPCClient(grpcAddress)
		if err != nil {
			return fmt.Errorf("failed to create gRPC client: %w", err)
		}
		defer grpcClient.Close()

		// Search manga
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		mangas, totalCount, err := grpcClient.SearchManga(ctx, query, limit, offset)
		if err != nil {
			return fmt.Errorf("failed to search manga: %w", err)
		}

		// Display search results
		fmt.Printf("\n🔍 Search Results (via gRPC): \"%s\"\n", query)
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("Total matches: %d\n", totalCount)
		fmt.Printf("Showing: %d results (offset: %d)\n\n", len(mangas), offset)

		if len(mangas) == 0 {
			fmt.Println("No manga found matching your search.")
		} else {
			for i, manga := range mangas {
				fmt.Printf("%d. [ID: %d] %s\n", i+1, manga.Id, manga.Title)
				if manga.Description != "" && len(manga.Description) > 100 {
					fmt.Printf("   %s...\n", manga.Description[:100])
				} else if manga.Description != "" {
					fmt.Printf("   %s\n", manga.Description)
				}
				if len(manga.Genres) > 0 {
					fmt.Printf("   Genres: %v\n", manga.Genres)
				}
				fmt.Printf("   Chapters: %d\n\n", manga.ChaptersCount)
			}
		}

		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

		return nil
	},
}

// grpcProgressCmd represents the grpc progress command
var grpcProgressCmd = &cobra.Command{
	Use:   "progress",
	Short: "Progress operations via gRPC",
	Long:  `Update reading progress using gRPC protocol.`,
}

// grpcProgressUpdateCmd updates user progress
var grpcProgressUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update reading progress",
	Long:  `Update your reading progress for a manga via gRPC.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		mangaIDStr, _ := cmd.Flags().GetString("manga-id")
		if mangaIDStr == "" {
			return fmt.Errorf("manga ID is required (--manga-id)")
		}

		mangaID, err := strconv.ParseInt(mangaIDStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid manga ID: %w", err)
		}

		chapter, _ := cmd.Flags().GetInt32("chapter")
		if chapter <= 0 {
			return fmt.Errorf("chapter number must be positive (--chapter)")
		}

		status, _ := cmd.Flags().GetString("status")
		mangaTitle, _ := cmd.Flags().GetString("title")

		// Get user ID from stored credentials
		userID := GetCurrentUserID()
		if userID == "" {
			return fmt.Errorf("user ID not found, please login again")
		}

		// Create gRPC client
		grpcClient, err := client.NewGRPCClient(grpcAddress)
		if err != nil {
			return fmt.Errorf("failed to create gRPC client: %w", err)
		}
		defer grpcClient.Close()

		// Update progress
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = grpcClient.UpdateProgress(ctx, userID, mangaID, mangaTitle, chapter, status)
		if err != nil {
			return fmt.Errorf("failed to update progress: %w", err)
		}

		fmt.Printf("\n✅ Progress updated successfully (via gRPC)\n")
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("Manga ID: %d\n", mangaID)
		if mangaTitle != "" {
			fmt.Printf("Title:    %s\n", mangaTitle)
		}
		fmt.Printf("Chapter:  %d\n", chapter)
		if status != "" {
			fmt.Printf("Status:   %s\n", status)
		}
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

		return nil
	},
}

func init() {
	// Default gRPC address
	defaultGRPCAddr := "localhost:8083"
	if v := os.Getenv("MANGAHUB_GRPC_ADDR"); v != "" {
		defaultGRPCAddr = v
	}

	// grpc root command flags
	grpcCmd.PersistentFlags().StringVar(&grpcAddress, "grpc-addr", defaultGRPCAddr, "gRPC server address")

	// grpc manga get flags
	grpcMangaGetCmd.Flags().String("id", "", "Manga ID (required)")
	grpcMangaGetCmd.MarkFlagRequired("id")

	// grpc manga search flags
	grpcMangaSearchCmd.Flags().String("query", "", "Search query (required)")
	grpcMangaSearchCmd.Flags().Int32("limit", 20, "Maximum number of results")
	grpcMangaSearchCmd.Flags().Int32("offset", 0, "Result offset for pagination")
	grpcMangaSearchCmd.MarkFlagRequired("query")

	// grpc progress update flags
	grpcProgressUpdateCmd.Flags().String("manga-id", "", "Manga ID (required)")
	grpcProgressUpdateCmd.Flags().Int32("chapter", 0, "Chapter number (required)")
	grpcProgressUpdateCmd.Flags().String("status", "reading", "Reading status (reading, completed, on_hold)")
	grpcProgressUpdateCmd.Flags().String("title", "", "Manga title (optional)")
	grpcProgressUpdateCmd.MarkFlagRequired("manga-id")
	grpcProgressUpdateCmd.MarkFlagRequired("chapter")

	// Build command hierarchy
	grpcMangaCmd.AddCommand(grpcMangaGetCmd)
	grpcMangaCmd.AddCommand(grpcMangaSearchCmd)
	grpcProgressCmd.AddCommand(grpcProgressUpdateCmd)
	grpcCmd.AddCommand(grpcMangaCmd)
	grpcCmd.AddCommand(grpcProgressCmd)
}
