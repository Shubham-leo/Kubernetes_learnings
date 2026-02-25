// Basic load test - 10 virtual users for 30 seconds
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 10,              // 10 virtual users
  duration: '30s',      // run for 30 seconds
};

export default function () {
  const res = http.get('http://localhost:8080');

  // Check if response is OK
  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 200ms': (r) => r.timings.duration < 200,
  });

  sleep(0.1);  // small pause between requests
}
