/**
 * K6 Load Test — Hit Go gateway during rolling restarts.
 *
 * 20 VUs for ~2 minutes. While this runs, trigger a rolling update:
 *   kubectl rollout restart deployment/go-gateway
 *
 * Usage:
 *   k6 run loadtest/test.js
 *   GATEWAY_URL=http://localhost:8080 k6 run loadtest/test.js
 */

import http from "k6/http";
import { check, sleep } from "k6";
import { Counter, Rate, Trend } from "k6/metrics";

// Custom metrics — separate 502/504 so we can tell connection-refused from timeout
const errorRate = new Rate("error_rate");
const errors502 = new Counter("errors_502");
const errors504 = new Counter("errors_504");
const errorsOther = new Counter("errors_other");
const successCount = new Counter("success_count");
const retryTotal = new Counter("retry_total");
const responseTimes = new Trend("response_time_ms");

const GATEWAY_URL = __ENV.GATEWAY_URL || "http://192.168.49.2:30080";

export const options = {
  stages: [
    { duration: "10s", target: 10 },  // warm up
    { duration: "90s", target: 20 },  // sustained — trigger rolling update here
    { duration: "10s", target: 0 },   // cool down
  ],
  thresholds: {
    error_rate: [{ threshold: "rate<0.01", abortOnFail: false }],
    http_req_duration: ["p(95)<5000"],
  },
};

export default function () {
  const res = http.get(`${GATEWAY_URL}/`, {
    timeout: "15s",
  });

  responseTimes.add(res.timings.duration);

  // Track retries reported by the gateway (v3 only)
  try {
    const body = JSON.parse(res.body);
    if (body.retries > 0) {
      retryTotal.add(body.retries);
    }
  } catch {
    // ignore parse errors
  }

  const success = check(res, {
    "status is 200": (r) => r.status === 200,
    "response is valid": (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.status === "ok" && body.worker && body.gateway;
      } catch {
        return false;
      }
    },
  });

  if (success) {
    successCount.add(1);
    errorRate.add(false);
  } else {
    errorRate.add(true);
    if (res.status === 502) {
      errors502.add(1);
    } else if (res.status === 504) {
      errors504.add(1);
    } else {
      errorsOther.add(1);
    }
    console.log(
      `ERROR: status=${res.status} body=${(res.body || "").substring(0, 200)}`
    );
  }

  sleep(0.1);
}

export function handleSummary(data) {
  const ok = data.metrics.success_count?.values?.count || 0;
  const e502 = data.metrics.errors_502?.values?.count || 0;
  const e504 = data.metrics.errors_504?.values?.count || 0;
  const eOther = data.metrics.errors_other?.values?.count || 0;
  const total = ok + e502 + e504 + eOther;
  const totalErrors = e502 + e504 + eOther;
  const retries = data.metrics.retry_total?.values?.count || 0;
  const avgMs = data.metrics.response_time_ms?.values?.avg?.toFixed(1) || "N/A";
  const p95Ms =
    data.metrics.response_time_ms?.values?.["p(95)"]?.toFixed(1) || "N/A";
  const maxMs = data.metrics.response_time_ms?.values?.max?.toFixed(1) || "N/A";

  console.log("\n" + "=".repeat(55));
  console.log("  GRACEFUL SHUTDOWN TEST RESULTS");
  console.log("=".repeat(55));
  console.log(`  Total requests:   ${total}`);
  console.log(`  Successful:       ${ok}`);
  console.log(`  502 errors:       ${e502}`);
  console.log(`  504 errors:       ${e504}`);
  console.log(`  Other errors:     ${eOther}`);
  console.log(
    `  Error rate:       ${total > 0 ? ((totalErrors / total) * 100).toFixed(2) : "0.00"}%`
  );
  console.log(`  Retries:          ${retries}`);
  console.log("-".repeat(55));
  console.log(`  Avg latency:      ${avgMs}ms`);
  console.log(`  p95 latency:      ${p95Ms}ms`);
  console.log(`  Max latency:      ${maxMs}ms`);
  console.log("=".repeat(55));

  if (totalErrors === 0) {
    console.log("  PASS — zero errors during rolling update");
  } else {
    console.log("  FAIL — requests dropped during rolling update");
  }
  console.log("");

  return {};
}
