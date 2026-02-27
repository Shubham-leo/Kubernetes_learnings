/**
 * K6 Rolling Update Test — Triggers a rolling update mid-load.
 *
 * This test does the same thing as k6-test.js but is designed to be
 * self-contained: you run this test, and in a SEPARATE terminal, trigger
 * the rolling update at the right moment.
 *
 * The test prints clear markers so you know when to trigger:
 *
 *   Terminal 1: k6 run k6-rolling-update-test.js
 *   Terminal 2: (when you see "TRIGGER NOW")
 *     kubectl rollout restart deployment/go-backend && \
 *     kubectl rollout restart deployment/python-frontend
 *
 * Usage:
 *   # Set the Minikube IP:
 *   k6 run -e FRONTEND_URL=http://$(minikube ip):30500 k6-rolling-update-test.js
 */

import http from "k6/http";
import { check, sleep } from "k6";
import { Rate, Counter, Trend } from "k6/metrics";

const errorRate = new Rate("error_rate");
const errors502 = new Counter("errors_502");
const errors504 = new Counter("errors_504");
const errorsOther = new Counter("errors_other");
const responseTimes = new Trend("response_time_ms");

const FRONTEND_URL = __ENV.FRONTEND_URL || "http://192.168.49.2:30500";

export const options = {
  scenarios: {
    // Phase 1: Warm up — verify everything works before the update
    warmup: {
      executor: "constant-vus",
      vus: 5,
      duration: "15s",
      startTime: "0s",
      tags: { phase: "warmup" },
    },
    // Phase 2: Sustained load — this is when you trigger the rolling update
    rolling_update: {
      executor: "constant-vus",
      vus: 20,
      duration: "90s",
      startTime: "15s",
      tags: { phase: "rolling_update" },
    },
    // Phase 3: Cool down — catch any trailing errors
    cooldown: {
      executor: "constant-vus",
      vus: 5,
      duration: "15s",
      startTime: "105s",
      tags: { phase: "cooldown" },
    },
  },
  thresholds: {
    error_rate: ["rate<0.01"],
    http_req_duration: ["p(95)<5000"],
  },
};

// Track which phase we're in
let phaseLogged = {};

export default function () {
  const scenario = __ENV.K6_SCENARIO || "unknown";

  // Log phase transitions
  if (!phaseLogged[scenario]) {
    phaseLogged[scenario] = true;
    if (scenario === "warmup") {
      console.log("\n>>> PHASE 1: WARMUP — verifying baseline...");
    } else if (scenario === "rolling_update") {
      console.log("\n>>> PHASE 2: SUSTAINED LOAD");
      console.log(">>> =============================================");
      console.log(">>> TRIGGER NOW! Run in another terminal:");
      console.log(">>>   kubectl rollout restart deployment/go-backend && \\");
      console.log(">>>   kubectl rollout restart deployment/python-frontend");
      console.log(">>> =============================================\n");
    } else if (scenario === "cooldown") {
      console.log("\n>>> PHASE 3: COOLDOWN — catching trailing errors...");
    }
  }

  const res = http.get(`${FRONTEND_URL}/`, {
    timeout: "15s",
  });

  responseTimes.add(res.timings.duration);

  const success = check(res, {
    "status is 200": (r) => r.status === 200,
  });

  errorRate.add(!success);

  if (res.status === 502) {
    errors502.add(1);
  } else if (res.status === 504) {
    errors504.add(1);
  } else if (res.status !== 200) {
    errorsOther.add(1);
  }

  sleep(0.3);
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
  console.log("ROLLING UPDATE TEST RESULTS");
  console.log("=".repeat(60));
  console.log(`Total requests:    ${total}`);
  console.log(`502 errors:        ${e502} (backend connection refused)`);
  console.log(`504 errors:        ${e504} (backend timeout)`);
  console.log(`Other errors:      ${eOther}`);
  console.log(`Total errors:      ${totalErrors}`);
  console.log(
    `Error rate:        ${total > 0 ? ((totalErrors / total) * 100).toFixed(2) : 0}%`
  );
  console.log("=".repeat(60));

  if (totalErrors === 0) {
    console.log("PASS: Zero-downtime rolling update achieved!");
  } else {
    console.log(
      "FAIL: Errors detected — this is expected with problem manifests."
    );
    console.log("       Deploy solution manifests and re-run to see the fix.");
  }

  return {};
}
