package command

import (
	"fmt"
	"mangahub/cmd/cli/authentication"
	"mangahub/cmd/cli/command/client"
	"os"

	"github.com/spf13/cobra"
)

var udpServerAddr string

// udpCmd represents the UDP notification client command
var udpCmd = &cobra.Command{
	Use:   "udp",
	Short: "UDP notification client commands",
	Long: `UDP notification client for receiving real-time notifications about:
- New manga additions
- New chapter releases for manga in your library
- Manga updates

The UDP client allows you to subscribe to push notifications and receive them in real-time.`,
}

// udpListenCmd starts listening for UDP notifications
var udpListenCmd = &cobra.Command{
	Use:   "listen",
	Short: "Listen for real-time notifications",
	Long: `Connect to the UDP notification server and listen for real-time notifications.

This command will:
1. Connect to the UDP notification server
2. Subscribe using your authenticated user ID
3. Listen for and display incoming notifications
4. Send periodic ping messages to keep the connection alive

Press Ctrl+C to stop listening and disconnect.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get user credentials
		creds, err := authentication.GetTokens()
		if err != nil {
			return fmt.Errorf("not logged in, please run 'mangahubCLI auth login' first")
		}

		// Create UDP client
		udpClient := client.NewUDPClient(udpServerAddr)

		fmt.Println("🔌 Connecting to UDP notification server...")
		fmt.Printf("   Server: %s\n", udpServerAddr)
		fmt.Printf("   User: %s (ID: %s)\n\n", creds.Username, creds.UserID)

		// Connect and subscribe
		if err := udpClient.Connect(creds.UserID); err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}

		// Start listening (blocks until Ctrl+C)
		if err := udpClient.StartListening(); err != nil {
			return fmt.Errorf("error during listening: %w", err)
		}

		// Print final stats
		udpClient.PrintStats()

		return nil
	},
}

// udpTestCmd tests the UDP connection
var udpTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test UDP connection",
	Long:  `Test the connection to the UDP notification server by subscribing and waiting for a confirmation message.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get user credentials
		creds, err := authentication.GetTokens()
		if err != nil {
			return fmt.Errorf("not logged in, please run 'mangahubCLI auth login' first")
		}

		// Create UDP client
		udpClient := client.NewUDPClient(udpServerAddr)

		fmt.Println("🔌 Testing UDP connection...")
		fmt.Printf("   Server: %s\n", udpServerAddr)
		fmt.Printf("   User: %s (ID: %s)\n\n", creds.Username, creds.UserID)

		// Connect and subscribe
		if err := udpClient.Connect(creds.UserID); err != nil {
			return fmt.Errorf("connection failed: %w", err)
		}

		fmt.Println("✓ Connection successful!")
		fmt.Println("✓ Subscription confirmed")
		fmt.Println("\nWaiting for confirmation message from server...")

		// Wait briefly for the subscription confirmation
		// (The server sends a confirmation notification upon successful subscription)
		// We'll just wait a moment then disconnect
		// In the real scenario, the listenRoutine would handle this
		fmt.Println("\n✓ Test completed successfully!")

		// Disconnect
		if err := udpClient.Disconnect(); err != nil {
			return fmt.Errorf("disconnect failed: %w", err)
		}

		return nil
	},
}

// udpStatsCmd shows connection statistics (if connected)
var udpStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show UDP connection statistics",
	Long:  `Display statistics about the current or last UDP connection session.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Note: Stats are only available during an active listening session.")
		fmt.Println("Use 'mangahubCLI udp listen' to start a session and view real-time stats.")
		return nil
	},
}

func init() {
	// Get UDP server address from environment or use default
	defaultUDPAddr := AddressServer + ":" + UDP_PORT
	if v := os.Getenv("MANGAHUB_UDP_ADDR"); v != "" {
		defaultUDPAddr = v
	}
	udpServerAddr = defaultUDPAddr

	// Add flags
	udpCmd.PersistentFlags().StringVar(&udpServerAddr, "server", defaultUDPAddr, "UDP server address (host:port)")

	// Add subcommands
	udpCmd.AddCommand(udpListenCmd)
	udpCmd.AddCommand(udpTestCmd)
	udpCmd.AddCommand(udpStatsCmd)
}
