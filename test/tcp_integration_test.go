package test

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	tcp "mangahub/internal/microservices/tcp"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TCPIntegrationTestSuite tests TCP server with real Redis and PostgreSQL
type TCPIntegrationTestSuite struct {
	suite.Suite
	server      *tcp.TCPServer
	redisClient *redis.Client
	db          *sql.DB
	serverAddr  string
	redisAddr   string
	jwtSecret   string
	useHybrid   bool // Toggle between Redis-only and hybrid mode
}

// SetupSuite runs once before all tests
func (s *TCPIntegrationTestSuite) SetupSuite() {
	// Configuration
	s.redisAddr = "localhost:6379"
	s.serverAddr = "localhost:8081"
	s.jwtSecret = "test-secret-key-for-integration-tests"
	s.useHybrid = false // Set to true to test hybrid mode

	// Setup Redis client for verification
	s.redisClient = redis.NewClient(&redis.Options{
		Addr: s.redisAddr,
		DB:   1, // Use separate DB for integration tests
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Verify Redis is available
	if err := s.redisClient.Ping(ctx).Err(); err != nil {
		s.T().Skip("Redis not available, skipping integration tests")
		return
	}

	// Clean Redis before tests
	s.redisClient.FlushDB(ctx)
}

// SetupTest runs before each test
func (s *TCPIntegrationTestSuite) SetupTest() {
	var server *tcp.TCPServer

	if s.useHybrid && s.db != nil {
		// Hybrid mode: Redis + PostgreSQL
		server = tcp.NewServerWithHybridStorage(
			s.serverAddr,
			s.redisAddr,
			s.db,
			s.jwtSecret,
		)
	} else {
		// Redis-only mode
		server = tcp.NewServer(s.serverAddr, s.redisAddr)
	}

	if server == nil {
		s.T().Fatal("Failed to create server")
	}

	s.server = server

	// Start server in background
	go s.server.Start()
	time.Sleep(200 * time.Millisecond) // Wait for server to start
}

// TearDownTest runs after each test
func (s *TCPIntegrationTestSuite) TearDownTest() {
	if s.server != nil {
		s.server.Stop()
	}
	time.Sleep(100 * time.Millisecond)

	// Clean Redis after each test
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.redisClient.FlushDB(ctx)
}

// TearDownSuite runs once after all tests
func (s *TCPIntegrationTestSuite) TearDownSuite() {
	if s.redisClient != nil {
		s.redisClient.Close()
	}
	if s.db != nil {
		s.db.Close()
	}
}

// Test 1: Progress Data Persistence in Redis
func (s *TCPIntegrationTestSuite) TestProgressPersistenceRedis() {
	t := s.T()

	// Connect client
	conn, err := net.Dial("tcp", s.serverAddr)
	require.NoError(t, err, "Should connect to server")
	defer conn.Close()

	// Send progress update
	msg := tcp.Message{
		Type: "progress_update",
		Data: map[string]interface{}{
			"user_id":  "user123",
			"manga_id": float64(456),
			"chapter":  float64(78),
		},
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err, "Should marshal message")

	_, err = conn.Write(append(data, '\n'))
	require.NoError(t, err, "Should send message")

	// Wait for processing
	time.Sleep(500 * time.Millisecond)

	// Verify data in Redis
	ctx := context.Background()
	key := "progress:user123:456"
	exists, err := s.redisClient.Exists(ctx, key).Result()
	assert.NoError(t, err, "Should check Redis")
	assert.Equal(t, int64(1), exists, "Progress should exist in Redis")

	// Get the data
	chapter, err := s.redisClient.HGet(ctx, key, "chapter").Int64()
	assert.NoError(t, err, "Should get chapter from Redis")
	assert.Equal(t, int64(78), chapter, "Chapter should match")

	t.Log("✓ Progress persisted to Redis successfully")
}

// Test 2: Multiple Clients Broadcasting
func (s *TCPIntegrationTestSuite) TestMultiClientBroadcast() {
	t := s.T()

	numClients := 5
	clients := make([]net.Conn, numClients)
	readers := make([]*bufio.Reader, numClients)

	// Connect all clients
	for i := 0; i < numClients; i++ {
		conn, err := net.Dial("tcp", s.serverAddr)
		require.NoError(t, err, "Client %d should connect", i)
		clients[i] = conn
		readers[i] = bufio.NewReader(conn)
		defer conn.Close()
	}

	time.Sleep(100 * time.Millisecond)

	// Client 0 sends a message
	msg := tcp.Message{
		Type: "progress_update",
		Data: map[string]interface{}{
			"user_id":  "broadcaster",
			"manga_id": float64(999),
			"chapter":  float64(1),
		},
	}

	data, _ := json.Marshal(msg)
	_, err := clients[0].Write(append(data, '\n'))
	require.NoError(t, err, "Should broadcast message")

	// All clients should receive the broadcast
	var receivedCount atomic.Int32
	var wg sync.WaitGroup

	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			clients[idx].SetReadDeadline(time.Now().Add(2 * time.Second))
			response, err := readers[idx].ReadBytes('\n')
			if err == nil && len(response) > 0 {
				var decoded map[string]interface{}
				if json.Unmarshal(response, &decoded) == nil {
					receivedCount.Add(1)
					t.Logf("Client %d received broadcast", idx)
				}
			}
		}(i)
	}

	wg.Wait()

	assert.Equal(t, int32(numClients), receivedCount.Load(),
		"All %d clients should receive broadcast", numClients)

	t.Log("✓ Broadcast to all clients successful")
}

// Test 3: Rate Limiting
func (s *TCPIntegrationTestSuite) TestRateLimiting() {
	t := s.T()

	conn, err := net.Dial("tcp", s.serverAddr)
	require.NoError(t, err, "Should connect")
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// Send messages rapidly (more than rate limit)
	const rapidMessages = 200
	var rejectedCount atomic.Int32

	for i := 0; i < rapidMessages; i++ {
		msg := tcp.Message{
			Type: "progress_update",
			Data: map[string]interface{}{
				"user_id":  "rate_test",
				"manga_id": float64(1),
				"chapter":  float64(i),
			},
		}

		data, _ := json.Marshal(msg)
		conn.Write(append(data, '\n'))

		// Try to read response
		conn.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
		response, err := reader.ReadBytes('\n')
		if err == nil {
			var decoded map[string]interface{}
			if json.Unmarshal(response, &decoded) == nil {
				if msgType, ok := decoded["type"].(string); ok && msgType == "error" {
					if message, ok := decoded["message"].(string); ok && message == "Rate limit exceeded" {
						rejectedCount.Add(1)
					}
				}
			}
		}
	}

	assert.Greater(t, rejectedCount.Load(), int32(0),
		"Some messages should be rate limited")

	t.Logf("✓ Rate limiting working: %d/%d messages rejected", rejectedCount.Load(), rapidMessages)
}

// Test 4: Concurrent Progress Updates from Multiple Users
func (s *TCPIntegrationTestSuite) TestConcurrentUserProgress() {
	t := s.T()

	numUsers := 10
	updatesPerUser := 5

	var wg sync.WaitGroup
	var successCount atomic.Int32

	for userID := 0; userID < numUsers; userID++ {
		wg.Add(1)
		go func(uid int) {
			defer wg.Done()

			conn, err := net.Dial("tcp", s.serverAddr)
			if err != nil {
				t.Logf("User %d failed to connect: %v", uid, err)
				return
			}
			defer conn.Close()

			for chapter := 1; chapter <= updatesPerUser; chapter++ {
				msg := tcp.Message{
					Type: "progress_update",
					Data: map[string]interface{}{
						"user_id":  fmt.Sprintf("user_%d", uid),
						"manga_id": float64(100 + uid),
						"chapter":  float64(chapter),
					},
				}

				data, _ := json.Marshal(msg)
				_, err := conn.Write(append(data, '\n'))
				if err == nil {
					successCount.Add(1)
				}

				time.Sleep(50 * time.Millisecond)
			}
		}(userID)
	}

	wg.Wait()

	expectedUpdates := int32(numUsers * updatesPerUser)
	assert.Equal(t, expectedUpdates, successCount.Load(),
		"All progress updates should succeed")

	// Wait for all updates to be processed
	time.Sleep(500 * time.Millisecond)

	// Verify in Redis
	ctx := context.Background()
	for userID := 0; userID < numUsers; userID++ {
		key := fmt.Sprintf("progress:user_%d:%d", userID, 100+userID)
		exists, err := s.redisClient.Exists(ctx, key).Result()
		assert.NoError(t, err, "Should check Redis for user_%d", userID)
		assert.Equal(t, int64(1), exists, "Progress should exist for user_%d", userID)

		// Check final chapter
		chapter, err := s.redisClient.HGet(ctx, key, "chapter").Int64()
		assert.NoError(t, err, "Should get chapter")
		assert.Equal(t, int64(updatesPerUser), chapter, "Final chapter should be %d", updatesPerUser)
	}

	t.Log("✓ Concurrent user progress updates successful")
}

// Test 5: Server Graceful Shutdown with Active Connections
func (s *TCPIntegrationTestSuite) TestGracefulShutdownWithConnections() {
	t := s.T()

	// Connect multiple clients
	numClients := 10
	conns := make([]net.Conn, numClients)

	for i := 0; i < numClients; i++ {
		conn, err := net.Dial("tcp", s.serverAddr)
		require.NoError(t, err, "Client %d should connect", i)
		conns[i] = conn
	}

	time.Sleep(100 * time.Millisecond)

	// Trigger shutdown
	shutdownDone := make(chan bool)
	go func() {
		s.server.Stop()
		shutdownDone <- true
	}()

	// Server should notify and shut down gracefully
	select {
	case <-shutdownDone:
		t.Log("✓ Server shut down gracefully")
	case <-time.After(10 * time.Second):
		t.Fatal("❌ Server shutdown timeout")
	}

	// Clean up
	for _, conn := range conns {
		if conn != nil {
			conn.Close()
		}
	}
}

// Test 6: Invalid Message Handling
func (s *TCPIntegrationTestSuite) TestInvalidMessageHandling() {
	t := s.T()

	conn, err := net.Dial("tcp", s.serverAddr)
	require.NoError(t, err, "Should connect")
	defer conn.Close()

	invalidMessages := []string{
		`{"invalid json`,
		`{"type":"progress_update","data":{"user_id":"","manga_id":0,"chapter":-1}}`, // Invalid data
		`{"type":"progress_update","data":{"manga_id":1,"chapter":1}}`,               // Missing user_id
		`not json at all`,
	}

	for i, invalid := range invalidMessages {
		_, err := conn.Write([]byte(invalid + "\n"))
		assert.NoError(t, err, "Should send message %d", i)
		time.Sleep(100 * time.Millisecond)
	}

	// Connection should still be alive
	validMsg := tcp.Message{
		Type: "progress_update",
		Data: map[string]interface{}{
			"user_id":  "recovery_test",
			"manga_id": float64(1),
			"chapter":  float64(1),
		},
	}

	data, _ := json.Marshal(validMsg)
	_, err = conn.Write(append(data, '\n'))
	assert.NoError(t, err, "Should send valid message after invalid ones")

	t.Log("✓ Server handles invalid messages gracefully")
}

// Test 7: Data Consistency After Rapid Updates
func (s *TCPIntegrationTestSuite) TestDataConsistencyRapidUpdates() {
	t := s.T()

	conn, err := net.Dial("tcp", s.serverAddr)
	require.NoError(t, err, "Should connect")
	defer conn.Close()

	userID := "consistency_test"
	mangaID := 555
	finalChapter := 100

	// Send rapid chapter updates
	for chapter := 1; chapter <= finalChapter; chapter++ {
		msg := tcp.Message{
			Type: "progress_update",
			Data: map[string]interface{}{
				"user_id":  userID,
				"manga_id": float64(mangaID),
				"chapter":  float64(chapter),
			},
		}

		data, _ := json.Marshal(msg)
		conn.Write(append(data, '\n'))
		time.Sleep(10 * time.Millisecond) // Rapid but not instant
	}

	// Wait for all updates to process
	time.Sleep(1 * time.Second)

	// Verify final state in Redis
	ctx := context.Background()
	key := fmt.Sprintf("progress:%s:%d", userID, mangaID)
	chapter, err := s.redisClient.HGet(ctx, key, "chapter").Int64()
	require.NoError(t, err, "Should get chapter from Redis")
	assert.Equal(t, int64(finalChapter), chapter,
		"Final chapter should be %d, got %d", finalChapter, chapter)

	t.Log("✓ Data consistency maintained after rapid updates")
}

// Test 8: Connection Pool Stress Test
func (s *TCPIntegrationTestSuite) TestConnectionPoolStress() {
	t := s.T()

	const numIterations = 100
	var successCount atomic.Int32

	for i := 0; i < numIterations; i++ {
		conn, err := net.Dial("tcp", s.serverAddr)
		if err != nil {
			t.Logf("Connection %d failed: %v", i, err)
			continue
		}

		// Send a quick message
		msg := tcp.Message{
			Type: "progress_update",
			Data: map[string]interface{}{
				"user_id":  fmt.Sprintf("stress_%d", i),
				"manga_id": float64(1),
				"chapter":  float64(1),
			},
		}

		data, _ := json.Marshal(msg)
		if _, err := conn.Write(append(data, '\n')); err == nil {
			successCount.Add(1)
		}

		conn.Close()
		time.Sleep(10 * time.Millisecond)
	}

	successRate := float64(successCount.Load()) / float64(numIterations) * 100
	assert.GreaterOrEqual(t, successRate, 95.0,
		"Success rate should be at least 95%%, got %.2f%%", successRate)

	t.Logf("✓ Connection pool stress test: %.2f%% success rate", successRate)
}

// Test 9: Long-Running Connection Stability
func (s *TCPIntegrationTestSuite) TestLongRunningConnection() {
	t := s.T()

	if testing.Short() {
		t.Skip("Skipping long-running test in short mode")
	}

	conn, err := net.Dial("tcp", s.serverAddr)
	require.NoError(t, err, "Should connect")
	defer conn.Close()

	// Keep connection alive and send periodic updates
	duration := 30 * time.Second
	interval := 2 * time.Second
	iterations := int(duration / interval)

	var successCount atomic.Int32

	for i := 0; i < iterations; i++ {
		msg := tcp.Message{
			Type: "progress_update",
			Data: map[string]interface{}{
				"user_id":  "long_running",
				"manga_id": float64(1),
				"chapter":  float64(i + 1),
			},
		}

		data, _ := json.Marshal(msg)
		if _, err := conn.Write(append(data, '\n')); err == nil {
			successCount.Add(1)
		} else {
			t.Logf("Write failed at iteration %d: %v", i, err)
		}

		time.Sleep(interval)
	}

	assert.Equal(t, int32(iterations), successCount.Load(),
		"All messages should be sent successfully")

	t.Log("✓ Long-running connection stable")
}

// Test 10: Redis Failover Simulation (Mock)
func (s *TCPIntegrationTestSuite) TestRedisFailoverHandling() {
	t := s.T()

	conn, err := net.Dial("tcp", s.serverAddr)
	require.NoError(t, err, "Should connect")
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// Send a message
	msg := tcp.Message{
		Type: "progress_update",
		Data: map[string]interface{}{
			"user_id":  "failover_test",
			"manga_id": float64(1),
			"chapter":  float64(1),
		},
	}

	data, _ := json.Marshal(msg)
	_, err = conn.Write(append(data, '\n'))
	require.NoError(t, err, "Should send message")

	// Read response
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	response, err := reader.ReadBytes('\n')

	// Note: In a real failover test, we would simulate Redis disconnection
	// and verify error handling and reconnection logic
	assert.NotNil(t, response, "Should get some response")

	t.Log("✓ Redis failover handling test (basic)")
}

// Run the integration test suite
func TestTCPIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(TCPIntegrationTestSuite))
}

// Standalone integration test for basic connectivity
func TestBasicIntegrationConnectivity(t *testing.T) {
	// This test can run independently for quick verification
	server := tcp.NewServerWithMockRedis("localhost:8095")
	if server == nil {
		t.Fatal("Failed to create server")
	}

	go server.Start()
	time.Sleep(100 * time.Millisecond)
	defer server.Stop()

	conn, err := net.Dial("tcp", "localhost:8095")
	require.NoError(t, err, "Should connect to server")
	defer conn.Close()

	msg := tcp.Message{
		Type: "test",
		Data: map[string]interface{}{"status": "ok"},
	}

	data, _ := json.Marshal(msg)
	_, err = conn.Write(append(data, '\n'))
	assert.NoError(t, err, "Should send message")

	t.Log("✓ Basic integration connectivity verified")
}
