package command

import (
	"fmt"
	"mangahub/cmd/cli/authentication"
	"mangahub/cmd/cli/command/client"
	"mangahub/cmd/cli/command/state"
	"mangahub/cmd/cli/dto"
	"os"
	"os/exec"
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
		// Check if already connected
		connState, _ := state.LoadConnectionState()
		if connState != nil && connState.IsProcessRunning() {
			fmt.Println("Already connected!")
			fmt.Printf("  Server: %s\n", connState.Server)
			fmt.Printf("  Session ID: %s\n", connState.SessionID)
			fmt.Printf("  Connected at: %s\n", connState.ConnectedAt.Format("2006-01-02 15:04:05"))
			return nil
		}

		fmt.Printf("Connecting to TCP sync server at %s...\n", tcpServer)

		// Start daemon in background
		daemonCmd := exec.Command(os.Args[0], "sync", "daemon", "--server", tcpServer)
		daemonCmd.Stdout = nil
		daemonCmd.Stderr = nil

		if err := daemonCmd.Start(); err != nil {
			return fmt.Errorf("failed to start daemon: %w", err)
		}

		// Wait for daemon to establish connection
		time.Sleep(3 * time.Second)

		// Verify connection
		connState, err := state.LoadConnectionState()
		if err != nil || connState == nil || !connState.Connected {
			return fmt.Errorf("daemon failed to establish connection")
		}

		fmt.Println("✓ Connected successfully!")
		fmt.Println("\nConnection Details:")
		fmt.Printf("  Server: %s\n", connState.Server)
		fmt.Printf("  User: %s\n", connState.Username)
		fmt.Printf("  Session ID: %s\n", connState.SessionID)
		fmt.Printf("  Daemon PID: %d\n", connState.PID)

		fmt.Println("\n✓ Background daemon is running")
		fmt.Println("Connection will remain active until 'mangahub sync disconnect'")

		return nil
	},
}

// syncDisconnectCmd represents the sync disconnect command
var syncDisconnectCmd = &cobra.Command{
	Use:   "disconnect",
	Short: "Disconnect from TCP sync server",
	Long:  `Close the connection to the TCP sync server.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		connState, err := state.LoadConnectionState()
		if err != nil {
			return fmt.Errorf("failed to load state: %w", err)
		}

		if connState == nil || !connState.Connected {
			return fmt.Errorf("not connected to sync server")
		}

		fmt.Println("Disconnecting from TCP sync server...")

		// Kill daemon process
		if connState.PID != 0 {
			process, err := os.FindProcess(connState.PID)
			if err == nil {
				// On Windows, use Kill() instead of Signal(Interrupt)
				// because Windows doesn't support Unix signals
				process.Kill()
				time.Sleep(500 * time.Millisecond)
			}
		}

		// Clear state
		if err := state.ClearConnectionState(); err != nil {
			return fmt.Errorf("failed to clear state: %w", err)
		}

		fmt.Println("✓ Disconnected successfully")
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

		connState, err := state.LoadConnectionState()
		if err != nil {
			return fmt.Errorf("failed to load state: %w", err)
		}

		if connState == nil || !connState.Connected {
			fmt.Println("  Connection: ✗ Not connected")
			fmt.Printf("  Server: %s\n", tcpServer)
			fmt.Println("\nTo connect:")
			fmt.Println("  mangahub sync connect")
			return nil
		}

		// Check if daemon is still running
		if !connState.IsProcessRunning() {
			fmt.Println("  Connection: ✗ Daemon not running")
			fmt.Println("\nConnection was lost. Clearing stale state...")
			state.ClearConnectionState()
			fmt.Println("\nTo reconnect:")
			fmt.Println("  mangahub sync connect")
			return nil
		}

		// Connected
		fmt.Println("  Connection: ✓ Active")
		fmt.Printf("  Server: %s\n", connState.Server)
		fmt.Printf("  User: %s\n", connState.Username)
		fmt.Printf("  Session ID: %s\n", connState.SessionID)
		fmt.Printf("  Daemon PID: %d\n", connState.PID)

		uptime := time.Since(connState.ConnectedAt)
		fmt.Printf("  Uptime: %s\n", formatDuration(uptime))

		return nil
	},
}

// syncMonitorCmd represents the sync monitor command
var syncMonitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Monitor real-time sync updates",
	Long:  `Display real-time progress synchronization updates from all connected devices.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if connected
		connState, err := state.LoadConnectionState()
		if err != nil {
			return err
		}

		if connState == nil || !connState.Connected || !connState.IsProcessRunning() {
			return fmt.Errorf("not connected to sync server. Run 'mangahub sync connect' first")
		}

		fmt.Println("Monitoring real-time sync updates... (Press Ctrl+C to exit)")

		// Create new temporary connection for monitoring
		monitorClient := client.NewTCPClient(connState.Server)
		username := GetCurrentUsername()

		if err := monitorClient.Connect(username, accessToken); err != nil {
			return fmt.Errorf("failed to connect for monitoring: %w", err)
		}
		defer monitorClient.Disconnect()

		// Set up signal handling for graceful exit
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		// Start monitoring in a goroutine
		msgChan := make(chan client.SyncMessage, 100)
		go monitorClient.StartMonitoring(msgChan)

		// Display messages
		for {
			select {
			case <-sigChan:
				fmt.Println("\n\nStopping monitor...")
				monitorClient.StopMonitoring()
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

// syncDaemonCmd runs TCP sync daemon in background (hidden command)
var syncDaemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Run TCP sync daemon in background",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDaemon()
	},
}

// runDaemon keeps TCP connection alive in background
func runDaemon() error {
	// Load credentials from storage (since daemon runs as separate process)
	creds, err := authentication.GetTokens()
	if err != nil {
		return fmt.Errorf("not authenticated: %w", err)
	}

	// Check if token needs refresh
	now := time.Now().Unix()
	accessTokenToUse := creds.AccessToken

	if creds.ExpiresAt-now < 300 { // Refresh if expires in less than 5 minutes
		// Token expired or about to expire, refresh it
		httpClient := client.NewHTTPClient(apiURL)
		req := &dto.RefreshTokenRequest{RefreshToken: creds.RefreshToken}

		refreshResp, err := httpClient.RefreshToken(req)
		if err != nil {
			return fmt.Errorf("token refresh failed, please login again: %w", err)
		}

		// Store refreshed tokens
		err = authentication.StoreTokens(&authentication.StoredCredentials{
			AccessToken:  refreshResp.AccessToken,
			RefreshToken: refreshResp.RefreshToken,
			Username:     creds.Username,
			UserID:       creds.UserID,
			ExpiresAt:    now + refreshResp.ExpiresIn,
		})
		if err != nil {
			return fmt.Errorf("failed to store refreshed tokens: %w", err)
		}

		accessTokenToUse = refreshResp.AccessToken
	}

	// Create TCP client
	tcpClient := client.NewTCPClient(tcpServer)

	// Connect using fresh credentials
	if err := tcpClient.Connect(creds.Username, accessTokenToUse); err != nil {
		return fmt.Errorf("TCP connection failed: %w", err)
	}

	// Save state
	connState := &state.TCPConnectionState{
		Connected:   true,
		Server:      tcpServer,
		Username:    creds.Username,
		SessionID:   tcpClient.GetSessionID(),
		ConnectedAt: time.Now(),
		PID:         os.Getpid(),
	}
	if err := state.SaveConnectionState(connState); err != nil {
		return err
	}

	// Handle signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Keep connection alive
	<-sigChan

	// Cleanup
	tcpClient.Disconnect()
	state.ClearConnectionState()
	return nil
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
	syncCmd.AddCommand(syncDaemonCmd)

	// TCP server flag
	defaultTCPServer := AddressServer + ":" + TCP_PORT
	if v := os.Getenv("MANGAHUB_TCP_SERVER"); v != "" {
		defaultTCPServer = v
	}
	tcpServer = defaultTCPServer
	syncCmd.PersistentFlags().StringVar(&tcpServer, "server", defaultTCPServer, "TCP sync server address")
}
