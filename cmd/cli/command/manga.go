package command

import (
	"fmt"
	"mangahub/cmd/cli/authentication"
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
	Short: "List all manga",
	RunE: func(cmd *cobra.Command, args []string) error {
		creds, err := authentication.GetTokens()
		if err != nil {
			return fmt.Errorf("not logged in, please run 'mangahub auth login'")
		}

		httpClient := client.NewHTTPClient(apiURL)
		httpClient.SetToken(creds.AccessToken)

		mangas, err := httpClient.GetAllManga()
		if err != nil {
			return fmt.Errorf("failed to get manga list: %w", err)
		}

		if len(mangas) == 0 {
			fmt.Println("No manga found.")
			return nil
		}

		fmt.Printf("Found %d manga:\n\n", len(mangas))
		for _, m := range mangas {
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
			fmt.Println(strings.Repeat("-", 50))
		}
		return nil
	},
}

var getMangaCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Get manga by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		creds, err := authentication.GetTokens()
		if err != nil {
			return fmt.Errorf("not logged in, please run 'mangahub auth login'")
		}

		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid manga ID: %w", err)
		}

		httpClient := client.NewHTTPClient(apiURL)
		httpClient.SetToken(creds.AccessToken)

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
		creds, err := authentication.GetTokens()
		if err != nil {
			return fmt.Errorf("not logged in, please run 'mangahub auth login'")
		}

		query := strings.Join(args, " ")

		httpClient := client.NewHTTPClient(apiURL)
		httpClient.SetToken(creds.AccessToken)

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
		creds, err := authentication.GetTokens()
		if err != nil {
			return fmt.Errorf("not logged in, please run 'mangahub auth login'")
		}

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

		httpClient := client.NewHTTPClient(apiURL)
		httpClient.SetToken(creds.AccessToken)

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
		creds, err := authentication.GetTokens()
		if err != nil {
			return fmt.Errorf("not logged in, please run 'mangahub auth login'")
		}

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

		httpClient := client.NewHTTPClient(apiURL)
		httpClient.SetToken(creds.AccessToken)

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
		creds, err := authentication.GetTokens()
		if err != nil {
			return fmt.Errorf("not logged in, please run 'mangahub auth login'")
		}

		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid manga ID: %w", err)
		}

		httpClient := client.NewHTTPClient(apiURL)
		httpClient.SetToken(creds.AccessToken)

		err = httpClient.DeleteManga(id)
		if err != nil {
			return fmt.Errorf("failed to delete manga: %w", err)
		}

		fmt.Printf("✓ Manga %d deleted successfully!\n", id)
		return nil
	},
}

var genresCmd = &cobra.Command{
	Use:   "genres [manga-id]",
	Short: "Get genres for a manga",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		creds, err := authentication.GetTokens()
		if err != nil {
			return fmt.Errorf("not logged in, please run 'mangahub auth login'")
		}

		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid manga ID: %w", err)
		}

		httpClient := client.NewHTTPClient(apiURL)
		httpClient.SetToken(creds.AccessToken)

		genres, err := httpClient.GetMangaGenres(id)
		if err != nil {
			return fmt.Errorf("failed to get genres: %w", err)
		}

		if len(genres) == 0 {
			fmt.Println("No genres found for this manga.")
			return nil
		}

		fmt.Printf("Genres for manga %d:\n", id)
		for _, g := range genres {
			fmt.Printf("- %s (ID: %d)\n", g.Name, g.ID)
		}

		return nil
	},
}

var addGenresCmd = &cobra.Command{
	Use:   "add-genres [manga-id] [genre-ids...]",
	Short: "Add genres to a manga",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		creds, err := authentication.GetTokens()
		if err != nil {
			return fmt.Errorf("not logged in, please run 'mangahub auth login'")
		}

		mangaID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid manga ID: %w", err)
		}

		genreIDs := make([]int64, 0, len(args)-1)
		for _, arg := range args[1:] {
			gID, err := strconv.ParseInt(arg, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid genre ID '%s': %w", arg, err)
			}
			genreIDs = append(genreIDs, gID)
		}

		httpClient := client.NewHTTPClient(apiURL)
		httpClient.SetToken(creds.AccessToken)

		err = httpClient.AddMangaGenres(mangaID, genreIDs)
		if err != nil {
			return fmt.Errorf("failed to add genres: %w", err)
		}

		fmt.Printf("✓ Genres added to manga %d successfully!\n", mangaID)
		return nil
	},
}

var removeGenresCmd = &cobra.Command{
	Use:   "remove-genres [manga-id] [genre-ids...]",
	Short: "Remove genres from a manga",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		creds, err := authentication.GetTokens()
		if err != nil {
			return fmt.Errorf("not logged in, please run 'mangahub auth login'")
		}

		mangaID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid manga ID: %w", err)
		}

		genreIDs := make([]int64, 0, len(args)-1)
		for _, arg := range args[1:] {
			gID, err := strconv.ParseInt(arg, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid genre ID '%s': %w", arg, err)
			}
			genreIDs = append(genreIDs, gID)
		}

		httpClient := client.NewHTTPClient(apiURL)
		httpClient.SetToken(creds.AccessToken)

		err = httpClient.RemoveMangaGenres(mangaID, genreIDs)
		if err != nil {
			return fmt.Errorf("failed to remove genres: %w", err)
		}

		fmt.Printf("✓ Genres removed from manga %d successfully!\n", mangaID)
		return nil
	},
}

func init() {
	// Add subcommands
	mangaCmd.AddCommand(listMangaCmd)
	mangaCmd.AddCommand(getMangaCmd)
	mangaCmd.AddCommand(searchMangaCmd)
	mangaCmd.AddCommand(createMangaCmd)
	mangaCmd.AddCommand(updateMangaCmd)
	mangaCmd.AddCommand(deleteMangaCmd)
	mangaCmd.AddCommand(genresCmd)
	mangaCmd.AddCommand(addGenresCmd)
	mangaCmd.AddCommand(removeGenresCmd)

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
