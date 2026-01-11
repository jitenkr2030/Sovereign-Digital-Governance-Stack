import http from 'k6/http';
import { check, sleep } from 'k6/metrics';
import { Trend, Counter, Gauge } from 'k6/metrics';
import { randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

// Soak test specific metrics
const memoryUsage = new Trend('memory_usage');
const cpuUsage = new Trend('cpu_usage');
const dbConnectionPool = new Trend('db_connection_pool');
const activeSessions = new Trend('active_sessions');
const leakDetection = new Counter('memory_leak_detected');
const errorAccumulation = new Counter('errors_over_time');

// Soak test configuration
export const options = {
  // 8-hour soak test
  long_soak: {
    stages: [
      { duration: '5m', target: 20 },    // Ramp up
      { duration: '8h', target: 20 },    // Sustained load
      { duration: '5m', target: 0 },     // Ramp down
    ],
    thresholds: {
      http_req_duration: ['p(95)<1500'],
      memory_usage: ['avg<80'],  // Memory utilization percentage
      cpu_usage: ['avg<70'],     // CPU utilization percentage
      leakDetection: ['count==0'],
      errorAccumulation: ['count<100'],
    },
  },
  
  // 24-hour endurance test
  endurance: {
    stages: [
      { duration: '10m', target: 30 },
      { duration: '24h', target: 30 },
      { duration: '5m', target: 0 },
    ],
    thresholds: {
      http_req_duration: ['p(95)<2000'],
      memory_usage: ['avg<85'],
      leakDetection: ['count==0'],
      errorAccumulation: ['count<500'],
    },
  },
  
  // Weekend simulation
  weekend_sim: {
    stages: [
      { duration: '30m', target: 10 },   // Light weekend load
      { duration: '48h', target: 10 },
      { duration: '10m', target: 0 },
    ],
    thresholds: {
      http_req_duration: ['p(95)<1000'],
      memory_usage: ['avg<75'],
      leakDetection: ['count==0'],
    },
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:3000';
const API_BASE_URL = __ENV.API_BASE_URL || 'http://localhost:3000/api/v1';

// Track metrics over time
let lastMemoryUsage = 0;
let consecutiveHighMemory = 0;

export function setup() {
  // Get initial authentication
  const loginRes = http.post(
    `${API_BASE_URL}/auth/login`,
    JSON.stringify({
      email: __ENV.TEST_ADMIN_EMAIL || 'admin@test.gov',
      password: __ENV.TEST_ADMIN_PASSWORD || 'test_password123',
    }),
    { headers: { 'Content-Type': 'application/json' } }
  );
  
  return {
    token: loginRes.status === 200 ? JSON.parse(loginRes.body).token : '',
  };
}

// Soak test - sustained load over extended period
export default function (data) {
  // Simulate realistic usage patterns
  const usagePatterns = [
    { weight: 40, action: 'dashboard_view' },
    { weight: 25, action: 'intelligence_review' },
    { weight: 20, action: 'intervention_tracking' },
    { weight: 10, action: 'report_generation' },
    { weight: 5, action: 'system_check' },
  ];
  
  // Select usage pattern
  const totalWeight = usagePatterns.reduce((sum, p) => sum + p.weight, 0);
  let randomWeight = randomIntBetween(1, totalWeight);
  let selectedPattern = usagePatterns[0];
  
  for (const pattern of usagePatterns) {
    randomWeight -= pattern.weight;
    if (randomWeight <= 0) {
      selectedPattern = pattern;
      break;
    }
  }
  
  // Execute the selected pattern
  switch (selectedPattern.action) {
    case 'dashboard_view':
      soakTestDashboard(data.token);
      break;
    case 'intelligence_review':
      soakTestIntelligence(data.token);
      break;
    case 'intervention_tracking':
      soakTestInterventions(data.token);
      break;
    case 'report_generation':
      soakTestReports(data.token);
      break;
    case 'system_check':
      soakTestSystemHealth(data.token);
      break;
  }
  
  // Simulate user think time
  sleep(randomIntBetween(5, 15));
}

function soakTestDashboard(token) {
  const startTime = Date.now();
  
  const res = http.get(`${BASE_URL}/dashboard`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  
  const duration = Date.now() - startTime;
  
  check(res, {
    'dashboard loads': (r) => r.status === 200,
    'dashboard renders within time': (r) => r.timings.duration < 2000,
  });
  
  // Fetch additional dashboard metrics
  const metricsRes = http.get(`${API_BASE_URL}/dashboard/metrics`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  
  if (metricsRes.status === 200) {
    const metrics = JSON.parse(metricsRes.body);
    // Track resource utilization
    if (metrics.memory) {
      memoryUsage.add(metrics.memory.usedPercent);
      detectMemoryLeak(metrics.memory.usedPercent);
    }
    if (metrics.cpu) {
      cpuUsage.add(metrics.cpu.utilization);
    }
    if (metrics.database) {
      dbConnectionPool.add(metrics.database.activeConnections);
    }
  }
}

function soakTestIntelligence(token) {
  // GET list
  const listRes = http.get(`${API_BASE_URL}/intelligence`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  
  check(listRes, {
    'intelligence list loads': (r) => r.status === 200,
  });
  
  // Occasionally update an item (5% chance)
  if (randomIntBetween(1, 100) <= 5 && listRes.status === 200) {
    const items = JSON.parse(listRes.body).items;
    if (items && items.length > 0) {
      const item = items[randomIntBetween(0, items.length - 1)];
      
      const updateRes = http.patch(
        `${API_BASE_URL}/intelligence/${item.id}`,
        JSON.stringify({
          status: 'under_review',
          notes: `Soak test update at ${new Date().toISOString()}`,
        }),
        {
          headers: {
            Authorization: `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
        }
      );
      
      check(updateRes, {
        'intelligence update succeeds': (r) => r.status === 200,
      });
    }
  }
}

function soakTestInterventions(token) {
  // GET interventions
  const listRes = http.get(`${API_BASE_URL}/interventions`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  
  check(listRes, {
    'interventions load': (r) => r.status === 200,
  });
  
  // GET statistics
  const statsRes = http.get(`${API_BASE_URL}/interventions/statistics`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  
  if (statsRes.status === 200) {
    activeSessions.add(JSON.parse(statsRes.body).activeInterventions || 0);
  }
}

function soakTestReports(token) {
  // GET reports list
  const listRes = http.get(`${API_BASE_URL}/reports`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  
  check(listRes, {
    'reports load': (r) => r.status === 200,
  });
  
  // GET templates (lightweight operation)
  const templatesRes = http.get(`${API_BASE_URL}/reports/templates`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  
  check(templatesRes, {
    'templates load': (r) => r.status === 200,
  });
}

function soakTestSystemHealth(token) {
  // GET system health
  const healthRes = http.get(`${API_BASE_URL}/health`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  
  if (healthRes.status === 200) {
    const health = JSON.parse(healthRes.body);
    
    check(health, {
      'system healthy': (h) => h.status === 'healthy',
      'database connected': (h) => h.database?.connected === true,
      'cache available': (h) => h.cache?.available === true,
    });
  }
  
  // GET metrics endpoint
  const metricsRes = http.get(`${API_BASE_URL}/metrics`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  
  if (metricsRes.status === 200) {
    const metrics = JSON.parse(metricsRes.body);
    
    // Check for resource warnings
    if (metrics.memory?.usedPercent > 85) {
      errorAccumulation.add(1, { type: 'memory_warning' });
    }
    if (metrics.cpu?.utilization > 80) {
      errorAccumulation.add(1, { type: 'cpu_warning' });
    }
  }
}

function detectMemoryLeak(currentUsage) {
  // Detect gradual memory increase over time
  if (currentUsage > lastMemoryUsage + 5) {
    consecutiveHighMemory++;
  } else if (currentUsage < lastMemoryUsage - 5) {
    consecutiveHighMemory = 0;
  }
  
  lastMemoryUsage = currentUsage;
  
  // If memory consistently increases over 10 checks, flag as potential leak
  if (consecutiveHighMemory > 10) {
    leakDetection.add(1);
    consecutiveHighMemory = 0; // Reset to avoid counting same issue multiple times
  }
}

// Custom summary for soak test results
export function handleSummary(data) {
  const memoryAvg = data.metrics.memory_usage?.values?.avg || 0;
  const cpuAvg = data.metrics.cpu_usage?.values?.avg || 0;
  const leaks = data.metrics.memory_leak_detected?.values?.count || 0;
  const errors = data.metrics.errors_over_time?.values?.count || 0;
  
  return {
    stdout: `
========================================
SOAK TEST RESULTS
========================================
Test Duration: ${(data.state.testRunDurationMs / 1000 / 60).toFixed(2)} minutes
Total Iterations: ${data.metrics.iterations?.values?.count || 0}

RESOURCE UTILIZATION:
--------------------
Average Memory Usage: ${memoryAvg.toFixed(2)}%
Average CPU Usage: ${cpuAvg.toFixed(2)}%
Memory Leaks Detected: ${leaks}
Resource Warnings: ${errors}

PERFORMANCE METRICS:
-------------------
Total Requests: ${data.metrics.http_reqs?.values?.count || 0}
Requests/sec: ${data.metrics.http_reqs?.values?.rate || 0}
Avg Response Time: ${data.metrics.http_req_duration?.values?.avg || 0}ms
P95 Response Time: ${data.metrics.http_req_duration?.values?.p95 || 0}ms
P99 Response Time: ${data.metrics.http_req_duration?.values?.p99 || 0}ms

ERRORS:
-------
Total Errors: ${data.metrics.http_req_failed?.values?.count || 0}
Error Rate: ${((data.metrics.http_req_failed?.values?.rate || 0) * 100).toFixed(2)}%

SYSTEM STATUS: ${leaks === 0 && errors < 100 ? 'PASSED' : 'NEEDS REVIEW'}
========================================
    `,
    'soak-test-results.json': JSON.stringify({
      summary: {
        duration: data.state.testRunDurationMs,
        iterations: data.metrics.iterations?.values?.count,
        resourceUtilization: {
          memory: { avg: memoryAvg },
          cpu: { avg: cpuAvg },
        },
        performance: {
          totalRequests: data.metrics.http_reqs?.values?.count,
          avgResponseTime: data.metrics.http_req_duration?.values?.avg,
          p95ResponseTime: data.metrics.http_req_duration?.values?.p95,
          errorRate: data.metrics.http_req_failed?.values?.rate,
        },
        issues: {
          memoryLeaks: leaks,
          resourceWarnings: errors,
        },
      },
    }, null, 2),
  };
}
