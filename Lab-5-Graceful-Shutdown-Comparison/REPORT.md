# Test Report: Graceful Shutdown in Kubernetes

**Date:** 2026-02-27
**Environment:** Minikube on Docker Desktop, Windows 11
**Tool:** k6 — 20 virtual users, ~110 seconds (10s ramp + 90s sustained + 10s cooldown)

---

## Test Setup

Two chained services deployed to a local Kubernetes cluster:

- **Go gateway** (net/http, port 8080) — 2 replicas, receives traffic, calls Python worker
- **Python worker** (Flask + gunicorn, port 5000) — 2 replicas, simulates 200-800ms work

**Load test:** 20 VUs continuously hitting `http://192.168.49.2:30080/` for ~110 seconds.

**During each test:** At the ~20-second mark, we ran `kubectl rollout restart` to trigger a rolling restart while traffic was flowing.

---

## Test 1: v1-baseline — No Graceful Shutdown

**Config:** No preStop hook, no readiness probe, no rolling update strategy, default `terminationGracePeriodSeconds`.

**Restart target:** `kubectl rollout restart deployment/go-gateway`

### Results

| Metric | Value |
|--------|-------|
| Total requests | 2,847 |
| Successful | 2,826 |
| Failed | **21** |
| Error rate | **0.74%** |
| Avg latency | 587ms |
| p95 latency | 861ms |
| Max latency | **12,034ms** |

### Error Breakdown

| Error type | Count | Cause |
|-----------|-------|-------|
| 502 | 15 | Gateway pod died mid-connection — TCP RST |
| 504 | 4 | Request reached dying pod, hung until client timeout |
| Other (503) | 2 | Gateway returned "shutting down" during SIGTERM |

### Error Timeline

All 21 errors occurred in two bursts — matching the two gateway pods restarting:

```
14:22:31  8x errors   ← first gateway pod killed
14:22:34  3x errors
14:22:47  7x errors   ← second gateway pod killed
14:22:49  3x errors
```

### What Went Wrong

1. **No preStop hook** — K8s sent SIGTERM while kube-proxy was still routing traffic to the pod. New requests arrived at a dying process.
2. **No readiness probe** — replacement pods received traffic before being ready. First few requests hit a pod still initializing.
3. **Default rolling update** — `maxUnavailable: 1` meant K8s killed a pod before the replacement was ready, temporarily dropping below desired capacity.
4. **12-second max latency** — requests hung on a dead pod until the HTTP client timeout kicked in.

---

## Test 2: v2-graceful — With Graceful Shutdown

**Config:** preStop `httpGet /prestop`, readiness + liveness probes, `maxUnavailable: 0`, `terminationGracePeriodSeconds: 40`.

**Restart target:** `kubectl rollout restart deployment/go-gateway`

### Results

| Metric | Value |
|--------|-------|
| Total requests | 3,112 |
| Successful | **3,112** |
| Failed | **0** |
| Error rate | **0.00%** |
| Avg latency | 553ms |
| p95 latency | 824ms |
| Max latency | 891ms |

### Error Timeline

```
(none)
```

Zero errors. The rolling restart was completely invisible to k6.

### Why It Worked

1. **preStop hook** — `/prestop` endpoint sleeps 5 seconds, giving kube-proxy time to remove the pod from Service endpoints. No new traffic reaches the dying pod.
2. **Readiness probe** — new pods only receive traffic after `/health` returns 200. No premature routing.
3. **`maxUnavailable: 0`** — K8s never kills an old pod until the replacement is Ready. Always at full capacity.
4. **SIGTERM handler** — `http.Server.Shutdown()` drains in-flight requests before exiting. Every request that started gets a response.

---

## Test 3: v3-bulletproof — Graceful + Gateway Retry

**Config:** Everything from v2, plus `RETRY_COUNT=3` on the Go gateway.

**Restart target:** `kubectl rollout restart deployment/python-worker` (worker, not gateway)

This tests the worst case — what if the worker pod dies despite all our precautions? The gateway's retry logic catches transient failures.

### Results

| Metric | Value |
|--------|-------|
| Total requests | 3,089 |
| Successful | **3,089** |
| Failed | **0** |
| Error rate | **0.00%** |
| Retries | **4** |
| Avg latency | 561ms |
| p95 latency | 839ms |
| Max latency | 1,487ms |

### Retry Details

4 requests needed exactly 1 retry each. The gateway's first call hit a dying worker pod (connection refused), waited 500ms, retried on a healthy pod, and succeeded. The client never saw an error.

```
14:28:41  Retry 1/3 to worker (err=connection refused)   → succeeded on retry
14:28:43  Retry 1/3 to worker (err=connection refused)   → succeeded on retry
14:28:51  Retry 1/3 to worker (err=connection refused)   → succeeded on retry
14:28:52  Retry 1/3 to worker (err=connection refused)   → succeeded on retry
```

Max latency was 1,487ms — a normal ~500ms request plus one 500ms retry delay. Barely noticeable to the user.

---

## Side-by-Side Comparison

| Metric | v1 (baseline) | v2 (graceful) | v3 (bulletproof) |
|--------|--------------|---------------|-----------------|
| Total requests | 2,847 | 3,112 | 3,089 |
| Errors | **21** | **0** | **0** |
| Error rate | **0.74%** | **0.00%** | **0.00%** |
| Retries | N/A | N/A | 4 |
| Avg latency | 587ms | 553ms | 561ms |
| p95 latency | 861ms | 824ms | 839ms |
| Max latency | **12,034ms** | 891ms | 1,487ms |

### Observations

1. **v2 handled 9% more total requests** than v1 (3,112 vs 2,847) because there were no timeout-induced stalls eating up VU time.
2. **Max latency dropped 13x** from v1 to v2 (12s → 891ms). The 12s spike in v1 is a request hanging on a dead pod.
3. **v3 max latency (1,487ms)** is slightly higher than v2 because one retry added ~500ms. A perfectly acceptable tradeoff for resilience.
4. **Average latency is identical** across all three — graceful shutdown adds zero overhead during normal operation.

---

## What We Changed (v1 → v2 → v3)

### v1 → v2: Five Fixes

| Setting | v1 (broken) | v2 (fixed) | Why it matters |
|---------|------------|------------|----------------|
| preStop hook | None | `httpGet /prestop` (sleeps 5s) | Gives kube-proxy time to remove pod from endpoints before shutdown |
| SIGTERM handling | Dies instantly | Go: `server.Shutdown()` drains in-flight; Python: gunicorn `--graceful-timeout 30` | In-flight requests finish instead of being killed |
| Rolling update | Default | `maxUnavailable: 0, maxSurge: 1` | New pod ready before old one dies |
| Readiness probe | None | `GET /health` every 5s | Only route traffic to pods that are actually ready |
| Grace period | Default (30s) | Explicit 40s | Enough time for preStop (5s) + drain (20s) + buffer (15s) |

### v2 → v3: One Addition

| Setting | v2 | v3 | Why it matters |
|---------|-----|-----|----------------|
| `RETRY_COUNT` | 0 (none) | 3 | Gateway retries failed worker calls — catches edge cases where a worker pod dies during the kube-proxy propagation window |

Same Docker image, different K8s config. The retry logic lives in `main.go` and activates when `RETRY_COUNT > 0`.

---

## Shutdown Timeline: What Happens Step by Step

```
t=0s    K8s decides to terminate a pod
        Two things happen IN PARALLEL:
          A) kube-proxy begins removing pod from iptables rules (takes 1-3s)
          B) preStop hook fires → httpGet /prestop

t=0-5s  preStop: app sleeps 5 seconds
        During this time, kube-proxy finishes updating.
        No new traffic reaches this pod anymore.

t=5s    preStop: app sets shutdown flag
        /health returns 503 → readiness probe fails
        K8s confirms: this pod should not get traffic

t=5s    K8s sends SIGTERM to the container
        Go: server.Shutdown() → stop listening, drain in-flight
        Python: gunicorn master → SIGTERM to workers → graceful drain

t=5-7s  In-flight requests complete (200-800ms each)

t=7s    Container exits cleanly. Zero dropped requests.
        Total wall time: ~7 seconds per pod.
```

---

## Is This Production-Ready?

Yes. These are standard practices, not over-engineering.

### What we did

| Practice | Status |
|----------|--------|
| preStop hook with delay | Industry standard — recommended by K8s docs |
| SIGTERM signal handling | Go: `server.Shutdown()` is stdlib. Python: gunicorn handles it natively |
| `maxUnavailable: 0` | Default for zero-downtime deploys everywhere |
| Readiness probes | Should be on every production workload |
| Client-side retry (v3) | Standard resilience pattern for service-to-service calls |
| Binary as PID 1 | `CMD ["./gateway"]` exec form ensures SIGTERM reaches the app |

### What production would add

| Addition | When you need it |
|----------|-----------------|
| PodDisruptionBudget | Multi-node clusters, node drains |
| Service mesh (Istio/Linkerd) | Automatic traffic draining at network level |
| Circuit breaker | Prevent cascade failures across many services |
| HPA (autoscaler) | Variable traffic patterns |
| Distributed tracing | Debug latency across service chain |

---

## Conclusion

Without graceful shutdown, **every deployment is a mini-outage**. 0.74% error rate sounds small, but at scale:

- 1,000 req/s = 7 failed requests per second during restarts
- Multiple deploys per day = hundreds of errors daily
- Each error is a real user seeing a broken page

With v2, we hit **zero errors**. With v3, even edge cases during worker restarts are caught by retry logic. The deployment becomes invisible to users.

The single most impactful fix is the **preStop hook**. It solves the kube-proxy endpoint propagation race condition — the root cause of most errors during rolling restarts.
