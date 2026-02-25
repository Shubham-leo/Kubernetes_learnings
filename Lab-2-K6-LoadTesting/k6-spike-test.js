// Spike test - sudden burst of traffic
// Tests how your app handles unexpected traffic spikes
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '10s', target: 5 },     // normal traffic
    { duration: '5s', target: 200 },    // SPIKE! sudden burst
    { duration: '30s', target: 200 },   // stay at spike
    { duration: '10s', target: 5 },     // back to normal
    { duration: '10s', target: 0 },     // cool down
  ],
};

export default function () {
  const res = http.get('http://localhost:8080');

  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 1000ms': (r) => r.timings.duration < 1000,
  });

  sleep(0.05);
}
