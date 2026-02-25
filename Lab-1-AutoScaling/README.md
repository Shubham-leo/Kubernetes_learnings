# Lab 1: Auto-Scaling in Kubernetes

## What is Auto-Scaling?
Instead of manually running `kubectl scale`, K8s can **automatically** add or remove pods based on load.

```
Low traffic:                    High traffic (auto-scaled):
  Pod-1                           Pod-1
                                  Pod-2  ← auto-created!
                                  Pod-3  ← auto-created!
                                  Pod-4  ← auto-created!

Traffic drops → K8s removes extra pods automatically
```

## 2 Types of Auto-Scaling

### HPA - Horizontal Pod Autoscaler (add more pods)
```
CPU > 50% → add more pods
CPU < 50% → remove extra pods

Pod-1 (80% CPU)  →  Pod-1 (40% CPU)
                     Pod-2 (40% CPU)  ← HPA added this
```

### VPA - Vertical Pod Autoscaler (give pod more resources)
```
HPA (horizontal):   1 pod → 2 pods → 4 pods (add more)
VPA (vertical):     1 pod (128MB) → 1 pod (256MB) → 1 pod (512MB) (make it bigger)
```

### How VPA works
```
Step 1: VPA watches pod resource usage
  Pod using 450m CPU, but requests 200m → needs more

Step 2: VPA recommends or auto-adjusts
  "This pod should have requests: cpu=300m, memory=256Mi"

Step 3: VPA restarts pod with new resource values
  Old: cpu request 200m → New: cpu request 300m
  ⚠️ Pod gets RESTARTED (brief downtime!)
```

### Why VPA is less common
```
❌ Restarts the pod to apply changes (downtime)
❌ Can't work together with HPA on same metric (CPU)
❌ Not installed by default (needs separate install)

✅ HPA is simpler and works without restarts
✅ Most apps prefer scaling horizontally
```

### When to use what?
```
HPA → 90% of apps (web servers, APIs, stateless apps)
VPA → 10% of apps (databases, single-instance apps)
```

### How load-generator works
```yaml
while true; do wget http://php-apache-svc > /dev/null; done
# Sends THOUSANDS of requests per second in an infinite loop

load-generator pod                    php-apache pod
┌──────────────┐                     ┌──────────────┐
│ wget! wget!  │ ──── request ────→  │ CPU: 181%! 🔥│
│ wget! wget!  │ ──── request ────→  │              │
└──────────────┘                     └──────────────┘
                                            ↓
                                     HPA: "181% > 50%!"
                                     HPA: "Scale to 4 pods!"
```

## The Problem
```
You deploy with replicas: 2
Normal day:   2 pods handle traffic fine ✅
Black Friday: 2 pods can't handle 10x traffic ❌

Manual fix:   kubectl scale --replicas=20
              But you're asleep at 3 AM!
```

## How HPA works step by step

```
Step 1: You set a rule
  "If CPU goes above 50%, add more pods"
  "Minimum 1 pod, maximum 10 pods"

Step 2: HPA watches metrics every 15 seconds
  metrics-server → tells HPA → "pod CPU is 30%"
  HPA: "30% < 50%, do nothing" 😴

Step 3: Traffic spikes!
  metrics-server → "pod CPU is 85%!"
  HPA: "85% > 50%! Need more pods!"

Step 4: HPA calculates how many pods needed
  Current: 2 pods at 85% CPU
  Target:  50% CPU
  Formula: 2 × (85/50) = 3.4 → round up = 4 pods
  HPA: "Scale to 4 pods!"

Step 5: Traffic drops
  metrics-server → "pod CPU is 20%"
  HPA: "20% < 50%, too many pods"
  HPA waits 5 minutes (cool-down) → scales back to 1 pod
```

### Visually
```
Traffic low:          Traffic high:         Traffic drops:
┌──────┐              ┌──────┐              ┌──────┐
│Pod-1 │ CPU: 20%     │Pod-1 │ CPU: 45%     │Pod-1 │ CPU: 15%
└──────┘              │Pod-2 │ CPU: 45%     └──────┘
                      │Pod-3 │ CPU: 45%
HPA: "1 pod enough"  │Pod-4 │ CPU: 45%     HPA: "back to 1"
                      └──────┘
                      HPA: "scaled to 4!"
```

## What HPA needs to work
```
1. metrics-server       → installed in cluster (provides CPU/memory data)
2. resource requests    → set in your deployment YAML (cpu: "200m")
3. HPA resource         → the autoscaler rule (target: 50% CPU)
```

Without `resource requests` in your deployment, HPA can't calculate percentages:
```yaml
resources:
  requests:
    cpu: "200m"    ← HPA uses this as 100% baseline
```
If pod uses 100m CPU → that's 50% of 200m → right at the threshold.

## Prerequisites
```bash
# Enable metrics-server (HPA needs CPU/memory metrics)
minikube addons enable metrics-server

# Wait 1 minute, then verify:
kubectl top nodes
kubectl top pods
```

## Files in this folder

| File | What it does |
|------|-------------|
| `app-deployment.yaml` | Simple app with resource requests |
| `hpa.yaml` | HPA - scales based on CPU usage |
| `load-generator.yaml` | Pod that generates CPU load for testing |

## Commands to Run

### 1. Enable metrics-server
```bash
minikube addons enable metrics-server
# Wait 1-2 minutes
kubectl top nodes
```

### 2. Deploy the app
```bash
kubectl apply -f app-deployment.yaml
kubectl get pods
```

### 3. Create the HPA
```bash
kubectl apply -f hpa.yaml
kubectl get hpa
# Shows: TARGETS 0%/50%  MINPODS 1  MAXPODS 10  REPLICAS 1
```

### 4. Watch HPA in action (open a second terminal for this)
```bash
kubectl get hpa -w
```

### 5. Generate load (in your first terminal)
```bash
kubectl apply -f load-generator.yaml
# This sends continuous requests to our app
```

### 6. Watch the magic!
```bash
# In the HPA watch terminal, you'll see:
# CPU goes up → REPLICAS increases from 1 → 2 → 3 → 4...
# HPA is auto-scaling!

kubectl get pods
# More pods appearing automatically!
```

### 7. Stop the load
```bash
kubectl delete -f load-generator.yaml

# Wait 5-10 minutes (cool-down period)
# HPA will scale pods back down to 1
kubectl get hpa
kubectl get pods
```

### 8. Clean up
```bash
kubectl delete -f hpa.yaml
kubectl delete -f app-deployment.yaml
```
