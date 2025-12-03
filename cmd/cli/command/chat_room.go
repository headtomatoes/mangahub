package command

import (
	"fmt"
	c "mangahub/cmd/cli/command/client"

	"github.com/spf13/cobra"
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Chat room related commands",
	Long:  `Commands to interact with chat rooms, send and receive messages in real-time.`,
}

var chatJoinCmd = &cobra.Command{
	Use:   "join",
	Short: "Join a manga chat room",
	RunE: func(cmd *cobra.Command, args []string) error {
		roomID, _ := cmd.Flags().GetString("room")

		if roomID == "" {
			return fmt.Errorf("--room is required (manga ID)")
		}

		// Get token from stored credentials if not provided
		token, _ := cmd.Flags().GetString("token")
		if token == "" {
			// Try to get from authentication
			token = accessToken
			if token == "" {
				return fmt.Errorf("not logged in, please run 'mangahub auth login' first or provide --token")
			}
		}

		return c.JoinChatRoom(roomID, token)
	},
}

func init() {
	chatCmd.AddCommand(chatJoinCmd)
	rootCmd.AddCommand(chatCmd)

	chatJoinCmd.Flags().StringP("room", "r", "", "Manga ID for the chat room (required)")
	chatJoinCmd.Flags().StringP("token", "t", "", "JWT token (optional, uses stored token if logged in)")
	chatJoinCmd.MarkFlagRequired("room")
}
