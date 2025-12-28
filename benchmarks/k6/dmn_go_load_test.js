import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const evaluationDuration = new Trend('evaluation_duration');

// Load test configuration
export const options = {
  stages: [
    { duration: '30s', target: 10 },  // Ramp up to 10 users
    { duration: '1m', target: 50 },   // Ramp up to 50 users
    { duration: '2m', target: 50 },   // Stay at 50 users
    { duration: '30s', target: 100 }, // Spike to 100 users
    { duration: '1m', target: 100 },  // Stay at 100 users
    { duration: '30s', target: 0 },   // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<1500'], // 95% of requests should be below 1500ms (realistic for large DMN sets)
    'http_req_duration{expected_response:true}': ['p(99)<3000'], // 99% below 3s
    errors: ['rate<0.1'],               // Error rate should be below 10%
  },
};

// Get available decision keys
let decisionKeys = [];

export function setup() {
  const baseUrl = __ENV.DMN_GO_URL || 'http://localhost:8080';
  
  console.log(`Fetching decision definitions from ${baseUrl}...`);
  
  const res = http.get(`${baseUrl}/api/v1/definitions`);
  
  if (res.status === 200) {
    const definitions = JSON.parse(res.body);
    decisionKeys = definitions.map(def => def.key);
    console.log(`Found ${decisionKeys.length} decisions`);
    console.log(`Sample keys: ${decisionKeys.slice(0, 5).join(', ')}`);
  } else {
    console.error(`Failed to fetch decisions: ${res.status}`);
  }
  
  return { baseUrl, decisionKeys };
}

export default function(data) {
  if (data.decisionKeys.length === 0) {
    console.error('No decision keys available');
    return;
  }
  
  // Pick random decision
  const randomKey = data.decisionKeys[Math.floor(Math.random() * data.decisionKeys.length)];
  
  // Random age for testing
  const age = Math.floor(Math.random() * 100) + 1;
  
  const payload = JSON.stringify({
    decisionKey: randomKey,
    variables: {
      age: age
    }
  });
  
  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };
  
  const url = `${data.baseUrl}/api/v1/evaluate`;
  const startTime = Date.now();
  const res = http.post(url, payload, params);
  const duration = Date.now() - startTime;
  
  evaluationDuration.add(duration);
  
  const success = check(res, {
    'status is 200': (r) => r.status === 200,
    'response has outputs': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.outputs && body.outputs.length > 0;
      } catch {
        return false;
      }
    },
  });
  
  errorRate.add(!success);
  
  sleep(0.1); // Small delay between requests
}

export function teardown(data) {
  console.log(`\n=== Test Summary ===`);
  console.log(`Base URL: ${data.baseUrl}`);
  console.log(`Decisions tested: ${data.decisionKeys.length}`);
}

