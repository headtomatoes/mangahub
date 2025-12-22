package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
)

type TCPConnectionState struct {
	Connected   bool      `json:"connected"`
	Server      string    `json:"server"`
	Username    string    `json:"username"`
	SessionID   string    `json:"session_id"`
	ConnectedAt time.Time `json:"connected_at"`
	PID         int       `json:"pid"` // Process ID of connection holder
}

func GetStateFilePath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".mangahub", "tcp_state.json")
}

func SaveConnectionState(state *TCPConnectionState) error {
	stateDir := filepath.Dir(GetStateFilePath())
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(GetStateFilePath(), data, 0600)
}

func LoadConnectionState() (*TCPConnectionState, error) {
	data, err := os.ReadFile(GetStateFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No state file = not connected
		}
		return nil, err
	}

	var state TCPConnectionState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

func ClearConnectionState() error {
	stateFile := GetStateFilePath()
	err := os.Remove(stateFile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// IsProcessRunning checks if the process holding the connection is still alive
func (s *TCPConnectionState) IsProcessRunning() bool {
	if s.PID == 0 {
		return false
	}

	// Check if process exists (Windows & Unix compatible)
	process, err := os.FindProcess(s.PID)
	if err != nil {
		return false
	}

	// Platform-specific process check
	if runtime.GOOS == "windows" {
		// On Windows, try to open the process handle
		// If it fails, the process doesn't exist
		// Signal(0) doesn't work on Windows, so we use a different approach
		err = process.Signal(syscall.Signal(0))
		// On Windows, this will always succeed if PID is valid
		// So we need to trust the state was updated recently
		// A better approach: check if state file is recent (< 60 seconds old)
		if time.Since(s.ConnectedAt) > 5*time.Minute {
			return false // Stale connection
		}
		return true
	}

	// On Unix, sending signal 0 checks if process exists
	err = process.Signal(syscall.Signal(0))
	return err == nil
}
