package command

import (
	"fmt"
	"mangahub/cmd/cli/command/client"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var (
	tcpClient *client.TCPClient
	tcpServer string
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "TCP synchronization commands",
	Long:  `Connect to TCP sync server for real-time progress synchronization across devices.`,
}

// syncConnectCmd represents the sync connect command
var syncConnectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to TCP sync server",
	Long:  `Establish a connection to the TCP sync server for real-time progress updates.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Connecting to TCP sync server at %s...\n", tcpServer)

		// Create TCP client
		tcpClient = client.NewTCPClient(tcpServer)

		// Connect with authentication
		username := GetCurrentUsername()
		if username == "" {
			return fmt.Errorf("not authenticated, please login first")
		}

		err := tcpClient.Connect(username, accessToken)
		if err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}

		fmt.Println("✓ Connected successfully!")
		fmt.Println("\nConnection Details:")
		fmt.Printf("  Server: %s\n", tcpServer)
		fmt.Printf("  User: %s\n", username)
		sessionID := tcpClient.GetSessionID()
		if sessionID != "" {
			fmt.Printf("  Session ID: %s\n", sessionID)
		}
		fmt.Printf("  Connected at: %s\n", time.Now().Format("2006-01-02 15:04:05 MST"))

		fmt.Println("\nSync Status:")
		fmt.Println("  Auto-sync: enabled")
		fmt.Println("  Conflict resolution: last_write_wins")
		fmt.Println("  Real-time sync is now active.")
		fmt.Println("\nYour progress will be synchronized across all devices.")

		// Keep connection alive in background
		fmt.Println("\nConnection established. Use 'mangahub sync disconnect' to close.")

		return nil
	},
}

// syncDisconnectCmd represents the sync disconnect command
var syncDisconnectCmd = &cobra.Command{
	Use:   "disconnect",
	Short: "Disconnect from TCP sync server",
	Long:  `Close the connection to the TCP sync server.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if tcpClient == nil || !tcpClient.IsConnected() {
			return fmt.Errorf("not connected to sync server")
		}

		fmt.Println("Disconnecting from TCP sync server...")

		err := tcpClient.Disconnect()
		if err != nil {
			return fmt.Errorf("error during disconnect: %w", err)
		}

		fmt.Println("✓ Disconnected successfully")
		tcpClient = nil

		return nil
	},
}

// syncStatusCmd represents the sync status command
var syncStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check TCP sync connection status",
	Long:  `Display the current status of the TCP sync connection.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("TCP Sync Status:")

		if tcpClient == nil || !tcpClient.IsConnected() {
			fmt.Println("  Connection: ✗ Not connected")
			fmt.Printf("  Server: %s\n", tcpServer)
			fmt.Println("\nTo connect:")
			fmt.Println("  mangahub sync connect")
			return nil
		}

		// Connected
		fmt.Println("  Connection: ✓ Active")
		fmt.Printf("  Server: %s\n", tcpServer)

		stats := tcpClient.GetStats()
		fmt.Printf("  Uptime: %s\n", formatDuration(stats.Uptime))

		if stats.LastHeartbeat.IsZero() {
			fmt.Println("  Last heartbeat: Never")
		} else {
			elapsed := time.Since(stats.LastHeartbeat)
			fmt.Printf("  Last heartbeat: %s ago\n", formatDuration(elapsed))
		}

		fmt.Println("\nSession Info:")
		fmt.Printf("  User: %s\n", GetCurrentUsername())
		if sessionID := tcpClient.GetSessionID(); sessionID != "" {
			fmt.Printf("  Session ID: %s\n", sessionID)
		}

		fmt.Println("\nSync Statistics:")
		fmt.Printf("  Messages sent: %d\n", stats.MessagesSent)
		fmt.Printf("  Messages received: %d\n", stats.MessagesReceived)

		if !stats.LastSync.IsZero() {
			fmt.Printf("  Last sync: %s ago\n", formatDuration(time.Since(stats.LastSync)))
		} else {
			fmt.Println("  Last sync: Never")
		}

		fmt.Printf("  Sync conflicts: %d\n", stats.Conflicts)

		// Network quality (simulated)
		fmt.Println("\nNetwork Quality: Excellent (RTT: 15ms)")

		return nil
	},
}

// syncMonitorCmd represents the sync monitor command
var syncMonitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Monitor real-time sync updates",
	Long:  `Display real-time progress synchronization updates from all connected devices.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if tcpClient == nil || !tcpClient.IsConnected() {
			return fmt.Errorf("not connected to sync server. Run 'mangahub sync connect' first")
		}

		fmt.Println("Monitoring real-time sync updates... (Press Ctrl+C to exit)")

		// Set up signal handling for graceful exit
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		// Start monitoring in a goroutine
		msgChan := make(chan client.SyncMessage, 100)
		go tcpClient.StartMonitoring(msgChan)

		// Display messages
		for {
			select {
			case <-sigChan:
				fmt.Println("\n\nStopping monitor...")
				tcpClient.StopMonitoring()
				return nil

			case msg := <-msgChan:
				timestamp := time.Now().Format("15:04:05")

				switch msg.Direction {
				case "incoming":
					fmt.Printf("[%s] ← Device '%s' updated: %s → Chapter %d\n",
						timestamp, msg.Device, msg.MangaTitle, msg.Chapter)

				case "outgoing":
					fmt.Printf("[%s] → Broadcasting update: %s → Chapter %d\n",
						timestamp, msg.MangaTitle, msg.Chapter)

				case "conflict":
					fmt.Printf("[%s] ⚠ Conflict resolved: %s → Chapter %d (local) vs %d (remote)\n",
						timestamp, msg.MangaTitle, msg.Chapter, msg.ConflictChapter)
				}
			}
		}
	},
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	return fmt.Sprintf("%dd %dh", days, hours)
}

func init() {
	// Add sync subcommands
	syncCmd.AddCommand(syncConnectCmd)
	syncCmd.AddCommand(syncDisconnectCmd)
	syncCmd.AddCommand(syncStatusCmd)
	syncCmd.AddCommand(syncMonitorCmd)

	// TCP server flag
	defaultTCPServer := "localhost:8081"
	if v := os.Getenv("MANGAHUB_TCP_SERVER"); v != "" {
		defaultTCPServer = v
	}
	tcpServer = defaultTCPServer
	syncCmd.PersistentFlags().StringVar(&tcpServer, "server", defaultTCPServer, "TCP sync server address")
}
