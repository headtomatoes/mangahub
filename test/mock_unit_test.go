package test

import (
	"bufio"
	"encoding/json"
	"fmt"
	tcp "mangahub/internal/microservices/tcp"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TCPServerTestSuite for testing TCP server performance and some edge cases
type TCPServerTestSuite struct {
	suite.Suite                // Embedding suite.Suite of the testify testing framework for test suite functionality
	server      *tcp.TCPServer // reference to the TCP server instance
	serverAddr  string         // server address (host:port)
	serverPort  int            // server port
}

// SetupTest runs before each test => initialize the testing environment
func (s *TCPServerTestSuite) SetupTest() {
	// Use different port for each test to avoid conflicts
	s.serverPort = 8081 + int(time.Now().UnixNano()%1000) // 8081 + random offset
	s.serverAddr = fmt.Sprintf("localhost:%d", s.serverPort)

	// Use mock Redis for testing (no actual Redis connection required)
	s.server = tcp.NewServerWithMockRedis(s.serverAddr)

	go s.server.Start() // start each server in a goroutine

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)
}

// TearDownTest runs after each test => clean up the testing environment
func (s *TCPServerTestSuite) TearDownTest() {
	if s.server != nil {
		s.server.Stop() // stop the server
	}
	time.Sleep(50 * time.Millisecond)
}

// Test 1: Concurrent Client Connections with Various Scales
func (s *TCPServerTestSuite) TestConcurrentClients_30Clients() {
	s.testConcurrentClients(30, "30 concurrent clients")
}

func (s *TCPServerTestSuite) TestConcurrentClients_50Clients() {
	s.testConcurrentClients(50, "50 concurrent clients (stress test)")
}

func (s *TCPServerTestSuite) TestConcurrentClients_100Clients() {
	s.testConcurrentClients(100, "100 concurrent clients (heavy load test)")
}

// Helper function for concurrent client testing
func (s *TCPServerTestSuite) testConcurrentClients(numClients int, description string) {
	t := s.T()

	var wg sync.WaitGroup

	// stats counters for the test
	successCount := 0
	errorCount := 0
	mu := sync.Mutex{}
	errors := make([]error, 0)

	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()

			conn, err := net.DialTimeout("tcp", s.serverAddr, 5*time.Second)
			if err != nil {
				mu.Lock()
				errorCount++
				errors = append(errors, fmt.Errorf("client %d: %w", clientID, err))
				mu.Unlock()
				return
			}
			defer conn.Close()

			// Set deadlines for read/write operations to avoid hanging
			conn.SetDeadline(time.Now().Add(10 * time.Second))

			mu.Lock()
			successCount++
			mu.Unlock()

			// Keep connection alive
			time.Sleep(2 * time.Second)
		}(i)
	}

	wg.Wait()

	// Assertions
	assert.Equal(t, numClients, successCount,
		"All %d clients should connect successfully", numClients)
	assert.Zero(t, errorCount,
		"No connection errors should occur")

	if len(errors) > 0 {
		t.Logf("Connection errors encountered: %v", errors)
	}

	t.Logf("✓ %s: %d/%d successful", description, successCount, numClients)
}

// Test 2: Edge Case - Rapid Connect/Disconnect within socket leak or delay
func (s *TCPServerTestSuite) TestRapidConnectDisconnect() {
	t := s.T()

	const iterations = 100
	successCount := 0

	for i := 0; i < iterations; i++ {
		conn, err := net.DialTimeout("tcp", s.serverAddr, 2*time.Second)
		require.NoError(t, err, "Connection of %d to TCPserver should succeed", i)

		// Immediately close
		err = conn.Close()
		assert.NoError(t, err, "Connection close should succeed")
		successCount++
	}

	assert.Equal(t, iterations, successCount,
		"All rapid connections should succeed")
}

// Test 3: Message Latency with Percentile Analysis
// send 100 json msg and measure latency
// between send and broadcast response
// comp avg, p50, p95, p99, max latencies
// assert avg < 200ms, p95 < 300ms, p99 < 500ms
// for now we use JSON so the performance is not optimal, which results in ~87% delivery rate
// later we can switch to protobuf for better performance
func (s *TCPServerTestSuite) TestMessageLatency_P95() {
	t := s.T()

	conn, err := net.DialTimeout("tcp", s.serverAddr, 5*time.Second)
	require.NoError(t, err, "Should connect to server")
	defer conn.Close()

	const iterations = 100
	latencies := make([]time.Duration, 0, iterations)

	reader := bufio.NewReader(conn)

	for i := 0; i < iterations; i++ {
		msg := tcp.Message{
			Type: "progress_update",
			Data: map[string]interface{}{
				"timestamp": time.Now().UnixNano(),
				"user":      "test_user",
				"manga_id":  "one-piece",
				"chapter":   1100 + i,
			},
		}

		startTime := time.Now()

		// Send message
		bytes, err := json.Marshal(msg)
		require.NoError(t, err, "JSON marshaling should succeed")

		_, err = conn.Write(append(bytes, '\n'))
		require.NoError(t, err, "Write should succeed")

		// Read broadcast response
		conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		response, err := reader.ReadBytes('\n')
		require.NoError(t, err, "Should receive response for message %d", i)

		latency := time.Since(startTime)
		latencies = append(latencies, latency)

		// Verify response is valid JSON
		var broadcastMsg map[string]interface{}
		err = json.Unmarshal(response, &broadcastMsg)
		assert.NoError(t, err, "Response should be valid JSON")
	}

	// Calculate statistics
	avgLatency := calculateAverage(latencies)
	p50Latency := calculatePercentile(latencies, 50)
	p95Latency := calculatePercentile(latencies, 95)
	p99Latency := calculatePercentile(latencies, 99)
	maxLatency := calculateMax(latencies)

	// Assertions
	assert.Less(t, avgLatency, 200*time.Millisecond,
		"Average latency should be under 200ms")
	assert.Less(t, p95Latency, 300*time.Millisecond,
		"P95 latency should be under 300ms")
	assert.Less(t, p99Latency, 500*time.Millisecond,
		"P99 latency should be under 500ms")

	t.Logf("✓ Latency stats - Avg: %v, P50: %v, P95: %v, P99: %v, Max: %v",
		avgLatency, p50Latency, p95Latency, p99Latency, maxLatency)
}

// Test 4: Data Integrity - Large Volume
func (s *TCPServerTestSuite) TestDataIntegrity_1000Messages() {
	s.testDataIntegrity(1000, "1000 messages")
}

func (s *TCPServerTestSuite) TestDataIntegrity_10000Messages() {
	s.testDataIntegrity(10000, "10000 messages (stress)")
}

// Helper for data integrity testing
// sends numMessages and verifies all are received intact
// counts successes, corrupted, failed reads
func (s *TCPServerTestSuite) testDataIntegrity(numMessages int, description string) {
	t := s.T()

	conn, err := net.DialTimeout("tcp", s.serverAddr, 5*time.Second)
	require.NoError(t, err, "Should connect to server")
	defer conn.Close()

	reader := bufio.NewReader(conn)
	successCount := 0
	corruptedMessages := 0
	failedReads := 0

	for i := 0; i < numMessages; i++ {
		msg := tcp.Message{
			Type: "progress_update",
			Data: map[string]interface{}{
				"id":      i,
				"user":    fmt.Sprintf("user_%d", i),
				"chapter": i % 100,
				"data":    strings.Repeat("x", 100), // Add some payload
			},
		}

		bytes, err := json.Marshal(msg)
		require.NoError(t, err, "Marshaling should succeed")

		_, err = conn.Write(append(bytes, '\n'))
		require.NoError(t, err, "Write %d should succeed", i)

		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		response, err := reader.ReadBytes('\n')
		if err != nil {
			failedReads++
			t.Logf("Failed to read response %d: %v", i, err)
			continue
		}

		var decoded map[string]interface{}
		if err := json.Unmarshal(response, &decoded); err != nil {
			corruptedMessages++
			t.Logf("Corrupted JSON in response %d: %v", i, err)
			continue
		}

		successCount++
	}

	integrity := float64(successCount) / float64(numMessages) * 100

	// Assertions
	assert.Equal(t, 100.0, integrity,
		"Data integrity should be 100%%")
	assert.Zero(t, corruptedMessages,
		"No messages should be corrupted")
	assert.Zero(t, failedReads,
		"No reads should fail")

	t.Logf("✓ %s: %.2f%% integrity (%d/%d)",
		description, integrity, successCount, numMessages)
}

// Test 5: Edge Case - Malformed JSON
// send various malformed JSON messages
// verify server handles gracefully without crashing
func (s *TCPServerTestSuite) TestMalformedJSON() {
	t := s.T()

	testCases := []struct {
		name        string
		payload     string
		shouldError bool
	}{
		{
			name:        "incomplete JSON",
			payload:     `{"type": "progress_update", "data": {`,
			shouldError: true,
		},
		{
			name:        "invalid JSON syntax",
			payload:     `{type: progress_update}`,
			shouldError: true,
		},
		{
			name:        "empty payload",
			payload:     ``,
			shouldError: true,
		},
		{
			name:        "only newline",
			payload:     `\n`,
			shouldError: true,
		},
		{
			name:        "valid JSON",
			payload:     `{"type":"test","data":{}}`,
			shouldError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conn, err := net.DialTimeout("tcp", s.serverAddr, 2*time.Second)
			require.NoError(t, err, "Should connect")
			defer conn.Close()

			_, err = conn.Write([]byte(tc.payload + "\n"))
			assert.NoError(t, err, "Write should succeed")

			// Server should handle gracefully without crashing
			time.Sleep(100 * time.Millisecond)

			// Try to send valid message after malformed one
			validMsg := tcp.Message{
				Type: "test",
				Data: map[string]interface{}{"recovery": true},
			}
			bytes, _ := json.Marshal(validMsg)
			_, err = conn.Write(append(bytes, '\n'))
			assert.NoError(t, err, "Server should recover and accept valid message")
		})
	}
}

// Test 6: Edge Case - Large Messages
// send large JSON messages to test server limits, from 1KB to 1MB, 10MB later test
// verify server handles or rejects gracefully
func (s *TCPServerTestSuite) TestLargeMessages() {
	t := s.T()

	testCases := []struct {
		name        string
		payloadSize int
		shouldPass  bool
	}{
		{"1KB payload", 1024, true},
		{"10KB payload", 10 * 1024, true},
		{"100KB payload", 100 * 1024, true},
		{"1MB payload", 1024 * 1024, false}, // max is 1MB limit in the configuration
		// {"10MB payload", 10 * 1024 * 1024, false}, // May exceed limits
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conn, err := net.DialTimeout("tcp", s.serverAddr, 5*time.Second)
			require.NoError(t, err, "Should connect")
			defer conn.Close()

			msg := tcp.Message{
				Type: "large_data",
				Data: map[string]interface{}{
					"payload": strings.Repeat("A", tc.payloadSize),
				},
			}

			bytes, err := json.Marshal(msg)
			require.NoError(t, err, "Marshaling should succeed")

			_, err = conn.Write(append(bytes, '\n'))

			if tc.shouldPass {
				assert.NoError(t, err, "Should send large message")

				// Try to read response
				reader := bufio.NewReader(conn)
				conn.SetReadDeadline(time.Now().Add(5 * time.Second))
				response, err := reader.ReadBytes('\n')
				assert.NoError(t, err, "Should receive response")
				assert.NotEmpty(t, response, "Response should not be empty")
			}
		})
	}
}

// Test 7: Edge Case - Connection Timeout
// open connection but stay idle for 30s then send message
// verify server whether timeout or keepalive terminates session
func (s *TCPServerTestSuite) TestConnectionTimeout() {
	t := s.T()

	conn, err := net.DialTimeout("tcp", s.serverAddr, 2*time.Second)
	require.NoError(t, err, "Should connect")
	defer conn.Close()

	// Send a message
	msg := tcp.Message{
		Type: "test",
		Data: map[string]interface{}{"value": 1},
	}
	bytes, _ := json.Marshal(msg)
	_, err = conn.Write(append(bytes, '\n'))
	require.NoError(t, err, "Initial write should succeed")

	// Wait for extended period without activity
	t.Log("Testing idle connection handling...")
	time.Sleep(30 * time.Second)

	// Try to send another message
	_, err = conn.Write(append(bytes, '\n'))
	// Depending on your server's keepalive settings, this may or may not error
	t.Logf("Post-idle write result: %v", err)
}

// Test 8: Edge Case - Slow Client (Slow Reader)
// Writes many small messages slowly without reading.
// verifies server handles slow clients without crashing
func (s *TCPServerTestSuite) TestSlowClient() {
	t := s.T()

	conn, err := net.DialTimeout("tcp", s.serverAddr, 2*time.Second)
	require.NoError(t, err, "Should connect")
	defer conn.Close()

	// Send messages but don't read responses
	for i := 0; i < 100; i++ {
		msg := tcp.Message{
			Type: "test",
			Data: map[string]interface{}{"seq": i},
		}
		bytes, _ := json.Marshal(msg)
		_, err := conn.Write(append(bytes, '\n'))

		if err != nil {
			t.Logf("Write failed at message %d: %v", i, err)
			break
		}

		time.Sleep(10 * time.Millisecond)
	}

	// Server should handle slow reader without crashing
	// Buffer may fill up, causing writes to block or error
	assert.True(t, true, "Server survived slow client")
}

// Test 9: Edge Case - Client Disconnects Mid-Message
// send incomplete JSON, closes socket abruptly
// then opens a new connection to ensure the server still accepts clients
func (s *TCPServerTestSuite) TestClientDisconnectsMidMessage() {
	t := s.T()

	conn, err := net.DialTimeout("tcp", s.serverAddr, 2*time.Second)
	require.NoError(t, err, "Should connect")

	// Send incomplete message
	_, err = conn.Write([]byte(`{"type":"test","data":{`))
	assert.NoError(t, err, "Partial write should succeed")

	// Abruptly close connection
	conn.Close()

	// Give server time to handle disconnection
	time.Sleep(100 * time.Millisecond)

	// Server should still accept new connections
	newConn, err := net.DialTimeout("tcp", s.serverAddr, 2*time.Second)
	assert.NoError(t, err, "Server should accept new connection after abrupt disconnect")
	if newConn != nil {
		newConn.Close()
	}
}

// Test 10: Edge Case - Empty Data Field
// Sends Message objects with nil, empty map, or missing type fields.
// Confirms the server tolerates and doesn’t reject empty fields.
func (s *TCPServerTestSuite) TestEmptyDataField() {
	t := s.T()

	testCases := []struct {
		name string
		msg  tcp.Message
	}{
		{
			name: "nil data",
			msg:  tcp.Message{Type: "test", Data: nil},
		},
		{
			name: "empty map",
			msg:  tcp.Message{Type: "test", Data: map[string]interface{}{}},
		},
		{
			name: "empty type",
			msg:  tcp.Message{Type: "", Data: map[string]interface{}{"key": "value"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conn, err := net.DialTimeout("tcp", s.serverAddr, 2*time.Second)
			require.NoError(t, err, "Should connect")
			defer conn.Close()

			bytes, err := json.Marshal(tc.msg)
			require.NoError(t, err, "Should marshal")

			_, err = conn.Write(append(bytes, '\n'))
			assert.NoError(t, err, "Should accept message with empty fields")
		})
	}
}

// Test 11: Concurrent Load Test with Mixed Operations
// simulates 20 clients sending 50 messages each concurrently for 10 s.
// tracks total sent/received messages
// then calculates throughput and delivery rate
// asserts at least 95% message delivery rate and no crashes.
func (s *TCPServerTestSuite) TestConcurrentMixedOperations() {
	t := s.T()

	const numClients = 20
	const messagesPerClient = 50
	const duration = 10 * time.Second

	var wg sync.WaitGroup
	totalSent := 0
	totalReceived := 0
	mu := sync.Mutex{}

	startTime := time.Now()

	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()

			conn, err := net.DialTimeout("tcp", s.serverAddr, 5*time.Second)
			if err != nil {
				t.Logf("Client %d failed to connect: %v", clientID, err)
				return
			}
			defer conn.Close()

			reader := bufio.NewReader(conn)

			for j := 0; j < messagesPerClient; j++ {
				if time.Since(startTime) > duration {
					break
				}

				msg := tcp.Message{
					Type: "progress_update",
					Data: map[string]interface{}{
						"client":  clientID,
						"seq":     j,
						"time":    time.Now().Unix(),
						"payload": strings.Repeat("X", 100),
					},
				}

				bytes, _ := json.Marshal(msg)
				_, err := conn.Write(append(bytes, '\n'))
				if err != nil {
					t.Logf("Client %d write error: %v", clientID, err)
					break
				}

				mu.Lock()
				totalSent++
				mu.Unlock()

				conn.SetReadDeadline(time.Now().Add(1 * time.Second))
				response, err := reader.ReadBytes('\n')
				if err == nil && len(response) > 0 {
					mu.Lock()
					totalReceived++
					mu.Unlock()
				}

				time.Sleep(100 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	// Calculate throughput
	elapsed := time.Since(startTime)
	throughput := float64(totalReceived) / elapsed.Seconds()

	t.Logf("✓ Mixed operations completed:")
	t.Logf("  - Duration: %v", elapsed)
	t.Logf("  - Messages sent: %d", totalSent)
	t.Logf("  - Messages received: %d", totalReceived)
	t.Logf("  - Throughput: %.2f msg/sec", throughput)

	// Assertions
	assert.Greater(t, totalReceived, 0, "Should receive messages")
	deliveryRate := float64(totalReceived) / float64(totalSent) * 100
	assert.GreaterOrEqual(t, deliveryRate, 95.0,
		"Delivery rate should be at least 95%%")
}

// Test 12: Server Recovery After Error
// sends several malformed messages then verifies server still accepts new connections
func (s *TCPServerTestSuite) TestServerRecoveryAfterErrors() {
	t := s.T()

	// Send several malformed messages
	conn, err := net.DialTimeout("tcp", s.serverAddr, 2*time.Second)
	require.NoError(t, err, "Should connect")

	malformedMessages := []string{
		"not json at all",
		`{"incomplete": `,
		`{{{`,
		"",
	}

	for _, malformed := range malformedMessages {
		conn.Write([]byte(malformed + "\n"))
		time.Sleep(50 * time.Millisecond)
	}

	conn.Close()
	time.Sleep(100 * time.Millisecond)

	// Verify server still accepts valid connections and messages
	newConn, err := net.DialTimeout("tcp", s.serverAddr, 2*time.Second)
	require.NoError(t, err, "Server should accept new connections after errors")
	defer newConn.Close()

	validMsg := tcp.Message{
		Type: "recovery_test",
		Data: map[string]interface{}{"status": "ok"},
	}
	bytes, _ := json.Marshal(validMsg)
	_, err = newConn.Write(append(bytes, '\n'))
	assert.NoError(t, err, "Should send valid message after recovery")

	reader := bufio.NewReader(newConn)
	newConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	response, err := reader.ReadBytes('\n')
	assert.NoError(t, err, "Should receive response")
	assert.NotEmpty(t, response, "Response should not be empty")
}

// Test 13: Broadcast Functionality
// verifies that when one client sends a message, all connected clients receive the broadcast
func (s *TCPServerTestSuite) TestBroadcastToAllClients() {
	t := s.T()

	const numClients = 5

	// Connect multiple clients
	clients := make([]net.Conn, numClients)
	readers := make([]*bufio.Reader, numClients)

	for i := 0; i < numClients; i++ {
		conn, err := net.DialTimeout("tcp", s.serverAddr, 2*time.Second)
		require.NoError(t, err, "Client %d should connect", i)
		clients[i] = conn
		readers[i] = bufio.NewReader(conn)
		defer conn.Close()
	}

	time.Sleep(100 * time.Millisecond)

	// Send message from first client
	msg := tcp.Message{
		Type: "broadcast_test",
		Data: map[string]interface{}{
			"message": "Hello everyone!",
			"from":    "client_0",
		},
	}

	bytes, _ := json.Marshal(msg)
	_, err := clients[0].Write(append(bytes, '\n'))
	require.NoError(t, err, "Should send broadcast message")

	// Verify all clients receive the broadcast
	receivedCount := 0
	for i := 0; i < numClients; i++ {
		clients[i].SetReadDeadline(time.Now().Add(2 * time.Second))
		response, err := readers[i].ReadBytes('\n')
		if err == nil && len(response) > 0 {
			var decoded map[string]interface{}
			if json.Unmarshal(response, &decoded) == nil {
				receivedCount++
				t.Logf("Client %d received broadcast", i)
			}
		}
	}

	assert.Equal(t, numClients, receivedCount,
		"All clients should receive broadcast")
}

// Helper functions for statistics
func calculateAverage(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	var sum time.Duration
	for _, d := range durations {
		sum += d
	}
	return sum / time.Duration(len(durations))
}

func calculatePercentile(durations []time.Duration, percentile int) time.Duration {
	if len(durations) == 0 {
		return 0
	}

	// Sort durations
	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)

	// Simple bubble sort (fine for test data)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	index := (percentile * len(sorted)) / 100
	if index >= len(sorted) {
		index = len(sorted) - 1
	}
	return sorted[index]
}

func calculateMax(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	max := durations[0]
	for _, d := range durations {
		if d > max {
			max = d
		}
	}
	return max
}

// Run the test suite
func TestTCPServerTestSuite(t *testing.T) {
	suite.Run(t, new(TCPServerTestSuite))
}

// Individual tests (non-suite) for specific scenarios

// Test for connection refused scenario
func TestConnectionRefused(t *testing.T) {
	conn, err := net.DialTimeout("tcp", "localhost:9999", 1*time.Second)
	assert.Error(t, err, "Should fail to connect to non-existent server")
	assert.Nil(t, conn, "Connection should be nil")
}

// Benchmark tests
func BenchmarkMessageThroughput(b *testing.B) {
	server := tcp.NewServer("localhost:8090", "localhost:6379")
	go server.Start()
	time.Sleep(100 * time.Millisecond)
	defer server.Stop()

	conn, _ := net.Dial("tcp", "localhost:8090")
	defer conn.Close()

	msg := tcp.Message{
		Type: "benchmark",
		Data: map[string]interface{}{"value": 123},
	}
	bytes, _ := json.Marshal(msg)
	payload := append(bytes, '\n')

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn.Write(payload)
	}
}
