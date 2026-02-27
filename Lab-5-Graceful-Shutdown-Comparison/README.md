# Lab 5: Graceful Shutdown in Kubernetes

Deploy two chained services, blast them with load, restart pods mid-traffic, and watch what breaks. Then fix it. Then make it bulletproof.

## Architecture

```
k6 (load test)
  |
  v
Go Gateway (:8080)  ---HTTP--->  Python Worker (:5000)
  Service: go-gateway-svc          Service: python-worker-svc
  (NodePort 30080)                 (ClusterIP)
```

- **Go gateway** — entry point, receives external traffic, calls Python worker, returns combined JSON
- **Python worker** — leaf service, simulates 200-800ms of processing, returns result to gateway

Both run as 2 replicas in Kubernetes.

## Three Versions

| Version | What's different | Expected result |
|---------|-----------------|-----------------|
| **v1-baseline** | No preStop, no probes, default rolling update | Errors during restart |
| **v2-graceful** | preStop hooks, readiness probes, `maxUnavailable: 0` | Zero errors |
| **v3-bulletproof** | v2 + gateway retries failed worker calls (`RETRY_COUNT=3`) | Zero errors + retry transparency |

## Setup

### 1. Build images (inside Minikube's Docker)

```bash
eval $(minikube docker-env)
docker build -t go-gateway:v1 ./go-gateway/
docker build -t python-worker:v1 ./python-worker/
```

### 2. Get Minikube IP

```bash
minikube ip
# If not 192.168.49.2, update GATEWAY_URL when running k6
```

## Test 1: v1 — No Protection

```bash
kubectl apply -f k8s/v1-baseline/deploy-all.yaml
kubectl get pods -w   # wait until all 4 pods show Running
```

Two terminals:

```bash
# Terminal 1 — start load test
k6 run loadtest/test.js

# Terminal 2 — restart gateway ~15s after k6 starts
kubectl rollout restart deployment/go-gateway
```

Watch k6 output — you'll see 502/504 errors spike during the restart.

## Test 2: v2 — Graceful Shutdown

```bash
kubectl delete -f k8s/v1-baseline/deploy-all.yaml
kubectl apply -f k8s/v2-graceful/deploy-all.yaml
kubectl get pods -w   # wait until all pods show Running + Ready
```

Same test:

```bash
# Terminal 1
k6 run loadtest/test.js

# Terminal 2
kubectl rollout restart deployment/go-gateway
```

Zero errors. The restart is invisible to the load test.

## Test 3: v3 — Bulletproof (with retry)

```bash
kubectl delete -f k8s/v2-graceful/deploy-all.yaml
kubectl apply -f k8s/v3-bulletproof/deploy-all.yaml
kubectl get pods -w
```

This time, restart the **worker** (not the gateway):

```bash
# Terminal 1
k6 run loadtest/test.js

# Terminal 2
kubectl rollout restart deployment/python-worker
```

Zero errors. The gateway retries failed calls to the worker transparently. Check the `retry_total` metric in k6 output.

## Cleanup

```bash
kubectl delete -f k8s/v3-bulletproof/deploy-all.yaml
```

## What Changed: v1 → v2 → v3

### v1 → v2 (five fixes)

| Fix | What it does |
|-----|-------------|
| **preStop hook** | `httpGet /prestop` — 5s delay for kube-proxy to update routing |
| **SIGTERM handler** | Go: `server.Shutdown()` drains in-flight. Python: gunicorn `--graceful-timeout 30` |
| **`maxUnavailable: 0`** | New pod ready before old pod dies. Never below desired replica count |
| **Readiness probe** | `GET /health` — K8s only routes traffic to pods that are actually ready |
| **`terminationGracePeriodSeconds: 40`** | Enough time for preStop (5s) + drain (20s) + buffer (15s) |

### v2 → v3 (one addition)

| Fix | What it does |
|-----|-------------|
| **`RETRY_COUNT=3`** | Gateway retries failed calls to worker up to 3 times (500ms between retries) |

Same Go image, different config. The retry logic is built into `main.go` and activated by the env var.

## The Shutdown Sequence (v2/v3)

```
1. K8s creates a NEW pod                         (maxSurge: 1)
2. New pod passes readiness probe                 → starts getting traffic
3. K8s picks an OLD pod to terminate
4. preStop hook fires → httpGet /prestop
   - App sleeps 5s                                → kube-proxy removes pod from endpoints
   - App sets shutdown flag                       → /health returns 503
5. K8s sends SIGTERM to the container
6. Go: server.Shutdown() stops new connections, waits for in-flight to finish
   Python: gunicorn master sends SIGTERM to workers, waits for graceful-timeout
7. All in-flight requests complete
8. Container exits cleanly
9. Zero dropped requests
```

## Project Structure

```
Lab-5-Graceful-Shutdown-Comparison/
├── go-gateway/
│   ├── main.go                  # net/http server, calls worker, optional retry
│   └── Dockerfile               # Multi-stage (golang:1.22 → alpine:3.19)
├── python-worker/
│   ├── app.py                   # Flask app, /process + /health + /prestop
│   ├── requirements.txt
│   └── Dockerfile               # python:3.12-slim + gunicorn
├── k8s/
│   ├── v1-baseline/
│   │   └── deploy-all.yaml      # No protection
│   ├── v2-graceful/
│   │   └── deploy-all.yaml      # preStop + probes + safe rolling update
│   └── v3-bulletproof/
│       └── deploy-all.yaml      # v2 + RETRY_COUNT=3 on gateway
├── loadtest/
│   └── test.js                  # k6: 20 VUs, ~2 min, tracks 502/504/retries
├── REPORT.md                    # Test results and analysis
└── README.md
```
