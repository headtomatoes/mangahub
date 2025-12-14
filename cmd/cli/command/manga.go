package command

import (
	"fmt"
	"mangahub/cmd/cli/command/client"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var mangaCmd = &cobra.Command{
	Use:   "manga",
	Short: "Manga management commands",
	Long:  `Manage manga: list, view, search, create, update, delete, and manage genres`,
}

var listMangaCmd = &cobra.Command{
	Use:   "list",
	Short: "List all manga with pagination",
	Long:  "List manga with optional page and page_size parameters",
	RunE: func(cmd *cobra.Command, args []string) error {
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		httpClient := GetAuthenticatedClient()

		result, err := httpClient.GetAllMangaPaginated(page, pageSize)
		if err != nil {
			return fmt.Errorf("failed to get manga list: %w", err)
		}

		if len(result.Data) == 0 {
			fmt.Println("No manga found.")
			return nil
		}

		fmt.Printf("Found %d manga (Page %d/%d, Total: %d):\n\n",
			len(result.Data), result.Page, result.TotalPages, result.Total)

		for _, m := range result.Data {
			fmt.Printf("ID: %d\n", m.ID)
			fmt.Printf("Title: %s\n", m.Title)
			if m.Author != nil {
				fmt.Printf("Author: %s\n", *m.Author)
			}
			if m.Status != nil {
				fmt.Printf("Status: %s\n", *m.Status)
			}
			if m.TotalChapters != nil {
				fmt.Printf("Chapters: %d\n", *m.TotalChapters)
			}
			// if m.AverageRating != nil {
			// 	fmt.Printf("Rating: %.2f\n", *m.AverageRating)
			// }
			fmt.Println(strings.Repeat("-", 50))
		}

		// Show pagination info
		fmt.Printf("\nPage %d of %d (Total: %d manga)\n", result.Page, result.TotalPages, result.Total)
		if result.Page < result.TotalPages {
			fmt.Printf("Use --page %d to see next page\n", result.Page+1)
		}

		return nil
	},
}

var getMangaCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Get manga by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid manga ID: %w", err)
		}

		httpClient := GetAuthenticatedClient()

		manga, err := httpClient.GetMangaByID(id)
		if err != nil {
			return fmt.Errorf("failed to get manga: %w", err)
		}

		fmt.Printf("ID: %d\n", manga.ID)
		fmt.Printf("Title: %s\n", manga.Title)
		if manga.Slug != nil {
			fmt.Printf("Slug: %s\n", *manga.Slug)
		}
		if manga.Author != nil {
			fmt.Printf("Author: %s\n", *manga.Author)
		}
		if manga.Status != nil {
			fmt.Printf("Status: %s\n", *manga.Status)
		}
		if manga.TotalChapters != nil {
			fmt.Printf("Total Chapters: %d\n", *manga.TotalChapters)
		}
		if manga.Description != nil {
			fmt.Printf("Description: %s\n", *manga.Description)
		}
		if manga.CoverURL != nil {
			fmt.Printf("Cover URL: %s\n", *manga.CoverURL)
		}
		if manga.CreatedAt != nil {
			fmt.Printf("Created At: %s\n", manga.CreatedAt.Format("2006-01-02 15:04:05"))
		}

		return nil
	},
}

var searchMangaCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search manga by title, author, or slug",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.Join(args, " ")

		httpClient := GetAuthenticatedClient()

		mangas, err := httpClient.SearchManga(query)
		if err != nil {
			return fmt.Errorf("search failed: %w", err)
		}

		if len(mangas) == 0 {
			fmt.Println("No manga found matching your search.")
			return nil
		}

		fmt.Printf("Found %d manga matching '%s':\n\n", len(mangas), query)
		for _, m := range mangas {
			fmt.Printf("ID: %d\n", m.ID)
			fmt.Printf("Title: %s\n", m.Title)
			if m.Author != nil {
				fmt.Printf("Author: %s\n", *m.Author)
			}
			if m.Status != nil {
				fmt.Printf("Status: %s\n", *m.Status)
			}
			fmt.Println(strings.Repeat("-", 50))
		}
		return nil
	},
}

var createMangaCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new manga",
	RunE: func(cmd *cobra.Command, args []string) error {
		title, _ := cmd.Flags().GetString("title")
		author, _ := cmd.Flags().GetString("author")
		status, _ := cmd.Flags().GetString("status")
		chapters, _ := cmd.Flags().GetInt("chapters")
		description, _ := cmd.Flags().GetString("description")
		coverURL, _ := cmd.Flags().GetString("cover-url")
		slug, _ := cmd.Flags().GetString("slug")

		request := &client.CreateMangaRequest{
			Title: title,
		}

		if author != "" {
			request.Author = &author
		}
		if status != "" {
			request.Status = &status
		}
		if chapters > 0 {
			request.TotalChapters = &chapters
		}
		if description != "" {
			request.Description = &description
		}
		if coverURL != "" {
			request.CoverURL = &coverURL
		}
		if slug != "" {
			request.Slug = &slug
		}

		httpClient := GetAuthenticatedClient()

		manga, err := httpClient.CreateManga(request)
		if err != nil {
			return fmt.Errorf("failed to create manga: %w", err)
		}

		fmt.Println("✓ Manga created successfully!")
		fmt.Printf("ID: %d\n", manga.ID)
		fmt.Printf("Title: %s\n", manga.Title)
		if manga.Slug != nil {
			fmt.Printf("Slug: %s\n", *manga.Slug)
		}

		return nil
	},
}

var updateMangaCmd = &cobra.Command{
	Use:   "update [id]",
	Short: "Update an existing manga",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid manga ID: %w", err)
		}

		request := &client.UpdateMangaRequest{}

		title, _ := cmd.Flags().GetString("title")
		author, _ := cmd.Flags().GetString("author")
		status, _ := cmd.Flags().GetString("status")
		chapters, _ := cmd.Flags().GetInt("chapters")
		description, _ := cmd.Flags().GetString("description")
		coverURL, _ := cmd.Flags().GetString("cover-url")
		slug, _ := cmd.Flags().GetString("slug")

		if title != "" {
			request.Title = &title
		}
		if author != "" {
			request.Author = &author
		}
		if status != "" {
			request.Status = &status
		}
		if chapters >= 0 {
			request.TotalChapters = &chapters
		}
		if description != "" {
			request.Description = &description
		}
		if coverURL != "" {
			request.CoverURL = &coverURL
		}
		if slug != "" {
			request.Slug = &slug
		}

		httpClient := GetAuthenticatedClient()

		manga, err := httpClient.UpdateManga(id, request)
		if err != nil {
			return fmt.Errorf("failed to update manga: %w", err)
		}

		fmt.Println("✓ Manga updated successfully!")
		fmt.Printf("ID: %d\n", manga.ID)
		fmt.Printf("Title: %s\n", manga.Title)

		return nil
	},
}

var deleteMangaCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete a manga",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid manga ID: %w", err)
		}

		httpClient := GetAuthenticatedClient()

		err = httpClient.DeleteManga(id)
		if err != nil {
			return fmt.Errorf("failed to delete manga: %w", err)
		}

		fmt.Printf("✓ Manga %d deleted successfully!\n", id)
		return nil
	},
}

// var genresCmd = &cobra.Command{
// 	Use:   "genres [manga-id]",
// 	Short: "Get genres for a manga",
// 	Args:  cobra.ExactArgs(1),
// 	RunE: func(cmd *cobra.Command, args []string) error {
// 		id, err := strconv.ParseInt(args[0], 10, 64)
// 		if err != nil {
// 			return fmt.Errorf("invalid manga ID: %w", err)
// 		}

// 		httpClient := GetAuthenticatedClient()

// 		genres, err := httpClient.GetMangaGenres(id)
// 		if err != nil {
// 			return fmt.Errorf("failed to get genres: %w", err)
// 		}

// 		if len(genres) == 0 {
// 			fmt.Println("No genres found for this manga.")
// 			return nil
// 		}

// 		fmt.Printf("Genres for manga %d:\n", id)
// 		for _, g := range genres {
// 			fmt.Printf("- %s (ID: %d)\n", g.Name, g.ID)
// 		}

// 		return nil
// 	},
// }

// var addGenresCmd = &cobra.Command{
// 	Use:   "add-genres [manga-id] [genre-ids...]",
// 	Short: "Add genres to a manga",
// 	Args:  cobra.MinimumNArgs(2),
// 	RunE: func(cmd *cobra.Command, args []string) error {
// 		mangaID, err := strconv.ParseInt(args[0], 10, 64)
// 		if err != nil {
// 			return fmt.Errorf("invalid manga ID: %w", err)
// 		}

// 		genreIDs := make([]int64, 0, len(args)-1)
// 		for _, arg := range args[1:] {
// 			gID, err := strconv.ParseInt(arg, 10, 64)
// 			if err != nil {
// 				return fmt.Errorf("invalid genre ID '%s': %w", arg, err)
// 			}
// 			genreIDs = append(genreIDs, gID)
// 		}

// 		httpClient := GetAuthenticatedClient()

// 		err = httpClient.AddMangaGenres(mangaID, genreIDs)
// 		if err != nil {
// 			return fmt.Errorf("failed to add genres: %w", err)
// 		}

// 		fmt.Printf("✓ Genres added to manga %d successfully!\n", mangaID)
// 		return nil
// 	},
// }

// var removeGenresCmd = &cobra.Command{
// 	Use:   "remove-genres [manga-id] [genre-ids...]",
// 	Short: "Remove genres from a manga",
// 	Args:  cobra.MinimumNArgs(2),
// 	RunE: func(cmd *cobra.Command, args []string) error {
// 		mangaID, err := strconv.ParseInt(args[0], 10, 64)
// 		if err != nil {
// 			return fmt.Errorf("invalid manga ID: %w", err)
// 		}

// 		genreIDs := make([]int64, 0, len(args)-1)
// 		for _, arg := range args[1:] {
// 			gID, err := strconv.ParseInt(arg, 10, 64)
// 			if err != nil {
// 				return fmt.Errorf("invalid genre ID '%s': %w", arg, err)
// 			}
// 			genreIDs = append(genreIDs, gID)
// 		}

// 		httpClient := GetAuthenticatedClient()

// 		err = httpClient.RemoveMangaGenres(mangaID, genreIDs)
// 		if err != nil {
// 			return fmt.Errorf("failed to remove genres: %w", err)
// 		}

// 		fmt.Printf("✓ Genres removed from manga %d successfully!\n", mangaID)
// 		return nil
// 	},
// }

func init() {
	// Add subcommands
	mangaCmd.AddCommand(listMangaCmd)
	mangaCmd.AddCommand(getMangaCmd)
	mangaCmd.AddCommand(searchMangaCmd)
	mangaCmd.AddCommand(createMangaCmd)
	mangaCmd.AddCommand(updateMangaCmd)
	mangaCmd.AddCommand(deleteMangaCmd)

	// List command flags
	listMangaCmd.Flags().Int("page", 1, "Page number (default: 1)")
	listMangaCmd.Flags().Int("page-size", 20, "Number of items per page (default: 20, max: 100)")

	// Create flags
	createMangaCmd.Flags().String("title", "", "Manga title (required)")
	createMangaCmd.Flags().String("author", "", "Manga author")
	createMangaCmd.Flags().String("status", "", "Manga status (ongoing/completed/hiatus)")
	createMangaCmd.Flags().Int("chapters", 0, "Total chapters")
	createMangaCmd.Flags().String("description", "", "Manga description")
	createMangaCmd.Flags().String("cover-url", "", "Cover image URL")
	createMangaCmd.Flags().String("slug", "", "URL slug")
	createMangaCmd.MarkFlagRequired("title")

	// Update flags
	updateMangaCmd.Flags().String("title", "", "Manga title")
	updateMangaCmd.Flags().String("author", "", "Manga author")
	updateMangaCmd.Flags().String("status", "", "Manga status (ongoing/completed/hiatus)")
	updateMangaCmd.Flags().Int("chapters", -1, "Total chapters")
	updateMangaCmd.Flags().String("description", "", "Manga description")
	updateMangaCmd.Flags().String("cover-url", "", "Cover image URL")
	updateMangaCmd.Flags().String("slug", "", "URL slug")
}
