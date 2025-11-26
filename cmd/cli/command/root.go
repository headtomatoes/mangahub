package command

// root.go defines the root command for the mangahubCLI application.
// set up the global flags and configuration here.

import (
	"fmt"
	"mangahub/cmd/cli/authentication"
	"mangahub/cmd/cli/command/client"
	"mangahub/cmd/cli/dto"
	"os"

	"time"

	"github.com/spf13/cobra"
)

const refreshBuffer = 5 * 60 // 5 minutes in seconds

var (
	apiURL           string                                                                         //Global flag for API server URL
	skipAuthCommands = map[string]bool{"auth": true, "help": true, "completion": true, "grpc": true} // Commands that skip authentication
	accessToken      string                                                                         // Global variable to hold the access token for the session
	currentUsername  string                                                                         // Global variable to hold the current username
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mangahubCLI",
	Short: "mangahubCLI - MangaHub Command Line Interface",
	Long: `mangahubCLI is a tool for user to interact with mangahub API. This application is
built for learning purpose and personal use. User can use this application to:
- Search and track manga
- Sync reading progress
- Reveice realtime notifications for new chapters
- Join community discussions chat rooms

Use "mangahubCLI command -help" or "mangahubCLI command -h" to see all available commands.`,
	// PersistentPreRun runs before every command except those in skipAuthCommands
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip authentication for auth commands and help
		cmdName := cmd.Name() // get current command name e.g., "manga"
		if cmd.Parent() != nil {
			cmdName = cmd.Parent().Name()
		}

		if skipAuthCommands[cmdName] { // skip if command in the skip list
			return nil
		}

		// Try to get stored credentials
		creds, err := authentication.GetTokens()
		if err != nil {
			return fmt.Errorf("not logged in, please run 'mangahubCLI auth login' first")
		}

		// Check if token needs refresh
		now := time.Now().Unix()
		if now >= (creds.ExpiresAt - refreshBuffer) { // now >= expiry - buffer time
			httpClient := client.NewHTTPClient(apiURL)
			req := &dto.RefreshTokenRequest{RefreshToken: creds.RefreshToken}

			refreshResp, err := httpClient.RefreshToken(req)
			if err != nil {
				// Clear invalid tokens
				_ = authentication.DeleteTokens()
				return fmt.Errorf("session expired, please login again: %w", err)
			}

			// Store refreshed tokens
			err = authentication.StoreTokens(&authentication.StoredCredentials{
				AccessToken:  refreshResp.AccessToken,
				RefreshToken: refreshResp.RefreshToken,
				Username:     creds.Username,
				ExpiresAt:    now + refreshResp.ExpiresIn,
			})
			if err != nil {
				return fmt.Errorf("failed to store refreshed tokens: %w", err)
			}

			// Update global variables
			accessToken = refreshResp.AccessToken
			currentUsername = creds.Username
		} else {
			// Use existing valid token
			accessToken = creds.AccessToken
			currentUsername = creds.Username
		}

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err) // Print error to standard error
		os.Exit(1)
	}
}

func init() {
	// Global persistent flags = available to all subcommands
	defaultURL := "http://localhost:8084"
	if v := os.Getenv("MANGAHUB_API_URL"); v != "" {
		defaultURL = v
	}
	apiURL = defaultURL
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", defaultURL, "MangaHub API server URL")
	// Add subcommands
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(mangaCmd)
	rootCmd.AddCommand(libraryCmd)
	rootCmd.AddCommand(progressCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(ratingCmd)
	rootCmd.AddCommand(commentCmd)
	rootCmd.AddCommand(genreCmd)
	rootCmd.AddCommand(grpcCmd)
}

// GetAuthenticatedClient returns an HTTP client with the current access token
// Helper for commands that need authentication
func GetAuthenticatedClient() *client.HTTPClient {
	httpClient := client.NewHTTPClient(apiURL)
	if accessToken != "" {
		httpClient.SetToken(accessToken)
	}
	return httpClient
}

// GetCurrentUsername returns the username of the currently logged-in user
func GetCurrentUsername() string {
	return currentUsername
}

// GetCurrentUserID returns the user ID of the currently logged-in user
func GetCurrentUserID() string {
	creds, err := authentication.GetTokens()
	if err != nil {
		return ""
	}
	return creds.UserID
}
