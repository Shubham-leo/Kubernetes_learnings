# Lab 2: Load Testing with Grafana K6

## What is K6?
K6 is an open-source **load testing tool** by Grafana. It simulates
thousands of users hitting your app to test performance.

```
K6 sends:
  100 virtual users → hitting your app → for 30 seconds

You see:
  Response time: 50ms avg
  Requests/sec: 2000
  Errors: 0%
  ✅ App handles the load!
```

## K6 vs wget (Lab-1)
```
wget (Lab-1):    just floods requests, no details
K6 (Lab-2):      controlled users + response times + error rates + pass/fail checks
```

## Why load test on K8s?
```
1. Find how many requests your app handles
2. Test if HPA auto-scaling works under load
3. Compare different runtimes (Django vs Go)
4. Find breaking points before production
```

## Install K6

### Windows
```powershell
# Option 1: Using chocolatey
choco install k6

# Option 2: Using winget
winget install k6

# Option 3: Download from
# https://github.com/grafana/k6/releases
# Extract k6.exe to a folder in your PATH
```

### Verify
```bash
k6 version
```

## Files in this folder

| File | What it does |
|------|-------------|
| `app-deployment.yaml` | Simple nginx app to test against |
| `k6-basic-test.js` | Basic load test (10 users, 30 seconds) |
| `k6-stress-test.js` | Stress test (ramp up to 100 users) |
| `k6-spike-test.js` | Spike test (sudden burst of traffic) |

## Commands to Run

### 1. Deploy an app to test against
```bash
kubectl apply -f app-deployment.yaml
kubectl get pods
kubectl get svc
```

### 2. Port-forward to access from K6
```bash
# Run in a separate terminal (keep it running):
kubectl port-forward service/test-app-svc 8080:80
```

### 3. Run basic load test
```bash
k6 run k6-basic-test.js
```

### 4. Read the results
```
K6 output explained:

  http_reqs............: 5000      ← total requests sent
  http_req_duration....: avg=50ms  ← average response time
  http_req_failed......: 0.00%     ← error rate
  vus..................: 10        ← virtual users (concurrent)
  iteration_duration...: avg=52ms  ← time per test iteration

Good results:
  ✅ avg response < 200ms
  ✅ error rate = 0%
  ✅ no timeouts

Bad results:
  ❌ avg response > 1s
  ❌ error rate > 1%
  ❌ timeouts increasing
```

### 5. How to read K6 results (example from basic test)
```
✅ http_req_failed: 0.00%        → zero errors, app is healthy
✅ status is 200: 100%           → all requests succeeded
✅ response time < 200ms: 99%    → almost all under 200ms

Response times:
  avg:  16ms     ← average response time
  med:  2.5ms    ← median (most requests this fast)
  p(95): 4ms     ← 95% of requests under this
  max:  3.4s     ← slowest single request

Throughput:
  2568 total requests in 30 seconds
  85 requests/second with 10 users
```

### Key metrics to watch
```
avg    = average (can be misleading if few slow requests)
med    = median (better indicator of typical experience)
p(95)  = 95th percentile (worst case for most users)
p(99)  = 99th percentile (worst case for almost all users)
max    = single slowest request (outlier)

Real world targets:
  API:      p(95) < 200ms
  Web page: p(95) < 500ms
  Database: p(95) < 50ms
```

### 6. Run stress test (ramp up users)
```bash
k6 run k6-stress-test.js
# Watch response times increase as load increases
```

### 6. Run spike test (sudden burst)
```bash
k6 run k6-spike-test.js
# Simulates a sudden traffic spike
```

### 7. Combine with HPA (the fun part!)
```bash
# Apply HPA from Lab-1:
kubectl apply -f ../Lab-1-AutoScaling/hpa.yaml

# Run stress test and watch pods auto-scale:
# Terminal 1: kubectl get hpa -w
# Terminal 2: kubectl get pods -w
# Terminal 3: k6 run k6-stress-test.js
```

### 8. Clean up
```bash
kubectl delete -f app-deployment.yaml
```
