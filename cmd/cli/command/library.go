package command

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var libraryCmd = &cobra.Command{
	Use:   "library",
	Short: "Manage your manga library",
	Long:  `Add, remove, and list manga in your personal library`,
}

var libraryAddCmd = &cobra.Command{
	Use:   "add [manga_id]",
	Short: "Add a manga to your library",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mangaID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fmt.Println("Invalid manga ID:", err)
			return
		}

		httpClient := GetAuthenticatedClient()

		if err := httpClient.AddToLibrary(mangaID); err != nil {
			fmt.Println("Failed to add manga to library:", err)
			return
		}

		fmt.Printf("✅ Successfully added manga (ID: %d) to your library\n", mangaID)
	},
}

var libraryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all manga in your library",
	Run: func(cmd *cobra.Command, args []string) {
		httpClient := GetAuthenticatedClient()

		library, err := httpClient.GetLibrary()
		if err != nil {
			fmt.Println("Failed to fetch library:", err)
			return
		}

		if len(library.Items) == 0 {
			fmt.Println("📚 Your library is empty")
			return
		}

		fmt.Printf("📚 Your Library (%d manga)\n", library.Total)
		fmt.Println("─────────────────────────────────────────────────────────")
		for i, item := range library.Items {
			fmt.Printf("%d. %s (ID: %d)\n", i+1, item.Manga.Title, item.MangaID)
			if item.Manga.Author != nil {
				fmt.Printf("   Author: %s\n", *item.Manga.Author)
			}
			if item.Manga.Status != nil {
				fmt.Printf("   Status: %s\n", *item.Manga.Status)
			}
			fmt.Printf("   Added: %s\n", item.AddedAt.Format("2006-01-02 15:04"))
			fmt.Println()
		}
	},
}

var libraryRemoveCmd = &cobra.Command{
	Use:   "remove [manga_id]",
	Short: "Remove a manga from your library",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mangaID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fmt.Println("Invalid manga ID:", err)
			return
		}

		httpClient := GetAuthenticatedClient()

		if err := httpClient.RemoveFromLibrary(mangaID); err != nil {
			fmt.Println("Failed to remove manga from library:", err)
			return
		}

		fmt.Printf("✅ Successfully removed manga (ID: %d) from your library\n", mangaID)
	},
}

func init() {
	libraryCmd.AddCommand(libraryAddCmd)
	libraryCmd.AddCommand(libraryListCmd)
	libraryCmd.AddCommand(libraryRemoveCmd)
}
