import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';
import { randomString, randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

// Custom metrics for performance tracking
const errorRate = new Rate('errors');
const responseTime = new Trend('response_time');
const pageLoadTime = new Trend('page_load_time');
const apiLatency = new Trend('api_latency');
const requestsPerSecond = new Counter('requests_per_second');

// Test configuration
export const options = {
  // Smoke test - quick sanity check
  smoke: {
    stages: [
      { duration: '10s', target: 1 },
      { duration: '10s', target: 1 },
      { duration: '5s', target: 0 },
    ],
    thresholds: {
      http_req_duration: ['p(95)<500'],
      errors: ['rate<0.01'],
    },
  },
  
  // Load test - normal expected load
  load: {
    stages: [
      { duration: '1m', target: 10 },   // Ramp up to 10 users
      { duration: '5m', target: 10 },   // Stay at 10 users
      { duration: '1m', target: 25 },   // Ramp up to 25 users
      { duration: '5m', target: 25 },   // Stay at 25 users
      { duration: '1m', target: 50 },   // Ramp up to 50 users
      { duration: '5m', target: 50 },   // Stay at 50 users
      { duration: '1m', target: 0 },    // Ramp down
    ],
    thresholds: {
      http_req_duration: ['p(95)<1000', 'p(99)<2000'],
      errors: ['rate<0.05'],
      response_time: ['avg<500'],
    },
  },
  
  // Stress test - push beyond normal capacity
  stress: {
    stages: [
      { duration: '2m', target: 20 },   // Warm up
      { duration: '3m', target: 50 },   // Normal load
      { duration: '5m', target: 100 },  // High load
      { duration: '5m', target: 200 },  // Stress test
      { duration: '5m', target: 300 },  // Peak load
      { duration: '2m', target: 0 },    // Cool down
    ],
    thresholds: {
      http_req_duration: ['p(95)<2000'],
      errors: ['rate<0.1'],
    },
  },
  
  // Spike test - sudden surge
  spike: {
    stages: [
      { duration: '30s', target: 10 },
      { duration: '10s', target: 100 },  // Sudden spike
      { duration: '2m', target: 100 },   // Hold spike
      { duration: '30s', target: 10 },
      { duration: '30s', target: 0 },
    ],
    thresholds: {
      http_req_duration: ['p(95)<3000'],
      errors: ['rate<0.15'],
    },
  },
  
  // Soak test - long duration endurance
  soak: {
    stages: [
      { duration: '2m', target: 30 },
      { duration: '8h', target: 30 },    // Long duration test
      { duration: '2m', target: 0 },
    ],
    thresholds: {
      http_req_duration: ['p(95)<1500'],
      errors: ['rate<0.05'],
      response_time: ['avg<800'],
    },
  },
};

// Base URL from environment or default
const BASE_URL = __ENV.BASE_URL || 'http://localhost:3000';
const API_BASE_URL = __ENV.API_BASE_URL || 'http://localhost:3000/api/v1';

// Authentication tokens storage
let authTokens = {
  admin: '',
  analyst: '',
  supervisor: '',
};

// Test data
const testUsers = {
  admin: {
    email: __ENV.TEST_ADMIN_EMAIL || 'admin@test.gov',
    password: __ENV.TEST_ADMIN_PASSWORD || 'test_password123',
  },
  analyst: {
    email: __ENV.TEST_ANALYST_EMAIL || 'analyst@test.gov',
    password: __ENV.TEST_ANALYST_PASSWORD || 'test_password123',
  },
  supervisor: {
    email: __ENV.TEST_SUPERVISOR_EMAIL || 'supervisor@test.gov',
    password: __ENV.TEST_SUPERVISOR_PASSWORD || 'test_password123',
  },
};

// Setup function - runs once per virtual user
export function setup() {
  // Authenticate and get tokens for different roles
  const tokens = {};
  
  for (const [role, credentials] of Object.entries(testUsers)) {
    const loginRes = http.post(
      `${API_BASE_URL}/auth/login`,
      JSON.stringify({
        email: credentials.email,
        password: credentials.password,
      }),
      {
        headers: { 'Content-Type': 'application/json' },
      }
    );
    
    if (loginRes.status === 200) {
      tokens[role] = JSON.parse(loginRes.body).token;
    }
  }
  
  return { tokens };
}

// Teardown function - runs once after test completes
export function teardown(data) {
  // Logout all users
  for (const token of Object.values(data.tokens)) {
    if (token) {
      http.post(
        `${API_BASE_URL}/auth/logout`,
        {},
        {
          headers: { Authorization: `Bearer ${token}` },
        }
      );
    }
  }
}

// Main test function - runs for each virtual user iteration
export default function (data) {
  // Track request count
  requestsPerSecond.add(1);
  
  // Select a random user role for this iteration
  const roles = Object.keys(data.tokens);
  const selectedRole = roles[Math.floor(Math.random() * roles.length)];
  const token = data.tokens[selectedRole];
  
  // Execute different test scenarios
  const scenarios = [
    () => testDashboardPage(token),
    () => testIntelligenceAPI(token),
    () => testInterventionsAPI(token),
    () => testCrisisAlertsAPI(token),
    () => testReportsAPI(token),
  ];
  
  // Execute random scenario
  const scenarioIndex = randomIntBetween(0, scenarios.length - 1);
  scenarios[scenarioIndex]();
  
  // Random delay between requests
  sleep(randomIntBetween(1, 5));
}

// Test Dashboard Page Load
function testDashboardPage(token) {
  const startTime = Date.now();
  
  const res = http.get(`${BASE_URL}/dashboard`, {
    headers: {
      Authorization: `Bearer ${token}`,
      Accept: 'text/html',
    },
  });
  
  const duration = Date.now() - startTime;
  pageLoadTime.add(duration);
  responseTime.add(res.timings.duration);
  
  const success = check(res, {
    'dashboard loads successfully': (r) => r.status === 200,
    'dashboard contains stats cards': (r) => r.body.includes('stats-card') || r.body.includes('statistics'),
    'dashboard contains navigation': (r) => r.body.includes('navigation') || r.body.includes('sidebar'),
    'response has no errors': (r) => r.status < 400,
  });
  
  errorRate.add(!success);
}

// Test Intelligence API endpoints
function testIntelligenceAPI(token) {
  const startTime = Date.now();
  
  // GET all intelligence items
  const listRes = http.get(`${API_BASE_URL}/intelligence`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  
  const listDuration = Date.now() - startTime;
  apiLatency.add(listDuration);
  
  const listSuccess = check(listRes, {
    'intelligence list returns 200': (r) => r.status === 200,
    'intelligence list has items': (r) => JSON.parse(r.body).items?.length >= 0,
    'intelligence list has pagination': (r) => JSON.parse(r.body).total !== undefined,
  });
  
  errorRate.add(!listSuccess);
  
  // GET single intelligence item (if items exist)
  if (listSuccess && JSON.parse(listRes.body).items?.length > 0) {
    const items = JSON.parse(listRes.body).items;
    const randomItem = items[randomIntBetween(0, items.length - 1)];
    
    sleep(randomIntBetween(1, 2));
    
    const itemStartTime = Date.now();
    const itemRes = http.get(`${API_BASE_URL}/intelligence/${randomItem.id}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    
    const itemDuration = Date.now() - itemStartTime;
    apiLatency.add(itemDuration);
    
    check(itemRes, {
      'single intelligence item returns 200': (r) => r.status === 200,
      'intelligence item has required fields': (r) => {
        const item = JSON.parse(r.body);
        return item.id && item.title && item.status;
      },
    });
  }
  
  // POST new intelligence item (10% chance to avoid clutter)
  if (randomIntBetween(1, 100) <= 10) {
    sleep(randomIntBetween(1, 2));
    
    const createStartTime = Date.now();
    const createRes = http.post(
      `${API_BASE_URL}/intelligence`,
      JSON.stringify({
        title: `Test Intelligence ${randomString(8)}`,
        description: 'Created during load test',
        severity: randomIntBetween(1, 4),
        source_id: 'source-001',
        category_id: 'threat-medium-001',
      }),
      {
        headers: {
          Authorization: `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      }
    );
    
    const createDuration = Date.now() - createStartTime;
    apiLatency.add(createDuration);
    
    check(createRes, {
      'create intelligence returns 201': (r) => r.status === 201,
      'created intelligence has ID': (r) => JSON.parse(r.body).id !== undefined,
    });
  }
}

// Test Interventions API endpoints
function testInterventionsAPI(token) {
  const startTime = Date.now();
  
  // GET all interventions
  const listRes = http.get(`${API_BASE_URL}/interventions`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  
  const listDuration = Date.now() - startTime;
  apiLatency.add(listDuration);
  
  const listSuccess = check(listRes, {
    'interventions list returns 200': (r) => r.status === 200,
    'interventions list has items': (r) => JSON.parse(r.body).items?.length >= 0,
  });
  
  errorRate.add(!listSuccess);
  
  // Filter by status
  sleep(randomIntBetween(1, 2));
  
  const filterStartTime = Date.now();
  const filterRes = http.get(`${API_BASE_URL}/interventions?status=pending`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  
  const filterDuration = Date.now() - filterStartTime;
  apiLatency.add(filterDuration);
  
  check(filterRes, {
    'filtered interventions returns 200': (r) => r.status === 200,
  });
  
  // GET intervention statistics
  sleep(randomIntBetween(1, 2));
  
  const statsStartTime = Date.now();
  const statsRes = http.get(`${API_BASE_URL}/interventions/statistics`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  
  const statsDuration = Date.now() - statsStartTime;
  apiLatency.add(statsDuration);
  
  check(statsRes, {
    'intervention statistics returns 200': (r) => r.status === 200,
    'statistics has required metrics': (r) => {
      const stats = JSON.parse(r.body);
      return stats.total !== undefined && stats.byStatus !== undefined;
    },
  });
}

// Test Crisis Alerts API endpoints
function testCrisisAlertsAPI(token) {
  const startTime = Date.now();
  
  // GET all active alerts
  const alertsRes = http.get(`${API_BASE_URL}/crisis/alerts?status=active`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  
  const alertsDuration = Date.now() - startTime;
  apiLatency.add(alertsDuration);
  
  check(alertsRes, {
    'crisis alerts returns 200': (r) => r.status === 200,
    'alerts has count': (r) => JSON.parse(r.body).count !== undefined,
  });
  
  // GET alert feed
  sleep(randomIntBetween(1, 2));
  
  const feedStartTime = Date.now();
  const feedRes = http.get(`${API_BASE_URL}/crisis/feed`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  
  const feedDuration = Date.now() - feedStartTime;
  apiLatency.add(feedDuration);
  
  check(feedRes, {
    'crisis feed returns 200': (r) => r.status === 200,
  });
  
  // GET console status
  sleep(randomIntBetween(1, 2));
  
  const statusStartTime = Date.now();
  const statusRes = http.get(`${API_BASE_URL}/crisis/status`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  
  const statusDuration = Date.now() - statusStartTime;
  apiLatency.add(statusDuration);
  
  check(statusRes, {
    'crisis status returns 200': (r) => r.status === 200,
    'status has active alerts count': (r) => {
      const status = JSON.parse(r.body);
      return status.activeAlerts !== undefined;
    },
  });
}

// Test Reports API endpoints
function testReportsAPI(token) {
  // GET saved reports
  const startTime = Date.now();
  
  const reportsRes = http.get(`${API_BASE_URL}/reports`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  
  const reportsDuration = Date.now() - startTime;
  apiLatency.add(reportsDuration);
  
  const reportsSuccess = check(reportsRes, {
    'reports list returns 200': (r) => r.status === 200,
    'reports list has items': (r) => JSON.parse(r.body).reports?.length >= 0,
  });
  
  errorRate.add(!reportsSuccess);
  
  // GET report templates
  sleep(randomIntBetween(1, 2));
  
  const templatesStartTime = Date.now();
  const templatesRes = http.get(`${API_BASE_URL}/reports/templates`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  
  const templatesDuration = Date.now() - templatesStartTime;
  apiLatency.add(templatesDuration);
  
  check(templatesRes, {
    'report templates returns 200': (r) => r.status === 200,
    'templates has available options': (r) => {
      const templates = JSON.parse(r.body);
      return templates.length >= 0;
    },
  });
  
  // Generate report (5% chance)
  if (randomIntBetween(1, 100) <= 5) {
    sleep(randomIntBetween(1, 3));
    
    const generateStartTime = Date.now();
    const generateRes = http.post(
      `${API_BASE_URL}/reports/generate`,
      JSON.stringify({
        template: 'weekly_intelligence',
        dateRange: {
          start: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
          end: new Date().toISOString(),
        },
      }),
      {
        headers: {
          Authorization: `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      }
    );
    
    const generateDuration = Date.now() - generateStartTime;
    apiLatency.add(generateDuration);
    
    check(generateRes, {
      'report generation returns 202': (r) => r.status === 202 || r.status === 200,
      'generation has job ID': (r) => JSON.parse(r.body).jobId !== undefined,
    });
  }
}

// Handle special thresholds for soak tests
export function handleSummary(data) {
  return {
    stdout: `Performance Test Summary:
========================
Duration: ${data.metrics.iterations_duration?.values?.avg || 'N/A'}s average iteration duration
Total Requests: ${data.metrics.http_reqs?.values?.count || 0}
Requests/sec: ${data.metrics.http_reqs?.values?.rate || 0}
Avg Response Time: ${data.metrics.http_req_duration?.values?.avg || 0}ms
P95 Response Time: ${data.metrics.http_req_duration?.values?.p95 || 0}ms
P99 Response Time: ${data.metrics.http_req_duration?.values?.p99 || 0}ms
Error Rate: ${(data.metrics.errors?.values?.rate || 0) * 100}%
Errors: ${data.metrics.errors?.values?.count || 0}
========================
    `,
    'performance-summary.json': JSON.stringify(data, null, 2),
  };
}
