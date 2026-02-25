import http from 'k6/http';
import { check, sleep } from 'k6';

// Spike test: 5 users → 100 users in 5 seconds, hold 1 minute
export const options = {
  stages: [
    { duration: '5s', target: 5 },
    { duration: '5s', target: 100 },
    { duration: '1m', target: 100 },
    { duration: '10s', target: 0 },
  ],
};

export default function () {
  const res = http.get('http://localhost:8085');
  check(res, {
    'status 200': (r) => r.status === 200,
  });
  sleep(0.05);
}
