# MangaHub CLI Implementation Plan

## 🎯 Overview

Build a comprehensive CLI client using Cobra that interacts with the MangaHub microservices backend (HTTP-API and TCP servers).

## 📋 Project Structure

```
mangahub/
├── cmd/
│   ├── mangahub-cli/          # CLI entry point
│   │   └── main.go
│   ├── api-server/            # Existing HTTP-API server
│   ├── tcp-server/            # Existing TCP server
│   └── ...
├── internal/
│   ├── cli/                   # NEW: CLI implementation
│   │   ├── cmd/              # Cobra commands
│   │   │   ├── root.go       # Root command
│   │   │   ├── server.go     # Server management
│   │   │   ├── auth.go       # Authentication
│   │   │   ├── manga.go      # Manga operations
│   │   │   ├── library.go    # Library management
│   │   │   └── progress.go   # Progress tracking
│   │   ├── client/           # API clients
│   │   │   ├── api_client.go # HTTP REST client
│   │   │   ├── tcp_client.go # TCP client
│   │   │   └── config.go     # Client configuration
│   │   └── utils/            # CLI utilities
│   │       ├── output.go     # Formatting/display
│   │       ├── spinner.go    # Progress indicators
│   │       └── errors.go     # Error handling
│   ├── microservices/         # Existing
│   └── ...
```

## 🔧 Implementation Tasks

### 1. Setup & Dependencies

**Install Cobra:**
```bash
go get -u github.com/spf13/cobra@latest
go get -u github.com/spf13/viper@latest  # For config management
go get -u github.com/fatih/color@latest  # For colored output
go get -u github.com/briandowns/spinner@latest  # For spinners
```

### 2. CLI Command Structure

```
mangahub                              # Root command
├── server                            # Server management
│   ├── start                        # Start all services
│   ├── stop                         # Stop all services
│   └── status                       # Check service status
├── auth                              # Authentication
│   ├── register                     # Create account
│   ├── login                        # Login and get token
│   └── logout                       # Clear local token
├── manga                             # Manga operations
│   ├── search <query>               # Search manga
│   ├── list                         # List all manga
│   ├── get <manga-id>               # Get manga details
│   └── add                          # Add new manga (admin)
├── library                           # Personal library
│   ├── add <manga-id>               # Add to library
│   ├── remove <manga-id>            # Remove from library
│   ├── list                         # List library manga
│   └── status <manga-id> <status>   # Update status
└── progress                          # Reading progress
    ├── update <manga-id> <chapter>  # Update progress
    ├── get <manga-id>               # Get progress
    ├── sync                         # Sync via TCP
    └── list                         # List all progress
```

### 3. Configuration Management

**Config File Location:** `~/.mangahub/config.yaml`

```yaml
# MangaHub CLI Configuration
server:
  http_url: "http://localhost:8084"
  tcp_url: "localhost:8081"
  timeout: 30s

auth:
  token: ""  # JWT token stored after login
  username: ""

preferences:
  output_format: "table"  # table, json, yaml
  color: true
  auto_sync: true  # Auto-sync progress via TCP
```

### 4. API Client Implementation

#### HTTP REST Client
```go
// internal/cli/client/api_client.go
type APIClient struct {
    BaseURL    string
    HTTPClient *http.Client
    Token      string
}

// Methods:
- Register(username, email, password) error
- Login(username, password) (token string, err error)
- SearchManga(query string) ([]Manga, error)
- GetManga(id string) (*Manga, error)
- UpdateProgress(mangaID string, chapter int64) error
- GetProgress(mangaID string) (*Progress, error)
```

#### TCP Client
```go
// internal/cli/client/tcp_client.go
type TCPClient struct {
    Address string
    conn    net.Conn
    Token   string
}

// Methods:
- Connect() error
- Authenticate(token string) error
- SendProgress(mangaID string, chapter int64) error
- ReceiveProgress() (*Progress, error)
- Close() error
```

## 📝 Command Examples & Implementation

### Example 1: Server Management
```bash
# Start all MangaHub services
mangahub server start
# Output:
# ✓ Starting HTTP-API server on port 8084...
# ✓ Starting TCP server on port 8081...
# ✓ All services running

# Check status
mangahub server status
# Output:
# HTTP-API: ✓ Running (http://localhost:8084)
# TCP:      ✓ Running (localhost:8081)
# Database: ✓ Connected
```

**Implementation:**
- Use `exec.Command()` to start `api-server` and `tcp-server` as background processes
- Store PIDs in `~/.mangahub/services.json`
- Check health endpoints: `/check-conn` for HTTP, TCP connection test

### Example 2: Authentication
```bash
# Register new account
mangahub auth register --username myuser --email myuser@example.com
# Prompt: Enter password: ********
# Output: ✓ Account created successfully!

# Login
mangahub auth login --username myuser
# Prompt: Enter password: ********
# Output: ✓ Logged in successfully!
#         Token saved to ~/.mangahub/config.yaml
```

**Implementation:**
```go
// POST /auth/register
{
  "username": "myuser",
  "email": "myuser@example.com",
  "password": "hashed_password"
}

// POST /auth/login
{
  "username": "myuser",
  "password": "password"
}
// Response: {"token": "eyJhbGc..."}
// Store token in config file
```

### Example 3: Manga Search & Library
```bash
# Search for manga
mangahub manga search "one piece"
# Output (table format):
# ID            | Title                  | Status    | Chapters
# one-piece     | One Piece             | Ongoing   | 1095
# one-punch-man | One Punch Man         | Ongoing   | 180

# Add to library
mangahub library add --manga-id one-piece --status reading
# Output: ✓ Added "One Piece" to your library

# List library
mangahub library list
# Output:
# Title          | Status    | Progress  | Last Updated
# One Piece      | Reading   | Ch. 0     | 2025-11-06
```

**Implementation:**
```go
// GET /api/manga/search?q=one+piece
// Headers: Authorization: Bearer <token>

// POST /api/library (hypothetical - needs implementation)
{
  "manga_id": "one-piece",
  "status": "reading"
}
```

### Example 4: Progress Tracking
```bash
# Update progress (via HTTP-API)
mangahub progress update --manga-id one-piece --chapter 1095
# Output: ✓ Updated progress: One Piece - Chapter 1095

# Update progress with real-time sync (via TCP)
mangahub progress update --manga-id one-piece --chapter 1095 --sync
# Output: ⟳ Connecting to TCP server...
#         ✓ Progress synced in real-time!

# Get current progress
mangahub progress get --manga-id one-piece
# Output:
# Manga: One Piece
# Current Chapter: 1095
# Last Updated: 2025-11-06 14:30:00
```

**Implementation:**
```go
// HTTP method:
// PUT /api/progress
// Headers: Authorization: Bearer <token>
{
  "manga_id": 1,
  "chapter": 1095,
  "last_page": 0
}

// TCP method:
// 1. Connect to TCP server
// 2. Send authentication message
// 3. Send progress message
{
  "type": "progress",
  "manga_id": 1,
  "chapter": 1095,
  "timestamp": "2025-11-06T14:30:00Z"
}
```

## 🛠 Implementation Steps

### Step 1: Basic CLI Setup (Task 1-2)
```bash
# Create directories
mkdir -p internal/cli/cmd internal/cli/client internal/cli/utils

# Install dependencies
go get github.com/spf13/cobra@latest
go get github.com/spf13/viper@latest
go get github.com/fatih/color@latest
go get github.com/briandowns/spinner@latest
```

### Step 2: Root Command & Config
Create `cmd/mangahub-cli/main.go`:
```go
package main

import (
    "mangahub/internal/cli/cmd"
    "os"
)

func main() {
    if err := cmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

Create `internal/cli/cmd/root.go`:
```go
package cmd

import (
    "fmt"
    "os"
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
    Use:   "mangahub",
    Short: "MangaHub CLI - Manage your manga library",
    Long:  `A comprehensive CLI tool to interact with MangaHub microservices`,
}

func Execute() error {
    return rootCmd.Execute()
}

func init() {
    cobra.OnInitialize(initConfig)
    // Add subcommands
    rootCmd.AddCommand(serverCmd)
    rootCmd.AddCommand(authCmd)
    rootCmd.AddCommand(mangaCmd)
    rootCmd.AddCommand(libraryCmd)
    rootCmd.AddCommand(progressCmd)
}

func initConfig() {
    home, err := os.UserHomeDir()
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    
    viper.AddConfigPath(home + "/.mangahub")
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AutomaticEnv()
    
    if err := viper.ReadInConfig(); err == nil {
        fmt.Println("Using config file:", viper.ConfigFileUsed())
    }
}
```

### Step 3: API Client
Create `internal/cli/client/api_client.go`:
```go
package client

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

type APIClient struct {
    BaseURL    string
    HTTPClient *http.Client
    Token      string
}

func NewAPIClient(baseURL, token string) *APIClient {
    return &APIClient{
        BaseURL: baseURL,
        HTTPClient: &http.Client{
            Timeout: 30 * time.Second,
        },
        Token: token,
    }
}

func (c *APIClient) doRequest(method, endpoint string, body interface{}) ([]byte, error) {
    var reqBody io.Reader
    if body != nil {
        jsonData, err := json.Marshal(body)
        if err != nil {
            return nil, err
        }
        reqBody = bytes.NewBuffer(jsonData)
    }
    
    req, err := http.NewRequest(method, c.BaseURL+endpoint, reqBody)
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Content-Type", "application/json")
    if c.Token != "" {
        req.Header.Set("Authorization", "Bearer "+c.Token)
    }
    
    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    
    if resp.StatusCode >= 400 {
        return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(respBody))
    }
    
    return respBody, nil
}

// Authentication methods
func (c *APIClient) Register(username, email, password string) error {
    payload := map[string]string{
        "username": username,
        "email":    email,
        "password": password,
    }
    _, err := c.doRequest("POST", "/auth/register", payload)
    return err
}

func (c *APIClient) Login(username, password string) (string, error) {
    payload := map[string]string{
        "username": username,
        "password": password,
    }
    
    respBody, err := c.doRequest("POST", "/auth/login", payload)
    if err != nil {
        return "", err
    }
    
    var result struct {
        Token string `json:"token"`
    }
    if err := json.Unmarshal(respBody, &result); err != nil {
        return "", err
    }
    
    return result.Token, nil
}

// Manga methods
func (c *APIClient) SearchManga(query string) ([]map[string]interface{}, error) {
    respBody, err := c.doRequest("GET", "/api/manga/search?q="+query, nil)
    if err != nil {
        return nil, err
    }
    
    var result struct {
        Data []map[string]interface{} `json:"data"`
    }
    if err := json.Unmarshal(respBody, &result); err != nil {
        return nil, err
    }
    
    return result.Data, nil
}

// Progress methods
func (c *APIClient) UpdateProgress(mangaID int64, chapter int64) error {
    payload := map[string]interface{}{
        "manga_id": mangaID,
        "chapter":  chapter,
    }
    _, err := c.doRequest("PUT", "/api/progress", payload)
    return err
}
```

### Step 4: Implement Commands

**Server Command** (`internal/cli/cmd/server.go`):
```go
package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
    "os/exec"
)

var serverCmd = &cobra.Command{
    Use:   "server",
    Short: "Manage MangaHub server",
}

var serverStartCmd = &cobra.Command{
    Use:   "start",
    Short: "Start MangaHub services",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("🚀 Starting MangaHub services...")
        
        // Start HTTP-API server
        apiCmd := exec.Command("mangahub/cmd/api-server/api-server")
        if err := apiCmd.Start(); err != nil {
            fmt.Printf("❌ Failed to start HTTP-API: %v\n", err)
            return
        }
        fmt.Println("✓ HTTP-API server started on port 8084")
        
        // Start TCP server
        tcpCmd := exec.Command("mangahub/cmd/tcp-server/tcp-server")
        if err := tcpCmd.Start(); err != nil {
            fmt.Printf("❌ Failed to start TCP server: %v\n", err)
            return
        }
        fmt.Println("✓ TCP server started on port 8081")
        
        fmt.Println("\n✓ All services running!")
    },
}

func init() {
    serverCmd.AddCommand(serverStartCmd)
}
```

**Auth Command** (`internal/cli/cmd/auth.go`):
```go
package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
    "golang.org/x/term"
    "mangahub/internal/cli/client"
    "os"
    "syscall"
)

var authCmd = &cobra.Command{
    Use:   "auth",
    Short: "Authentication commands",
}

var registerCmd = &cobra.Command{
    Use:   "register",
    Short: "Register a new account",
    Run: func(cmd *cobra.Command, args []string) {
        username, _ := cmd.Flags().GetString("username")
        email, _ := cmd.Flags().GetString("email")
        
        fmt.Print("Enter password: ")
        passwordBytes, _ := term.ReadPassword(int(syscall.Stdin))
        password := string(passwordBytes)
        fmt.Println()
        
        apiClient := client.NewAPIClient(viper.GetString("server.http_url"), "")
        if err := apiClient.Register(username, email, password); err != nil {
            fmt.Printf("❌ Registration failed: %v\n", err)
            return
        }
        
        fmt.Println("✓ Account created successfully!")
    },
}

var loginCmd = &cobra.Command{
    Use:   "login",
    Short: "Login to your account",
    Run: func(cmd *cobra.Command, args []string) {
        username, _ := cmd.Flags().GetString("username")
        
        fmt.Print("Enter password: ")
        passwordBytes, _ := term.ReadPassword(int(syscall.Stdin))
        password := string(passwordBytes)
        fmt.Println()
        
        apiClient := client.NewAPIClient(viper.GetString("server.http_url"), "")
        token, err := apiClient.Login(username, password)
        if err != nil {
            fmt.Printf("❌ Login failed: %v\n", err)
            return
        }
        
        // Save token to config
        viper.Set("auth.token", token)
        viper.Set("auth.username", username)
        viper.WriteConfig()
        
        fmt.Println("✓ Logged in successfully!")
    },
}

func init() {
    registerCmd.Flags().String("username", "", "Username")
    registerCmd.Flags().String("email", "", "Email address")
    registerCmd.MarkFlagRequired("username")
    registerCmd.MarkFlagRequired("email")
    
    loginCmd.Flags().String("username", "", "Username")
    loginCmd.MarkFlagRequired("username")
    
    authCmd.AddCommand(registerCmd)
    authCmd.AddCommand(loginCmd)
}
```

**Progress Command** (`internal/cli/cmd/progress.go`):
```go
package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
    "mangahub/internal/cli/client"
    "strconv"
)

var progressCmd = &cobra.Command{
    Use:   "progress",
    Short: "Manage reading progress",
}

var updateProgressCmd = &cobra.Command{
    Use:   "update",
    Short: "Update reading progress",
    Run: func(cmd *cobra.Command, args []string) {
        mangaID, _ := cmd.Flags().GetString("manga-id")
        chapter, _ := cmd.Flags().GetInt64("chapter")
        sync, _ := cmd.Flags().GetBool("sync")
        
        token := viper.GetString("auth.token")
        if token == "" {
            fmt.Println("❌ Please login first: mangahub auth login")
            return
        }
        
        if sync {
            // Use TCP client for real-time sync
            tcpClient := client.NewTCPClient(viper.GetString("server.tcp_url"))
            if err := tcpClient.Connect(); err != nil {
                fmt.Printf("❌ TCP connection failed: %v\n", err)
                return
            }
            defer tcpClient.Close()
            
            if err := tcpClient.Authenticate(token); err != nil {
                fmt.Printf("❌ Authentication failed: %v\n", err)
                return
            }
            
            mangaIDInt, _ := strconv.ParseInt(mangaID, 10, 64)
            if err := tcpClient.SendProgress(mangaIDInt, chapter); err != nil {
                fmt.Printf("❌ Progress update failed: %v\n", err)
                return
            }
            
            fmt.Println("✓ Progress synced in real-time via TCP!")
        } else {
            // Use HTTP-API
            apiClient := client.NewAPIClient(viper.GetString("server.http_url"), token)
            mangaIDInt, _ := strconv.ParseInt(mangaID, 10, 64)
            if err := apiClient.UpdateProgress(mangaIDInt, chapter); err != nil {
                fmt.Printf("❌ Progress update failed: %v\n", err)
                return
            }
            
            fmt.Printf("✓ Updated progress: %s - Chapter %d\n", mangaID, chapter)
        }
    },
}

func init() {
    updateProgressCmd.Flags().String("manga-id", "", "Manga ID")
    updateProgressCmd.Flags().Int64("chapter", 0, "Chapter number")
    updateProgressCmd.Flags().Bool("sync", false, "Use TCP for real-time sync")
    updateProgressCmd.MarkFlagRequired("manga-id")
    updateProgressCmd.MarkFlagRequired("chapter")
    
    progressCmd.AddCommand(updateProgressCmd)
}
```

## 🎨 Output Formatting

### Table Output
```go
// internal/cli/utils/output.go
package utils

import (
    "fmt"
    "os"
    "text/tabwriter"
)

func PrintTable(headers []string, rows [][]string) {
    w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
    
    // Print headers
    for i, h := range headers {
        fmt.Fprint(w, h)
        if i < len(headers)-1 {
            fmt.Fprint(w, "\t")
        }
    }
    fmt.Fprintln(w)
    
    // Print separator
    for i := range headers {
        fmt.Fprint(w, "---")
        if i < len(headers)-1 {
            fmt.Fprint(w, "\t")
        }
    }
    fmt.Fprintln(w)
    
    // Print rows
    for _, row := range rows {
        for i, cell := range row {
            fmt.Fprint(w, cell)
            if i < len(row)-1 {
                fmt.Fprint(w, "\t")
            }
        }
        fmt.Fprintln(w)
    }
    
    w.Flush()
}
```

## 🧪 Testing Commands

```bash
# Build CLI
go build -o mangahub cmd/mangahub-cli/main.go

# Initialize config
./mangahub config init

# Start services
./mangahub server start

# Register and login
./mangahub auth register --username testuser --email test@example.com
./mangahub auth login --username testuser

# Search manga
./mangahub manga search "naruto"

# Add to library
./mangahub library add --manga-id naruto --status reading

# Update progress
./mangahub progress update --manga-id naruto --chapter 700

# Update progress with TCP sync
./mangahub progress update --manga-id naruto --chapter 701 --sync
```

## 📦 Build & Distribution

### Makefile targets
```makefile
# Add to existing Makefile
.PHONY: cli
cli:
	go build -o bin/mangahub cmd/mangahub-cli/main.go

.PHONY: install-cli
install-cli:
	go install cmd/mangahub-cli/main.go

.PHONY: cli-test
cli-test:
	go test ./internal/cli/...
```

## 🚀 Next Steps

1. **Task 1-2**: Create basic CLI structure with Cobra and API client
2. **Task 3**: Implement server management commands
3. **Task 4**: Implement auth commands
4. **Task 5-7**: Implement manga, library, and progress commands
5. **Task 8**: Add utilities, testing, and documentation
6. **Future**: Add autocomplete, man pages, and release binaries

## 🎯 Key Features

- ✅ Clean command structure with Cobra
- ✅ Support both HTTP-API and TCP connections
- ✅ JWT authentication with local token storage
- ✅ Colored, formatted output
- ✅ Interactive prompts for sensitive data
- ✅ Config file management
- ✅ Real-time progress sync via TCP
- ✅ Graceful error handling
- ✅ Extensible architecture

---

**Ready to start implementation?** Let's begin with Task 1! 🚀
