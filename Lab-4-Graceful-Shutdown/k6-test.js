/**
 * K6 Load Test — Continuous requests to the Python frontend.
 *
 * This test hammers the frontend for 2 minutes at a steady rate.
 * While this runs, you trigger a rolling update in another terminal:
 *   kubectl rollout restart deployment/go-backend
 *   kubectl rollout restart deployment/python-frontend
 *
 * Watch the error rate:
 *   - Problem manifests: you'll see http_req_failed spike (502/504 errors)
 *   - Solution manifests: http_req_failed stays at 0%
 *
 * Usage:
 *   k6 run k6-test.js
 */

import http from "k6/http";
import { check, sleep } from "k6";
import { Rate, Counter, Trend } from "k6/metrics";

// Custom metrics for clear visibility
const errorRate = new Rate("error_rate");
const errors502 = new Counter("errors_502");
const errors504 = new Counter("errors_504");
const errorsOther = new Counter("errors_other");
const responseTimes = new Trend("response_time_ms");

// Get frontend URL from environment or use default Minikube NodePort
const FRONTEND_URL = __ENV.FRONTEND_URL || "http://192.168.49.2:30500";

export const options = {
  // Ramp up to 20 VUs, sustain for 2 minutes, ramp down
  stages: [
    { duration: "10s", target: 10 }, // ramp up
    { duration: "2m", target: 20 }, // sustained load — trigger rolling update during this phase
    { duration: "10s", target: 0 }, // ramp down
  ],
  // Thresholds — the test "passes" if error rate is under 1%
  thresholds: {
    error_rate: ["rate<0.01"], // less than 1% errors
    http_req_duration: ["p(95)<5000"], // 95th percentile under 5s
  },
};

export default function () {
  const res = http.get(`${FRONTEND_URL}/`, {
    timeout: "15s", // generous timeout to catch 504s (not mask them)
  });

  // Track response time
  responseTimes.add(res.timings.duration);

  // Check for success
  const success = check(res, {
    "status is 200": (r) => r.status === 200,
    "response has backend data": (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.backend && body.backend.hostname;
      } catch {
        return false;
      }
    },
  });

  // Track errors by type
  errorRate.add(!success);

  if (res.status === 502) {
    errors502.add(1);
    console.log(`502 Bad Gateway — backend connection refused (pod died)`);
  } else if (res.status === 504) {
    errors504.add(1);
    console.log(`504 Gateway Timeout — backend hung during shutdown`);
  } else if (res.status !== 200) {
    errorsOther.add(1);
    console.log(`Unexpected status ${res.status}: ${res.body}`);
  }

  // Small pause between requests (simulates real user behavior)
  sleep(0.5);
}

export function handleSummary(data) {
  const total = data.metrics.http_reqs ? data.metrics.http_reqs.values.count : 0;
  const e502 = data.metrics.errors_502
    ? data.metrics.errors_502.values.count
    : 0;
  const e504 = data.metrics.errors_504
    ? data.metrics.errors_504.values.count
    : 0;
  const eOther = data.metrics.errors_other
    ? data.metrics.errors_other.values.count
    : 0;
  const totalErrors = e502 + e504 + eOther;

  console.log("\n" + "=".repeat(60));
  console.log("GRACEFUL SHUTDOWN TEST RESULTS");
  console.log("=".repeat(60));
  console.log(`Total requests:    ${total}`);
  console.log(`502 errors:        ${e502}`);
  console.log(`504 errors:        ${e504}`);
  console.log(`Other errors:      ${eOther}`);
  console.log(`Total errors:      ${totalErrors}`);
  console.log(
    `Error rate:        ${total > 0 ? ((totalErrors / total) * 100).toFixed(2) : 0}%`
  );
  console.log("=".repeat(60));

  if (totalErrors === 0) {
    console.log("PASS: Zero errors during rolling update!");
  } else {
    console.log("FAIL: Errors detected — graceful shutdown is not working correctly.");
  }

  return {};
}
