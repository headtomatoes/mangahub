package command

// root.go defines the root command for the mangahubCLI application.
// set up the global flags and configuration here.

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	apiURL  string //Global flag for API server URL
	cfgFile string // config file path
	token   string // authentication token(jwt)
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
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
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
	// Load configuration from file or environment variables || add later
	// for now use hardcoded values
	// cfg, err := loadConfig()
	// if err != nil {
	// 	fmt.Fprintln(os.Stderr, "Error loading config:", err)
	// 	os.Exit(1)
	// }

	// Global persistent flags = available to all subcommands
	rootCmd.PersistentFlags().StringVar(&apiURL, "api", "http://localhost:8084", "API server URL")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "$HOME/.mangahub/config.json", "config file path")

	// loadConfig() for now is load token from config file
	loadConfig()
}

func loadConfig() {
	// Placeholder function to load jwt token from config file
	// implement actual file reading and parsing later
}
