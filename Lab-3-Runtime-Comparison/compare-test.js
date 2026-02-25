// K6 comparison test - hits all 3 apps and compares performance
import http from 'k6/http';
import { check, sleep, group } from 'k6';

export const options = {
  vus: 20,
  duration: '30s',
};

export default function () {
  // Test Nginx (baseline)
  group('Nginx (C)', () => {
    const res = http.get('http://localhost:8081');
    check(res, {
      'nginx: status 200': (r) => r.status === 200,
      'nginx: response < 100ms': (r) => r.timings.duration < 100,
    });
  });

  // Test Go app
  group('Go App', () => {
    const res = http.get('http://localhost:8082');
    check(res, {
      'go: status 200': (r) => r.status === 200,
      'go: response < 100ms': (r) => r.timings.duration < 100,
    });
  });

  // Test Django app
  group('Django App', () => {
    const res = http.get('http://localhost:8083');
    check(res, {
      'django: status 200': (r) => r.status === 200,
      'django: response < 500ms': (r) => r.timings.duration < 500,
    });
  });

  sleep(0.1);
}
