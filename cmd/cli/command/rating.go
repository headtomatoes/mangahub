package command

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var ratingCmd = &cobra.Command{
	Use:   "rating",
	Short: "Rating management commands",
	Long:  `Manage manga ratings: create/update, view, delete, and list ratings`,
}

var rateCmd = &cobra.Command{
	Use:   "rate [manga-id] [rating]",
	Short: "Rate a manga (1-10)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		mangaID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid manga ID: %w", err)
		}

		rating, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid rating: %w", err)
		}

		if rating < 1 || rating > 10 {
			return fmt.Errorf("rating must be between 1 and 10")
		}

		httpClient := GetAuthenticatedClient()

		result, err := httpClient.CreateOrUpdateRating(mangaID, rating)
		if err != nil {
			return fmt.Errorf("failed to rate manga: %w", err)
		}

		fmt.Println("✓ Rating submitted successfully!")
		fmt.Printf("Manga ID: %d\n", mangaID)
		fmt.Printf("Your Rating: %d/10\n", result.Rating)
		fmt.Printf("Updated at: %s\n", result.UpdatedAt.Format("2006-01-02 15:04:05"))

		return nil
	},
}

var getRatingCmd = &cobra.Command{
	Use:   "get [manga-id]",
	Short: "Get your rating for a manga",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mangaID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid manga ID: %w", err)
		}

		httpClient := GetAuthenticatedClient()

		result, err := httpClient.GetUserRating(mangaID)
		if err != nil {
			return fmt.Errorf("failed to get rating: %w", err)
		}

		fmt.Printf("Your rating for manga %d:\n", mangaID)
		fmt.Printf("Rating: %d/10\n", result.Rating)
		fmt.Printf("Created at: %s\n", result.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Updated at: %s\n", result.UpdatedAt.Format("2006-01-02 15:04:05"))

		return nil
	},
}

var deleteRatingCmd = &cobra.Command{
	Use:   "delete [manga-id]",
	Short: "Delete your rating for a manga",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mangaID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid manga ID: %w", err)
		}

		httpClient := GetAuthenticatedClient()

		err = httpClient.DeleteRating(mangaID)
		if err != nil {
			return fmt.Errorf("failed to delete rating: %w", err)
		}

		fmt.Printf("✓ Rating deleted successfully for manga %d!\n", mangaID)
		return nil
	},
}

var listRatingsCmd = &cobra.Command{
	Use:   "list [manga-id]",
	Short: "List all ratings for a manga",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mangaID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid manga ID: %w", err)
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		httpClient := GetAuthenticatedClient()

		result, err := httpClient.ListMangaRatings(mangaID, page, pageSize)
		if err != nil {
			return fmt.Errorf("failed to list ratings: %w", err)
		}

		if len(result.Data) == 0 {
			fmt.Println("No ratings found for this manga.")
			return nil
		}

		fmt.Printf("Ratings for manga %d (Page %d/%d, Total: %d):\n\n", mangaID, result.Page, result.TotalPages, result.Total)
		for _, r := range result.Data {
			fmt.Printf("User: %s\n", r.Username)
			fmt.Printf("Rating: %d/10\n", r.Rating)
			fmt.Printf("Created: %s\n", r.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Println(strings.Repeat("-", 50))
		}

		return nil
	},
}

var averageRatingCmd = &cobra.Command{
	Use:   "average [manga-id]",
	Short: "Get average rating for a manga",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mangaID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid manga ID: %w", err)
		}

		httpClient := GetAuthenticatedClient()

		result, err := httpClient.GetAverageRating(mangaID)
		if err != nil {
			return fmt.Errorf("failed to get average rating: %w", err)
		}

		fmt.Printf("Average rating for manga %d:\n", mangaID)
		fmt.Printf("Average: %.2f/10\n", result.AverageRating)
		fmt.Printf("Total ratings: %d\n", result.TotalRatings)

		return nil
	},
}

func init() {
	// Add subcommands
	ratingCmd.AddCommand(rateCmd)
	ratingCmd.AddCommand(getRatingCmd)
	ratingCmd.AddCommand(deleteRatingCmd)
	ratingCmd.AddCommand(listRatingsCmd)
	ratingCmd.AddCommand(averageRatingCmd)

	// Flags for list command
	listRatingsCmd.Flags().Int("page", 1, "Page number")
	listRatingsCmd.Flags().Int("page-size", 20, "Number of ratings per page")
}
