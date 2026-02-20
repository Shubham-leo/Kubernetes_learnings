# 05 - Resource Limits

## The Problem
Without limits, one bad pod can eat ALL the CPU/memory on a node
and crash everything else.

```
Node (4 CPU, 8GB RAM)
  ├── Pod-A: using 3.5 CPU, 7GB RAM  ← memory leak!
  ├── Pod-B: can't get resources      ← starving
  └── Pod-C: can't get resources      ← starving
```

## The Solution: Requests & Limits

| | What it does | Analogy |
|---|-------------|---------|
| **Request** | Minimum guaranteed resources | "I need at least 256MB RAM" |
| **Limit** | Maximum allowed resources | "Never use more than 512MB RAM" |

```
request ← guaranteed minimum -------- limit ← hard ceiling
  |                                      |
  256MB .............. can use up to .. 512MB → killed if exceeded!
```

### What happens when limits are exceeded?
- **CPU limit exceeded** → pod is throttled (slowed down, not killed)
- **Memory limit exceeded** → pod is OOMKilled (Out Of Memory killed)

## Files in this folder

| File | What it does |
|------|-------------|
| `pod-with-limits.yaml` | Pod with CPU and memory limits |
| `pod-oom-kill.yaml` | Pod that exceeds memory limit (gets killed!) |

## Commands to Run

### 1. Pod with resource limits
```bash
kubectl apply -f pod-with-limits.yaml
kubectl get pods

# See the resource allocation:
kubectl describe pod limited-pod
# Look for "Requests" and "Limits" in the output
```

### 2. Check node resource usage
```bash
kubectl top nodes       # requires metrics-server
kubectl top pods        # requires metrics-server

# If metrics-server isn't installed:
minikube addons enable metrics-server
# Wait 1 minute, then try again
```

### 3. Watch a pod get OOM Killed
```bash
kubectl apply -f pod-oom-kill.yaml
kubectl get pods -w

# This pod tries to allocate 200MB but only has 50MB limit
# Watch STATUS change to OOMKilled!
# K8s will keep restarting it (CrashLoopBackOff)
```

### 4. Check why it was killed
```bash
kubectl describe pod oom-pod
# Look at "Last State" → Reason: OOMKilled
```

### 5. Clean up
```bash
kubectl delete -f pod-with-limits.yaml
kubectl delete -f pod-oom-kill.yaml
```
