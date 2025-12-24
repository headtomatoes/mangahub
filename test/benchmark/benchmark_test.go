// Package benchmark provides comprehensive benchmark tests for MangaHub
// Run with: go test -v -timeout 10m ./test/benchmark/...
package benchmark

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// ============================================================================
// CONFIGURATION
// ============================================================================

const (
	// Service endpoints
	APIBaseURL = "http://10.238.20.112:8084"
	TCPAddr    = "10.238.20.112:8081"
	WSEndpoint = "ws://10.238.20.112:8084/ws"

	// Test credentials (create a test user first)
	TestUsername = "benchmarkuser"
	TestPassword = "benchmark123"
	TestEmail    = "benchmark@test.com"
)

// BenchmarkConfig holds test configuration
type BenchmarkConfig struct {
	APIBaseURL      string
	TCPAddr         string
	WSEndpoint      string
	ConcurrentUsers int
	TCPConnections  int
	WSConnections   int
	SearchTimeout   time.Duration
	TestDuration    time.Duration
}

// DefaultConfig returns default benchmark configuration matching your targets
func DefaultConfig() BenchmarkConfig {
	return BenchmarkConfig{
		APIBaseURL:      APIBaseURL,
		TCPAddr:         TCPAddr,
		WSEndpoint:      WSEndpoint,
		ConcurrentUsers: 100,                    // Target: 50-100 concurrent users
		TCPConnections:  30,                     // Target: 20-30 TCP connections
		WSConnections:   20,                     // Target: 10-20 WebSocket users
		SearchTimeout:   500 * time.Millisecond, // Target: <500ms
		TestDuration:    30 * time.Second,
	}
}

// ============================================================================
// TEST UTILITIES
// ============================================================================

type TestClient struct {
	httpClient *http.Client
	baseURL    string
	token      string
}

func NewTestClient(baseURL string) *TestClient {
	return &TestClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: baseURL,
	}
}

// Register creates a new test user
func (c *TestClient) Register(username, email, password string) error {
	payload := map[string]string{
		"username": username,
		"email":    email,
		"password": password,
	}
	body, _ := json.Marshal(payload)

	resp, err := c.httpClient.Post(c.baseURL+"/auth/register", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 201 Created or 409 Conflict (already exists) are both OK
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("register failed: %d - %s", resp.StatusCode, string(bodyBytes))
	}
	return nil
}

// Login authenticates and stores the JWT token
func (c *TestClient) Login(username, password string) error {
	payload := map[string]string{
		"username": username,
		"password": password,
	}
	body, _ := json.Marshal(payload)

	resp, err := c.httpClient.Post(c.baseURL+"/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	c.token = result.AccessToken
	return nil
}

// AuthenticatedGet performs an authenticated GET request
func (c *TestClient) AuthenticatedGet(path string) (*http.Response, error) {
	req, err := http.NewRequest("GET", c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	return c.httpClient.Do(req)
}

// ============================================================================
// BENCHMARK 1: CONCURRENT USERS (50-100)
// ============================================================================

func TestConcurrentUsers(t *testing.T) {
	cfg := DefaultConfig()

	t.Logf("🎯 Target: Support %d concurrent users", cfg.ConcurrentUsers)
	t.Log("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Setup: Create and login test user
	client := NewTestClient(cfg.APIBaseURL)
	if err := client.Register(TestUsername, TestEmail, TestPassword); err != nil {
		t.Logf("Registration note: %v (may already exist)", err)
	}
	if err := client.Login(TestUsername, TestPassword); err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	// Metrics
	var (
		successCount int64
		failureCount int64
		totalLatency int64
		maxLatency   int64
	)

	// Run concurrent requests
	var wg sync.WaitGroup
	startTime := time.Now()

	for i := 0; i < cfg.ConcurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			reqStart := time.Now()
			resp, err := client.AuthenticatedGet("/api/manga?page=1&page_size=10")
			latency := time.Since(reqStart).Milliseconds()

			if err != nil {
				atomic.AddInt64(&failureCount, 1)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				atomic.AddInt64(&successCount, 1)
				atomic.AddInt64(&totalLatency, latency)

				// Track max latency
				for {
					current := atomic.LoadInt64(&maxLatency)
					if latency <= current || atomic.CompareAndSwapInt64(&maxLatency, current, latency) {
						break
					}
				}
			} else {
				atomic.AddInt64(&failureCount, 1)
			}
		}(i)
	}

	wg.Wait()
	totalTime := time.Since(startTime)

	// Report results
	success := atomic.LoadInt64(&successCount)
	failures := atomic.LoadInt64(&failureCount)
	avgLatency := float64(0)
	if success > 0 {
		avgLatency = float64(atomic.LoadInt64(&totalLatency)) / float64(success)
	}

	t.Log("")
	t.Log("📊 Results:")
	t.Logf("   ├── Total Requests:    %d", cfg.ConcurrentUsers)
	t.Logf("   ├── Successful:        %d (%.1f%%)", success, float64(success)/float64(cfg.ConcurrentUsers)*100)
	t.Logf("   ├── Failed:            %d", failures)
	t.Logf("   ├── Avg Latency:       %.2f ms", avgLatency)
	t.Logf("   ├── Max Latency:       %d ms", atomic.LoadInt64(&maxLatency))
	t.Logf("   └── Total Time:        %v", totalTime)
	t.Log("")

	// Pass/Fail criteria
	successRate := float64(success) / float64(cfg.ConcurrentUsers) * 100
	if successRate >= 95 {
		t.Logf("✅ PASS: %.1f%% success rate (target: ≥95%%)", successRate)
	} else {
		t.Errorf("❌ FAIL: %.1f%% success rate (target: ≥95%%)", successRate)
	}
}

// ============================================================================
// BENCHMARK 2: SEARCH LATENCY (<500ms)
// ============================================================================

func TestSearchLatency(t *testing.T) {
	cfg := DefaultConfig()

	t.Logf("🎯 Target: Search queries < %v", cfg.SearchTimeout)
	t.Log("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Setup
	client := NewTestClient(cfg.APIBaseURL)
	if err := client.Register(TestUsername, TestEmail, TestPassword); err != nil {
		t.Logf("Registration note: %v", err)
	}
	if err := client.Login(TestUsername, TestPassword); err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	// Test queries
	queries := []string{
		"naruto",
		"one piece",
		"attack",
		"dragon",
		"hero",
	}

	var latencies []time.Duration
	var failures int

	for _, q := range queries {
		start := time.Now()

		path := "/api/manga/search?" + url.Values{"q": {q}}.Encode()
		resp, err := client.AuthenticatedGet(path)

		latency := time.Since(start)
		latencies = append(latencies, latency)

		if err != nil {
			t.Logf("   ❌ Query '%s': ERROR - %v", q, err)
			failures++
			continue
		}
		defer resp.Body.Close()

		// Read response to measure full latency
		io.ReadAll(resp.Body)

		status := "✅"
		if latency > cfg.SearchTimeout {
			status = "⚠️"
		}
		t.Logf("   %s Query '%s': %v (status: %d)", status, q, latency, resp.StatusCode)
	}

	// Calculate statistics
	var total time.Duration
	var max time.Duration
	var withinTarget int

	for _, l := range latencies {
		total += l
		if l > max {
			max = l
		}
		if l <= cfg.SearchTimeout {
			withinTarget++
		}
	}

	avg := total / time.Duration(len(latencies))

	t.Log("")
	t.Log("📊 Results:")
	t.Logf("   ├── Queries Tested:    %d", len(queries))
	t.Logf("   ├── Avg Latency:       %v", avg)
	t.Logf("   ├── Max Latency:       %v", max)
	t.Logf("   ├── Within Target:     %d/%d (%.1f%%)", withinTarget, len(queries), float64(withinTarget)/float64(len(queries))*100)
	t.Logf("   └── Failures:          %d", failures)
	t.Log("")

	// Pass/Fail
	if avg <= cfg.SearchTimeout {
		t.Logf("✅ PASS: Average latency %v ≤ %v target", avg, cfg.SearchTimeout)
	} else {
		t.Errorf("❌ FAIL: Average latency %v > %v target", avg, cfg.SearchTimeout)
	}
}

// ============================================================================
// BENCHMARK 3: TCP CONNECTIONS (20-30)
// ============================================================================

func TestTCPConnections(t *testing.T) {
	cfg := DefaultConfig()

	t.Logf("🎯 Target: Support %d concurrent TCP connections", cfg.TCPConnections)
	t.Log("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Get auth token first
	client := NewTestClient(cfg.APIBaseURL)
	if err := client.Register(TestUsername, TestEmail, TestPassword); err != nil {
		t.Logf("Registration note: %v", err)
	}
	if err := client.Login(TestUsername, TestPassword); err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	var (
		connectedCount int64
		authSuccess    int64
		authFailed     int64
		connections    []net.Conn
		mu             sync.Mutex
	)

	var wg sync.WaitGroup

	// Establish connections concurrently
	for i := 0; i < cfg.TCPConnections; i++ {
		wg.Add(1)
		go func(connID int) {
			defer wg.Done()

			// Connect
			conn, err := net.DialTimeout("tcp", cfg.TCPAddr, 5*time.Second)
			if err != nil {
				t.Logf("   ❌ Connection %d: dial failed - %v", connID, err)
				return
			}

			atomic.AddInt64(&connectedCount, 1)

			// Send auth message
			authMsg := map[string]interface{}{
				"type": "auth",
				"data": map[string]string{
					"token": client.token,
				},
			}
			authJSON, _ := json.Marshal(authMsg)
			authJSON = append(authJSON, '\n')

			conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if _, err := conn.Write(authJSON); err != nil {
				t.Logf("   ❌ Connection %d: auth write failed - %v", connID, err)
				conn.Close()
				return
			}

			// Read auth response
			conn.SetReadDeadline(time.Now().Add(5 * time.Second))
			reader := bufio.NewReader(conn)
			response, err := reader.ReadBytes('\n')
			if err != nil {
				t.Logf("   ❌ Connection %d: auth read failed - %v", connID, err)
				conn.Close()
				return
			}

			var resp map[string]interface{}
			if err := json.Unmarshal(response, &resp); err == nil {
				if resp["type"] == "auth_success" {
					atomic.AddInt64(&authSuccess, 1)
					mu.Lock()
					connections = append(connections, conn)
					mu.Unlock()
					return
				}
			}

			atomic.AddInt64(&authFailed, 1)
			conn.Close()
		}(i)
	}

	wg.Wait()

	// Keep connections alive for a moment to verify they're stable
	time.Sleep(2 * time.Second)

	// Check how many are still connected
	var stillConnected int
	mu.Lock()
	for _, conn := range connections {
		conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		one := make([]byte, 1)
		_, err := conn.Read(one)
		if err == nil || err.Error() == "i/o timeout" || err == io.EOF {
			stillConnected++
		}
	}
	mu.Unlock()

	// Cleanup
	mu.Lock()
	for _, conn := range connections {
		conn.Close()
	}
	mu.Unlock()

	t.Log("")
	t.Log("📊 Results:")
	t.Logf("   ├── Target Connections: %d", cfg.TCPConnections)
	t.Logf("   ├── TCP Connected:      %d", atomic.LoadInt64(&connectedCount))
	t.Logf("   ├── Auth Success:       %d", atomic.LoadInt64(&authSuccess))
	t.Logf("   ├── Auth Failed:        %d", atomic.LoadInt64(&authFailed))
	t.Logf("   └── Still Connected:    %d", stillConnected)
	t.Log("")

	// Pass/Fail
	authSuccessCount := atomic.LoadInt64(&authSuccess)
	if authSuccessCount >= int64(cfg.TCPConnections*80/100) {
		t.Logf("✅ PASS: %d/%d connections authenticated (≥80%%)", authSuccessCount, cfg.TCPConnections)
	} else {
		t.Errorf("❌ FAIL: Only %d/%d connections authenticated (<80%%)", authSuccessCount, cfg.TCPConnections)
	}
}

// ============================================================================
// BENCHMARK 4: WEBSOCKET CONNECTIONS (10-20)
// ============================================================================

func TestWebSocketConnections(t *testing.T) {
	cfg := DefaultConfig()

	t.Logf("🎯 Target: Support %d concurrent WebSocket connections", cfg.WSConnections)
	t.Log("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Get auth token
	client := NewTestClient(cfg.APIBaseURL)
	if err := client.Register(TestUsername, TestEmail, TestPassword); err != nil {
		t.Logf("Registration note: %v", err)
	}
	if err := client.Login(TestUsername, TestPassword); err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	var (
		connectedCount   int64
		messagesSent     int64
		messagesReceived int64
		wsConns          []*websocket.Conn
		mu               sync.Mutex
	)

	var wg sync.WaitGroup

	// Establish WebSocket connections
	for i := 0; i < cfg.WSConnections; i++ {
		wg.Add(1)
		go func(connID int) {
			defer wg.Done()

			header := http.Header{}
			header.Add("Authorization", "Bearer "+client.token)

			dialer := websocket.Dialer{
				HandshakeTimeout: 10 * time.Second,
			}

			conn, _, err := dialer.Dial(cfg.WSEndpoint, header)
			if err != nil {
				t.Logf("   ❌ WS %d: dial failed - %v", connID, err)
				return
			}

			atomic.AddInt64(&connectedCount, 1)

			mu.Lock()
			wsConns = append(wsConns, conn)
			mu.Unlock()

			// Join a room
			joinMsg := map[string]interface{}{
				"type":    "join",
				"room_id": int64(1), // Test room
			}
			if err := conn.WriteJSON(joinMsg); err != nil {
				t.Logf("   ⚠️ WS %d: join write failed - %v", connID, err)
				return
			}
			atomic.AddInt64(&messagesSent, 1)

			// Read join confirmation
			conn.SetReadDeadline(time.Now().Add(5 * time.Second))
			var response map[string]interface{}
			if err := conn.ReadJSON(&response); err == nil {
				atomic.AddInt64(&messagesReceived, 1)
			}
		}(i)
	}

	wg.Wait()

	// Test message broadcast
	t.Log("")
	t.Log("📡 Testing message broadcast...")

	mu.Lock()
	if len(wsConns) > 0 {
		// Send a chat message from first connection
		chatMsg := map[string]interface{}{
			"type":    "chat",
			"content": "Benchmark test message",
		}
		wsConns[0].WriteJSON(chatMsg)
		atomic.AddInt64(&messagesSent, 1)

		// Give time for broadcast
		time.Sleep(500 * time.Millisecond)

		// Try to read from other connections
		for i := 1; i < len(wsConns) && i < 5; i++ {
			wsConns[i].SetReadDeadline(time.Now().Add(500 * time.Millisecond))
			var msg map[string]interface{}
			if err := wsConns[i].ReadJSON(&msg); err == nil {
				atomic.AddInt64(&messagesReceived, 1)
			}
		}
	}
	mu.Unlock()

	// Cleanup
	mu.Lock()
	for _, conn := range wsConns {
		conn.Close()
	}
	mu.Unlock()

	t.Log("")
	t.Log("📊 Results:")
	t.Logf("   ├── Target Connections: %d", cfg.WSConnections)
	t.Logf("   ├── WS Connected:       %d", atomic.LoadInt64(&connectedCount))
	t.Logf("   ├── Messages Sent:      %d", atomic.LoadInt64(&messagesSent))
	t.Logf("   └── Messages Received:  %d", atomic.LoadInt64(&messagesReceived))
	t.Log("")

	// Pass/Fail
	connected := atomic.LoadInt64(&connectedCount)
	if connected >= int64(cfg.WSConnections*80/100) {
		t.Logf("✅ PASS: %d/%d WebSocket connections established (≥80%%)", connected, cfg.WSConnections)
	} else {
		t.Errorf("❌ FAIL: Only %d/%d WebSocket connections established (<80%%)", connected, cfg.WSConnections)
	}
}

// ============================================================================
// BENCHMARK 5: DATABASE CAPACITY (30-40 manga)
// ============================================================================

func TestDatabaseCapacity(t *testing.T) {
	t.Log("🎯 Target: Handle 30-40 manga series in database")
	t.Log("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	cfg := DefaultConfig()

	// Setup
	client := NewTestClient(cfg.APIBaseURL)
	if err := client.Register(TestUsername, TestEmail, TestPassword); err != nil {
		t.Logf("Registration note: %v", err)
	}
	if err := client.Login(TestUsername, TestPassword); err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	// Get manga count
	resp, err := client.AuthenticatedGet("/api/manga?page=1&page_size=100")
	if err != nil {
		t.Fatalf("Failed to get manga: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data  []interface{} `json:"data"`
		Total int           `json:"total"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	t.Log("")
	t.Log("📊 Results:")
	t.Logf("   ├── Manga in Database:  %d", result.Total)
	t.Logf("   ├── Target Range:       30-40")
	t.Logf("   └── Returned in Page:   %d", len(result.Data))
	t.Log("")

	if result.Total >= 30 {
		t.Logf("✅ PASS: %d manga in database (target: ≥30)", result.Total)
	} else {
		t.Logf("⚠️ INFO: %d manga in database (target: ≥30) - consider importing more data", result.Total)
	}
}

// ============================================================================
// BENCHMARK 6: ERROR HANDLING & GRACEFUL DEGRADATION
// ============================================================================

func TestErrorHandling(t *testing.T) {
	t.Log("🎯 Target: Basic error handling and recovery")
	t.Log("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	cfg := DefaultConfig()
	client := NewTestClient(cfg.APIBaseURL)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{"Invalid endpoint", "GET", "/api/nonexistent", http.StatusNotFound},
		{"Unauthorized access", "GET", "/api/manga", http.StatusUnauthorized},
		{"Invalid manga ID", "GET", "/api/manga/99999999", http.StatusNotFound},
		{"Health check", "GET", "/check-conn", http.StatusOK},
	}

	var passed, failed int

	for _, tc := range tests {
		req, _ := http.NewRequest(tc.method, cfg.APIBaseURL+tc.path, nil)
		resp, err := client.httpClient.Do(req)

		if err != nil {
			t.Logf("   ❌ %s: request failed - %v", tc.name, err)
			failed++
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == tc.expectedStatus {
			t.Logf("   ✅ %s: got expected status %d", tc.name, resp.StatusCode)
			passed++
		} else {
			t.Logf("   ❌ %s: expected %d, got %d", tc.name, tc.expectedStatus, resp.StatusCode)
			failed++
		}
	}

	t.Log("")
	t.Log("📊 Results:")
	t.Logf("   ├── Tests Passed: %d/%d", passed, len(tests))
	t.Logf("   └── Tests Failed: %d/%d", failed, len(tests))
	t.Log("")

	if passed == len(tests) {
		t.Log("✅ PASS: All error handling tests passed")
	} else {
		t.Errorf("❌ FAIL: %d/%d error handling tests failed", failed, len(tests))
	}
}

// ============================================================================
// COMPREHENSIVE BENCHMARK SUITE
// ============================================================================

func TestFullBenchmarkSuite(t *testing.T) {
	t.Log("")
	t.Log("╔══════════════════════════════════════════════════════════════════╗")
	t.Log("║           MANGAHUB COMPREHENSIVE BENCHMARK SUITE                 ║")
	t.Log("║                     Testing Phase Targets                        ║")
	t.Log("╚══════════════════════════════════════════════════════════════════╝")
	t.Log("")

	cfg := DefaultConfig()

	t.Log("📋 Configuration:")
	t.Logf("   ├── API URL:           %s", cfg.APIBaseURL)
	t.Logf("   ├── TCP Address:       %s", cfg.TCPAddr)
	t.Logf("   ├── WebSocket URL:     %s", cfg.WSEndpoint)
	t.Logf("   ├── Concurrent Users:  %d", cfg.ConcurrentUsers)
	t.Logf("   ├── TCP Connections:   %d", cfg.TCPConnections)
	t.Logf("   ├── WS Connections:    %d", cfg.WSConnections)
	t.Logf("   └── Search Timeout:    %v", cfg.SearchTimeout)
	t.Log("")

	// Run sub-tests
	t.Run("ConcurrentUsers", TestConcurrentUsers)
	t.Run("SearchLatency", TestSearchLatency)
	t.Run("TCPConnections", TestTCPConnections)
	t.Run("WebSocketConnections", TestWebSocketConnections)
	t.Run("DatabaseCapacity", TestDatabaseCapacity)
	t.Run("ErrorHandling", TestErrorHandling)

	t.Log("")
	t.Log("╔══════════════════════════════════════════════════════════════════╗")
	t.Log("║                    BENCHMARK SUITE COMPLETE                      ║")
	t.Log("╚══════════════════════════════════════════════════════════════════╝")
}
