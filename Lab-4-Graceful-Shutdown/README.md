# Lab 4: Graceful Shutdown in Kubernetes

## The Problem

When Kubernetes terminates a pod (rolling update, scale-down, node drain), there's a **race condition**:

```
Timeline:
  t=0    K8s decides to terminate the pod
  t=0    SIGTERM sent to the pod          ← pod starts dying
  t=0    kube-proxy starts updating       ← but iptables still routes here!
  t=1-3s kube-proxy finishes              ← NOW traffic stops
  t=???  Pod exits
```

During that 1-3 second window, **requests still arrive at the dying pod**. If the pod has already closed its listener, clients get **502 Bad Gateway**. If the pod accepted the connection but is shutting down mid-response, clients get **504 Gateway Timeout**.

This is especially bad in upstream chains where a Python frontend calls a Go backend — both can be terminating at the same time.

## Architecture

```
┌──────────────┐     ┌─────────────────────┐     ┌──────────────────┐
│  K6 Load     │────▶│  Python Frontend    │────▶│  Go Backend      │
│  Tester      │     │  (Flask + gunicorn) │     │  (net/http)      │
│              │     │  port 5000          │     │  port 8080       │
└──────────────┘     └─────────────────────┘     └──────────────────┘
                           │                           │
                     frontend-svc:5000           backend-svc:8080
                     (NodePort 30500)            (ClusterIP)
```

## The Fix

```
Timeline (with preStop hook):
  t=0    K8s decides to terminate the pod
  t=0    preStop hook runs (sleep 5s)     ← pod KEEPS SERVING
  t=0    kube-proxy starts updating       ← happening in parallel
  t=1-3s kube-proxy finishes              ← traffic stops arriving
  t=5s   preStop finishes, SIGTERM sent   ← NOW pod starts shutdown
  t=5s   http.Server.Shutdown()           ← drains in-flight requests
  t=5-30s Pod exits cleanly               ← zero dropped requests
```

## Quick Start

### Prerequisites
- Minikube running (`minikube start`)
- kubectl configured
- K6 installed (`brew install k6` / `choco install k6`)

### Step 1: Build Images

```bash
# Point Docker to Minikube's daemon
eval $(minikube docker-env)

# Build both images
docker build -t go-backend:v1 ./go-backend/
docker build -t python-frontend:v1 ./python-frontend/
```

### Step 2: Deploy the "Problem" Version

```bash
kubectl apply -f k8s-problem/
```

Wait for pods to be ready:
```bash
kubectl get pods -w
```

### Step 3: Run the Load Test (Problem Version)

Terminal 1 — start the load test:
```bash
k6 run -e FRONTEND_URL=http://$(minikube ip):30500 k6-test.js
```

Terminal 2 — trigger a rolling update while the test is running:
```bash
kubectl rollout restart deployment/go-backend && \
kubectl rollout restart deployment/python-frontend
```

**Expected result:** You'll see 502 and 504 errors in the K6 output.

### Step 4: Deploy the "Solution" Version

```bash
kubectl delete -f k8s-problem/
kubectl apply -f k8s-solution/
```

Wait for pods to be ready:
```bash
kubectl get pods -w
```

### Step 5: Run the Load Test (Solution Version)

Terminal 1 — start the load test:
```bash
k6 run -e FRONTEND_URL=http://$(minikube ip):30500 k6-test.js
```

Terminal 2 — trigger a rolling update:
```bash
kubectl rollout restart deployment/go-backend && \
kubectl rollout restart deployment/python-frontend
```

**Expected result:** Zero errors. The error_rate metric stays at 0%.

## What Changed Between Problem and Solution?

| Setting | Problem | Solution |
|---------|---------|----------|
| `preStop` hook | None | `httpGet /prestop` (5s sleep) |
| `terminationGracePeriodSeconds` | 30 (default) | 60 |
| `maxUnavailable` | 1 | 0 |
| `maxSurge` | 1 | 1 |
| Readiness probe | None | `httpGet /health` every 3s |
| Liveness probe | None | `httpGet /health` every 10s |
| Backend timeouts | Default (infinite) | connect=2s, read=10s |

## Deep Dive: Why Each Fix Matters

### 1. preStop Hook (5-second sleep)

The preStop hook runs **in parallel** with kube-proxy removing the pod from endpoints. The 5-second sleep gives kube-proxy enough time to update iptables/IPVS rules so no new traffic arrives at this pod.

Without it, the pod starts shutting down while kube-proxy still routes traffic to it.

### 2. SIGTERM Handling (http.Server.Shutdown)

After preStop completes, K8s sends SIGTERM. Go's `http.Server.Shutdown()`:
1. **Stops accepting new connections** (closes the listener)
2. **Waits for in-flight requests to finish** (drains gracefully)
3. **Returns** when all requests are done

Without it, the process exits immediately and drops in-flight requests.

### 3. maxUnavailable: 0

With `maxUnavailable: 1`, K8s can terminate a pod before the new one is ready. With `maxUnavailable: 0` and `maxSurge: 1`, K8s:
1. Creates a new pod
2. Waits for it to pass readiness probe
3. Only then terminates the old pod

This ensures there's always enough capacity to handle traffic.

### 4. Readiness Probes

Without readiness probes, K8s considers a pod "Ready" as soon as the container starts. With probes, K8s waits until the app is actually serving before routing traffic to it.

During shutdown, the health endpoint returns 503, which causes the readiness probe to fail. K8s removes the pod from endpoints (belt-and-suspenders with the preStop hook).

### 5. Explicit Timeouts (Python → Go)

Without explicit timeouts, Python's `requests.get()` waits **indefinitely** for a response. If the Go backend dies mid-request, the Python process hangs until the TCP connection times out (which can be minutes).

With `timeout=(2, 10)` (2s connect, 10s read), the Python frontend fails fast and returns a clear error.

### 6. terminationGracePeriodSeconds: 60

This is the **total time budget** for the entire shutdown sequence:
- 5s preStop hook
- Up to 25s for in-flight request drain
- 30s buffer

The default 30s may not be enough if you have long-running requests + the preStop sleep.

## TCP Connection Lifecycle

Understanding why 502 and 504 are different errors:

```
502 Bad Gateway:
  Client ──TCP SYN──▶ Dead Pod
  Client ◀──TCP RST── Dead Pod    ← connection refused, pod is already gone

504 Gateway Timeout:
  Client ──TCP SYN──▶ Dying Pod
  Client ◀──TCP ACK── Dying Pod   ← connection established!
  Client ──HTTP GET─▶ Dying Pod
  Client ◀──........── Dying Pod  ← no response, pod dying mid-request
  ... timeout ...
```

TCP is **full duplex** — the connection has two independent directions. A pod can accept a TCP connection (SYN-ACK) but never send the HTTP response if it's killed mid-processing.

## Automated Rolling Update Test

For a more structured test with distinct phases:

```bash
k6 run -e FRONTEND_URL=http://$(minikube ip):30500 k6-rolling-update-test.js
```

This test has three phases:
1. **Warmup** (15s) — verify baseline works
2. **Sustained load** (90s) — trigger rolling update during this phase
3. **Cooldown** (15s) — catch trailing errors

## Cleanup

```bash
kubectl delete -f k8s-solution/   # or k8s-problem/
```
