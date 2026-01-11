import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';
import { randomIntBetween } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

// Spike test specific metrics
const spikeErrorRate = new Rate('spike_errors');
const spikeResponseTime = new Trend('spike_response_time');
const concurrentConnections = new Trend('concurrent_connections');

// Spike test configuration
export const options = {
  // Extreme spike test
  extreme_spike: {
    stages: [
      { duration: '10s', target: 5 },      // Baseline
      { duration: '5s', target: 5 },       // Hold baseline
      { duration: '2s', target: 200 },     // Extreme spike
      { duration: '30s', target: 200 },    // Sustain spike
      { duration: '5s', target: 5 },       // Drop to baseline
      { duration: '10s', target: 0 },      // Cool down
    ],
    thresholds: {
      http_req_duration: ['p(95)<5000'],
      spike_errors: ['rate<0.2'],
      spike_response_time: ['avg<2000'],
    },
  },
  
  // Step spike test
  step_spike: {
    stages: [
      { duration: '30s', target: 10 },
      { duration: '30s', target: 50 },
      { duration: '30s', target: 100 },
      { duration: '30s', target: 200 },
      { duration: '30s', target: 300 },
      { duration: '30s', target: 500 },
      { duration: '30s', target: 0 },
    ],
    thresholds: {
      http_req_duration: ['p(95)<4000'],
      spike_errors: ['rate<0.15'],
    },
  },
  
  // Sudden burst test
  sudden_burst: {
    stages: [
      { duration: '10s', target: 10 },
      { duration: '1s', target: 300 },    // Very sudden
      { duration: '10s', target: 300 },
      { duration: '5s', target: 0 },
    ],
    thresholds: {
      http_req_duration: ['p(95)<6000'],
      spike_errors: ['rate<0.25'],
    },
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:3000';
const API_BASE_URL = __ENV.API_BASE_URL || 'http://localhost:3000/api/v1';

// Setup authentication
export function setup() {
  const analystRes = http.post(
    `${API_BASE_URL}/auth/login`,
    JSON.stringify({
      email: __ENV.TEST_ANALYST_EMAIL || 'analyst@test.gov',
      password: __ENV.TEST_ANALYST_PASSWORD || 'test_password123',
    }),
    { headers: { 'Content-Type': 'application/json' } }
  );
  
  return {
    token: analystRes.status === 200 ? JSON.parse(analystRes.body).token : '',
  };
}

// Spike test - burst of concurrent requests
export default function (data) {
  // Track concurrent connections
  concurrentConnections.add(1);
  
  // Execute burst of requests
  const burstSize = randomIntBetween(5, 15);
  
  for (let i = 0; i < burstSize; i++) {
    spikeTestRequest(data.token);
    sleep(randomIntBetween(50, 200) / 1000); // 50-200ms delay
  }
  
  spikeErrorRate.add(0); // Will be updated by check failures
}

function spikeTestRequest(token) {
  // Test critical endpoints that would see traffic spike
  const endpoints = [
    { method: 'GET', path: '/dashboard', weight: 3 },
    { method: 'GET', path: '/intelligence', weight: 2 },
    { method: 'GET', path: '/interventions', weight: 2 },
    { method: 'GET', path: '/crisis/status', weight: 4 },
    { method: 'GET', path: '/reports', weight: 1 },
  ];
  
  // Weighted random endpoint selection
  const totalWeight = endpoints.reduce((sum, e) => sum + e.weight, 0);
  let randomWeight = randomIntBetween(1, totalWeight);
  let selectedEndpoint = endpoints[0];
  
  for (const endpoint of endpoints) {
    randomWeight -= endpoint.weight;
    if (randomWeight <= 0) {
      selectedEndpoint = endpoint;
      break;
    }
  }
  
  const startTime = Date.now();
  
  const res = http.get(
    `${BASE_URL}${selectedEndpoint.path}`,
    {
      headers: {
        Authorization: `Bearer ${token}`,
        Accept: 'application/json',
      },
    }
  );
  
  const duration = Date.now() - startTime;
  spikeResponseTime.add(duration);
  
  const success = check(res, {
    'spike request succeeds': (r) => r.status < 500,
    'response time acceptable': (r) => r.timings.duration < 5000,
  });
  
  spikeErrorRate.add(!success);
}
