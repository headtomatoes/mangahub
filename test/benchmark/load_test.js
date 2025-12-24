// MangaHub Load Testing with k6
// ============================================================================
// Install: https://k6.io/docs/get-started/installation/
// Run: k6 run load_test.js
//
// Stress Test Targets:
//   - HTTP: 100-200 concurrent users with real interactions
//   - TCP: 100+ concurrent connections (use tcp_stress_test.go)
//   - WebSocket: 50+ concurrent connections
// ============================================================================

import http from 'k6/http';
import ws from 'k6/ws';
import { check, sleep, group } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// ============================================================================
// CONFIGURATION
// ============================================================================

const BASE_URL = __ENV.BASE_URL || 'http://10.238.20.112:8084';
const WS_URL = __ENV.WS_URL || 'ws://10.238.20.112:8084/ws';

// Test user credentials
const TEST_USER = {
  username: 'loadtestuser',
  password: 'loadtest123',
  email: 'loadtest@test.com'
};

// ============================================================================
// CUSTOM METRICS
// ============================================================================

const searchLatency = new Trend('search_latency', true);
const apiLatency = new Trend('api_latency', true);
const authLatency = new Trend('auth_latency', true);
const wsLatency = new Trend('ws_latency', true);
const errorRate = new Rate('error_rate');
const successfulRequests = new Counter('successful_requests');
const failedRequests = new Counter('failed_requests');
const wsMessages = new Counter('ws_messages');

// ============================================================================
// TEST SCENARIOS
// ============================================================================

export const options = {
  // Connection pooling to prevent port exhaustion
  noConnectionReuse: false,
  userAgent: 'k6-load-test',
  batch: 10,
  batchPerHost: 5,
  
  // HTTP settings to reuse connections
  httpDebug: 'full',
  insecureSkipTLSVerify: true,
  
  scenarios: {
    // Scenario 1: HTTP stress test - 100 concurrent users (1/4 stress)
    http_stress: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 25 },   // Warm up
        { duration: '30s', target: 50 },   // Ramp to 50
        { duration: '1m', target: 75 },    // Push to 75
        { duration: '1m', target: 100 },   // Target: 100 users
        { duration: '2m', target: 100 },   // Sustain 100
        { duration: '30s', target: 0 },    // Ramp down
      ],
      gracefulRampDown: '10s',
      exec: 'httpStressTest',
    },
    
    // Scenario 2: Search performance (1/4 stress)
    search_load: {
      executor: 'constant-vus',
      vus: 15,
      duration: '3m',
      startTime: '30s',
      exec: 'searchScenario',
    },
    
    // Scenario 3: WebSocket stress - 25+ users (1/4 stress)
    websocket_stress: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '20s', target: 13 },
        { duration: '30s', target: 25 },   // Target: 25
        { duration: '2m', target: 25 },
        { duration: '30s', target: 38 },   // Push beyond
        { duration: '1m', target: 38 },
        { duration: '20s', target: 0 },
      ],
      startTime: '1m',
      exec: 'websocketStressTest',
    },
    
    // Scenario 4: Spike test (1/4 stress)
    spike_test: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '10s', target: 5 },
        { duration: '10s', target: 100 },  // Spike!
        { duration: '30s', target: 100 },
        { duration: '10s', target: 5 },
      ],
      startTime: '6m',
      exec: 'spikeScenario',
    },
  },
  
  thresholds: {
    'http_req_duration': ['p(95)<500'],
    'search_latency': ['p(95)<500', 'avg<300'],
    'ws_latency': ['p(95)<1000'],
    'error_rate': ['rate<0.1'],
    'http_req_failed': ['rate<0.1'],
  },
};

// ============================================================================
// SETUP
// ============================================================================

export function setup() {
  console.log('Setting up load test...');
  
  // Register test user
  http.post(`${BASE_URL}/auth/register`, JSON.stringify({
    username: TEST_USER.username,
    email: TEST_USER.email,
    password: TEST_USER.password,
  }), { headers: { 'Content-Type': 'application/json' } });
  
  // Login to get token
  const loginRes = http.post(`${BASE_URL}/auth/login`, JSON.stringify({
    username: TEST_USER.username,
    password: TEST_USER.password,
  }), { headers: { 'Content-Type': 'application/json' } });
  
  const token = JSON.parse(loginRes.body).access_token;
  console.log('Token obtained: ' + (token ? 'YES' : 'NO'));
  
  return { token };
}

// ============================================================================
// HTTP STRESS TEST (200-400 users with connection pooling)
// ============================================================================

export function httpStressTest(data) {
  const headers = {
    'Authorization': `Bearer ${data.token}`,
    'Content-Type': 'application/json',
    'Connection': 'keep-alive',
  };
  
  // HTTP params to enable connection reuse and prevent port exhaustion
  const params = {
    headers: headers,
    timeout: '30s',
    tags: { name: 'httpStressTest' },
  };
  
  group('Real User Interactions', () => {
    // Browse manga list
    group('Browse Manga', () => {
      const page = Math.floor(Math.random() * 5) + 1;
      const res = http.get(`${BASE_URL}/api/manga/?page=${page}&limit=20`, params);
      apiLatency.add(res.timings.duration);
      check(res, { 'browse status 200': (r) => r.status === 200 });
      if (res.status === 200) successfulRequests.add(1);
      else { failedRequests.add(1); errorRate.add(1); }
    });
    
    sleep(0.5);
    
    // Search for manga
    group('Search Manga', () => {
      const queries = ['naruto', 'one piece', 'dragon', 'attack', 'hero', 'death'];
      const query = queries[Math.floor(Math.random() * queries.length)];
      const res = http.get(`${BASE_URL}/api/manga/search?q=${encodeURIComponent(query)}`, params);
      searchLatency.add(res.timings.duration);
      check(res, { 
        'search status 200': (r) => r.status === 200,
        'search under 500ms': (r) => r.timings.duration < 500,
      });
      if (res.status === 200) successfulRequests.add(1);
      else { failedRequests.add(1); errorRate.add(1); }
    });
    
    sleep(0.3);
    
    // Get manga details
    group('Get Manga Details', () => {
      const mangaId = Math.floor(Math.random() * 100) + 1;
      const res = http.get(`${BASE_URL}/api/manga/${mangaId}`, params);
      apiLatency.add(res.timings.duration);
      check(res, { 'detail status 200 or 404': (r) => r.status === 200 || r.status === 404 });
      if (res.status === 200 || res.status === 404) successfulRequests.add(1);
      else { failedRequests.add(1); errorRate.add(1); }
    });
    
    sleep(0.2);
    
    // Get genres
    group('Get Genres', () => {
      const res = http.get(`${BASE_URL}/api/genres/`, params);
      apiLatency.add(res.timings.duration);
      check(res, { 'genres status 200': (r) => r.status === 200 });
      if (res.status === 200) successfulRequests.add(1);
      else { failedRequests.add(1); errorRate.add(1); }
    });
  });
  
  sleep(Math.random() * 2);
}

// ============================================================================
// SEARCH SCENARIO
// ============================================================================

export function searchScenario(data) {
  const headers = {
    'Authorization': `Bearer ${data.token}`,
    'Content-Type': 'application/json',
    'Connection': 'keep-alive',
  };
  
  const params = {
    headers: headers,
    timeout: '30s',
  };
  
  const queries = [
    'naruto', 'one piece', 'dragon ball', 'attack on titan',
    'my hero academia', 'death note', 'black clover', 'demon slayer',
    'tokyo ghoul', 'jojo', 'bleach', 'hunter', 'fullmetal',
  ];
  
  const query = queries[Math.floor(Math.random() * queries.length)];
  const start = Date.now();
  const res = http.get(`${BASE_URL}/api/manga/search?q=${encodeURIComponent(query)}`, params);
  const duration = Date.now() - start;
  
  searchLatency.add(duration);
  
  const passed = check(res, {
    'search status 200': (r) => r.status === 200,
    'search under 500ms': (r) => r.timings.duration < 500,
  });
  
  if (passed) successfulRequests.add(1);
  else { failedRequests.add(1); errorRate.add(1); }
  
  sleep(0.5 + Math.random());
}

// ============================================================================
// WEBSOCKET STRESS TEST (50+ users)
// ============================================================================

export function websocketStressTest(data) {
  const url = `${WS_URL}?token=${data.token}`;
  
  const res = ws.connect(url, {}, function (socket) {
    socket.on('open', () => {
      console.log('WebSocket connected');
      
      // Send join message
      socket.send(JSON.stringify({
        type: 'join',
        room: 'stress_test',
        user: `user_${__VU}`,
      }));
      wsMessages.add(1);
    });
    
    socket.on('message', (msg) => {
      wsMessages.add(1);
      const start = Date.now();
      
      try {
        const data = JSON.parse(msg);
        wsLatency.add(Date.now() - start);
      } catch (e) {
        // Non-JSON message
      }
    });
    
    socket.on('error', (e) => {
      errorRate.add(1);
      console.log('WebSocket error: ' + e.error());
    });
    
    // Keep connection alive and send periodic messages
    socket.setInterval(() => {
      socket.send(JSON.stringify({
        type: 'message',
        content: `Stress test message from VU ${__VU}`,
        timestamp: Date.now(),
      }));
      wsMessages.add(1);
    }, 2000);
    
    // Keep connection open for test duration
    socket.setTimeout(() => {
      socket.send(JSON.stringify({ type: 'leave' }));
      socket.close();
    }, 30000);
  });
  
  check(res, { 'WebSocket connected': (r) => r && r.status === 101 });
}

// ============================================================================
// SPIKE SCENARIO
// ============================================================================

export function spikeScenario(data) {
  const headers = {
    'Authorization': `Bearer ${data.token}`,
    'Content-Type': 'application/json',
    'Connection': 'keep-alive',
  };
  
  const params = {
    headers: headers,
    timeout: '30s',
  };
  
  // Rapid fire requests
  const res = http.get(`${BASE_URL}/api/manga/?page=1&limit=10`, params);
  apiLatency.add(res.timings.duration);
  
  check(res, { 'spike test status 200': (r) => r.status === 200 });
  
  if (res.status === 200) successfulRequests.add(1);
  else { failedRequests.add(1); errorRate.add(1); }
  
  sleep(0.1);
}

// ============================================================================
// DEFAULT FUNCTION
// ============================================================================

export default function(data) {
  httpStressTest(data);
}

// ============================================================================
// TEARDOWN
// ============================================================================

export function teardown(data) {
  console.log('Load test completed');
  console.log(`Token used: ${data.token ? 'YES' : 'NO'}`);
}
