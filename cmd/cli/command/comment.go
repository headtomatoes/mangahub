package command

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var commentCmd = &cobra.Command{
	Use:   "comment",
	Short: "Comment management commands",
	Long:  `Manage manga comments: create, update, delete, view, and list comments`,
}

var createCommentCmd = &cobra.Command{
	Use:   "create [manga-id] [content]",
	Short: "Create a comment on a manga",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		mangaID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid manga ID: %w", err)
		}

		content := strings.Join(args[1:], " ")

		httpClient := GetAuthenticatedClient()

		result, err := httpClient.CreateComment(mangaID, content)
		if err != nil {
			return fmt.Errorf("failed to create comment: %w", err)
		}

		fmt.Println("✓ Comment created successfully!")
		fmt.Printf("Manga ID: %d\n", mangaID)
		fmt.Printf("Posted by: %s\n", result.Username)
		fmt.Printf("Content: %s\n", result.Content)
		fmt.Printf("Created at: %s\n", result.CreatedAt.Format("2006-01-02 15:04:05"))

		return nil
	},
}

var updateCommentCmd = &cobra.Command{
	Use:   "update [comment-id] [content]",
	Short: "Update your comment",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		commentID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid comment ID: %w", err)
		}

		content := strings.Join(args[1:], " ")

		httpClient := GetAuthenticatedClient()

		result, err := httpClient.UpdateComment(commentID, content)
		if err != nil {
			return fmt.Errorf("failed to update comment: %w", err)
		}

		fmt.Println("✓ Comment updated successfully!")
		fmt.Printf("Comment ID: %d\n", commentID)
		fmt.Printf("Content: %s\n", result.Content)
		fmt.Printf("Updated at: %s\n", result.UpdatedAt.Format("2006-01-02 15:04:05"))

		return nil
	},
}

var deleteCommentCmd = &cobra.Command{
	Use:   "delete [comment-id]",
	Short: "Delete your comment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		commentID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid comment ID: %w", err)
		}

		httpClient := GetAuthenticatedClient()

		err = httpClient.DeleteComment(commentID)
		if err != nil {
			return fmt.Errorf("failed to delete comment: %w", err)
		}

		fmt.Printf("✓ Comment %d deleted successfully!\n", commentID)
		return nil
	},
}

var getCommentCmd = &cobra.Command{
	Use:   "get [comment-id]",
	Short: "Get a specific comment by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		commentID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid comment ID: %w", err)
		}

		httpClient := GetAuthenticatedClient()

		result, err := httpClient.GetCommentByID(commentID)
		if err != nil {
			return fmt.Errorf("failed to get comment: %w", err)
		}

		fmt.Printf("Comment ID: %d\n", commentID)
		fmt.Printf("User: %s\n", result.Username)
		fmt.Printf("Content: %s\n", result.Content)
		fmt.Printf("Created: %s\n", result.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Updated: %s\n", result.UpdatedAt.Format("2006-01-02 15:04:05"))

		return nil
	},
}

var listCommentsCmd = &cobra.Command{
	Use:   "list [manga-id]",
	Short: "List all comments for a manga",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mangaID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid manga ID: %w", err)
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		httpClient := GetAuthenticatedClient()

		result, err := httpClient.ListMangaComments(mangaID, page, pageSize)
		if err != nil {
			return fmt.Errorf("failed to list comments: %w", err)
		}

		if len(result.Data) == 0 {
			fmt.Println("No comments found for this manga.")
			return nil
		}

		fmt.Printf("Comments for manga %d (Page %d/%d, Total: %d):\n\n", mangaID, result.Page, result.TotalPages, result.Total)
		for _, c := range result.Data {
			fmt.Printf("User: %s\n", c.Username)
			fmt.Printf("Content: %s\n", c.Content)
			fmt.Printf("Posted: %s\n", c.CreatedAt.Format("2006-01-02 15:04:05"))
			if c.UpdatedAt.After(c.CreatedAt) {
				fmt.Printf("(Edited: %s)\n", c.UpdatedAt.Format("2006-01-02 15:04:05"))
			}
			fmt.Println(strings.Repeat("-", 50))
		}

		return nil
	},
}

var myCommentsCmd = &cobra.Command{
	Use:   "my-comments",
	Short: "List all your comments",
	RunE: func(cmd *cobra.Command, args []string) error {
		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		httpClient := GetAuthenticatedClient()

		result, err := httpClient.ListUserComments(page, pageSize)
		if err != nil {
			return fmt.Errorf("failed to list your comments: %w", err)
		}

		if len(result.Data) == 0 {
			fmt.Println("You haven't posted any comments yet.")
			return nil
		}

		fmt.Printf("Your comments (Page %d/%d, Total: %d):\n\n", result.Page, result.TotalPages, result.Total)
		for _, c := range result.Data {
			fmt.Printf("User: %s\n", c.Username)
			fmt.Printf("Content: %s\n", c.Content)
			fmt.Printf("Posted: %s\n", c.CreatedAt.Format("2006-01-02 15:04:05"))
			if c.UpdatedAt.After(c.CreatedAt) {
				fmt.Printf("(Edited: %s)\n", c.UpdatedAt.Format("2006-01-02 15:04:05"))
			}
			fmt.Println(strings.Repeat("-", 50))
		}

		return nil
	},
}

func init() {
	// Add subcommands
	commentCmd.AddCommand(createCommentCmd)
	commentCmd.AddCommand(updateCommentCmd)
	commentCmd.AddCommand(deleteCommentCmd)
	commentCmd.AddCommand(getCommentCmd)
	commentCmd.AddCommand(listCommentsCmd)
	commentCmd.AddCommand(myCommentsCmd)

	// Flags for list commands
	listCommentsCmd.Flags().Int("page", 1, "Page number")
	listCommentsCmd.Flags().Int("page-size", 20, "Number of comments per page")

	myCommentsCmd.Flags().Int("page", 1, "Page number")
	myCommentsCmd.Flags().Int("page-size", 20, "Number of comments per page")
}
