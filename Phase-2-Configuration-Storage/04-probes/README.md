# 04 - Probes (Health Checks)

## The Problem
How does Kubernetes know if your app is **actually working**?
A container can be "Running" but the app inside could be crashed or stuck.

```
Pod status: Running ✅
App inside: completely frozen 💀
Users:      getting errors 😡
```

## The Solution: Probes
Probes are **health checks** that K8s runs on your containers.

## 3 Types of Probes

### 1. Liveness Probe - "Is the app alive?"
```
K8s checks → App responds? → Yes → do nothing
                            → No  → RESTART the container
```
Use when: app might freeze or deadlock

### 2. Readiness Probe - "Is the app ready to serve traffic?"
```
K8s checks → App ready? → Yes → send traffic to it
                         → No  → STOP sending traffic (but don't restart)
```
Use when: app needs warmup time (loading cache, connecting to DB)

### 3. Startup Probe - "Has the app finished starting?"
```
K8s checks → App started? → Yes → hand off to liveness/readiness probes
                           → No  → keep waiting (don't restart yet)
```
Use when: app takes a long time to start

### Simple analogy
```
Startup Probe   = "Are you awake yet?"    (checked once at boot)
Readiness Probe = "Are you ready to work?" (checked continuously)
Liveness Probe  = "Are you still alive?"   (checked continuously)
```

## Files in this folder

| File | What it does |
|------|-------------|
| `pod-liveness.yaml` | Pod with liveness probe (auto-restarts if unhealthy) |
| `pod-readiness.yaml` | Pod with readiness probe (removed from service if not ready) |
| `pod-all-probes.yaml` | Pod with all 3 probes |

## Commands to Run

### 1. Liveness Probe - watch it restart an unhealthy pod
```bash
kubectl apply -f pod-liveness.yaml
kubectl get pods -w    # watch mode - see RESTARTS column increase

# This pod deliberately becomes unhealthy after 30 seconds
# K8s detects it and restarts the container automatically!
# Wait ~1 min and watch RESTARTS go from 0 → 1 → 2...
# Press Ctrl+C to stop watching
```

### 2. Check the events
```bash
kubectl describe pod liveness-pod
# Look at Events section - you'll see "Liveness probe failed" and "Restarting"
```

### 3. Readiness Probe
```bash
kubectl apply -f pod-readiness.yaml
kubectl get pods -w

# This pod isn't "ready" for the first 15 seconds
# READY column shows 0/1 initially, then 1/1
# If this was behind a Service, no traffic would go to it until ready
```

### 4. All probes together
```bash
kubectl apply -f pod-all-probes.yaml
kubectl describe pod all-probes-pod
# Look at the Conditions and Events sections
```

### 5. Clean up
```bash
kubectl delete -f pod-liveness.yaml
kubectl delete -f pod-readiness.yaml
kubectl delete -f pod-all-probes.yaml
```
