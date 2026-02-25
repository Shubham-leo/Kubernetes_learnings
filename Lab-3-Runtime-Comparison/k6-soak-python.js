import http from 'k6/http';
import { check, sleep } from 'k6';

// Soak test: 30 users sustained for 5 minutes
export const options = {
  stages: [
    { duration: '10s', target: 30 },
    { duration: '5m', target: 30 },
    { duration: '10s', target: 0 },
  ],
};

export default function () {
  const res = http.get('http://localhost:8083');
  check(res, {
    'status 200': (r) => r.status === 200,
  });
  sleep(0.05);
}
