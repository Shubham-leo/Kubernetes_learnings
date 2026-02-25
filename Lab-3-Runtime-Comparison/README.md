# Lab 3: Runtime Comparison on Kubernetes

> Rust vs Go vs Python — same app, same cluster, same load. Only the language changes.

**Tested on:** 2026-02-25 | Minikube + Docker Desktop | Windows 11 | K6 load testing

---

## What This Lab Does

Deploys the same fibonacci(30) API in 3 languages on K8s and load tests each with Grafana K6 (100 concurrent users). See [RESULTS.md](./RESULTS.md) for findings.

---

## Apps

| | Rust | Go | Python |
|--|------|-----|--------|
| Version | 1.85 | 1.22 | 3.12 |
| Framework | Actix-Web 4 | net/http (stdlib) | Flask |
| Base image | debian:bookworm-slim | golang:1.22-alpine | python:3.12-slim |
| Build method | Multi-stage Docker | Inline `go run` | Inline + `pip install` |
| Port | 8090 | 8080 | 5000 |
| Port-forward | localhost:8085 | localhost:8082 | localhost:8083 |

### K8s Resources (Per Pod)

| | Rust | Go | Python |
|--|------|-----|--------|
| Memory request/limit | 32Mi / 128Mi | 256Mi / 512Mi | 256Mi / 768Mi |
| CPU request/limit | 100m / 500m | 100m / 500m | 100m / 500m |
| Replicas | 1 | 1 | 1 |

---

## How to Run

### Prerequisites
- Minikube running (`minikube start --driver=docker`)
- K6 installed
- Docker CLI connected to Minikube

### 1. Build Rust Image
```bash
# Connect Docker to Minikube
eval $(minikube docker-env)
# PowerShell: minikube docker-env | Invoke-Expression

# Build
docker build -t rust-fib-server:latest rust-app/
```

### 2. Deploy All Apps
```bash
kubectl apply -f deploy-all.yaml
kubectl get pods -w   # wait for all Running
```

### 3. Port-Forward (one terminal per app)
```bash
kubectl port-forward svc/rust-app-svc 8085:8090
kubectl port-forward svc/go-app-svc 8082:8080
kubectl port-forward svc/python-app-svc 8083:5000
```

### 4. Verify
```bash
curl localhost:8085   # Rust  → {"language":"Rust","fibonacci_30":832040,...}
curl localhost:8082   # Go    → {"language":"Go","fibonacci_30":832040,...}
curl localhost:8083   # Python → {"language":"Python (Flask)","fibonacci_30":832040,...}
```

### 5. Run K6 Load Tests
```bash
k6 run k6-test-rust.js
k6 run k6-test-go.js
k6 run k6-test-python.js
```

### 6. Check Resources During Load
```bash
kubectl top pods
```

### 7. Clean Up
```bash
kubectl delete -f deploy-all.yaml
```

---

## Files

| File | What It Is |
|------|------------|
| `deploy-all.yaml` | K8s Deployments + Services for all apps |
| `rust-app/` | Rust source, Cargo.toml, multi-stage Dockerfile |
| `k6-test-rust.js` | K6 load test for Rust |
| `k6-test-go.js` | K6 load test for Go |
| `k6-test-python.js` | K6 load test for Python |
| `RESULTS.md` | Benchmark findings and analysis |
