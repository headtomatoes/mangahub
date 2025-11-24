package command

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var genreCmd = &cobra.Command{
	Use:   "genre",
	Short: "Genre management commands",
	Long:  `Manage genres: list all genres, create new genres, and find manga by genre`,
}

var listGenresCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available genres",
	RunE: func(cmd *cobra.Command, args []string) error {
		httpClient := GetAuthenticatedClient()

		genres, err := httpClient.GetAllGenres()
		if err != nil {
			return fmt.Errorf("failed to get genres: %w", err)
		}

		if len(genres) == 0 {
			fmt.Println("No genres found.")
			return nil
		}

		fmt.Printf("Available genres (%d total):\n\n", len(genres))
		for _, g := range genres {
			fmt.Printf("ID: %d | Name: %s\n", g.ID, g.Name)
		}

		return nil
	},
}

var createGenreCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new genre",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")

		httpClient := GetAuthenticatedClient()

		genre, err := httpClient.CreateGenre(name)
		if err != nil {
			return fmt.Errorf("failed to create genre: %w", err)
		}

		fmt.Println("✓ Genre created successfully!")
		fmt.Printf("ID: %d\n", genre.ID)
		fmt.Printf("Name: %s\n", genre.Name)

		return nil
	},
}

var mangaByGenreCmd = &cobra.Command{
	Use:   "mangas [genre-id]",
	Short: "Get all manga in a specific genre",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		genreID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid genre ID: %w", err)
		}

		httpClient := GetAuthenticatedClient()

		mangas, err := httpClient.GetMangasByGenre(genreID)
		if err != nil {
			return fmt.Errorf("failed to get manga by genre: %w", err)
		}

		if len(mangas) == 0 {
			fmt.Printf("No manga found for genre ID %d.\n", genreID)
			return nil
		}

		fmt.Printf("Manga in genre %d (%d total):\n\n", genreID, len(mangas))
		for _, m := range mangas {
			fmt.Printf("ID: %d\n", m.ID)
			fmt.Printf("Title: %s\n", m.Title)
			if m.Slug != nil {
				fmt.Printf("Slug: %s\n", *m.Slug)
			}
			if m.Author != nil {
				fmt.Printf("Author: %s\n", *m.Author)
			}
			if m.Status != nil {
				fmt.Printf("Status: %s\n", *m.Status)
			}
			if m.TotalChapters != nil {
				fmt.Printf("Chapters: %d\n", *m.TotalChapters)
			}
			fmt.Println(strings.Repeat("-", 50))
		}

		return nil
	},
}

func init() {
	// Add subcommands
	genreCmd.AddCommand(listGenresCmd)
	genreCmd.AddCommand(createGenreCmd)
	genreCmd.AddCommand(mangaByGenreCmd)
}
