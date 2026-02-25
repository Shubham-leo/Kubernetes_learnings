// Stress test - gradually increase users to find breaking point
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 10 },    // ramp up to 10 users
    { duration: '30s', target: 50 },    // ramp up to 50 users
    { duration: '30s', target: 100 },   // ramp up to 100 users
    { duration: '30s', target: 100 },   // stay at 100 users
    { duration: '30s', target: 0 },     // ramp down to 0
  ],
};

export default function () {
  const res = http.get('http://localhost:8080');

  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });

  sleep(0.1);
}
