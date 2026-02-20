# 06 - Network Policies

## What is a Network Policy?
By default, **every pod can talk to every other pod**. Network Policies are
firewall rules that control which pods can communicate.

```
Without Network Policy:          With Network Policy:
  Pod-A ←→ Pod-B ←→ Pod-C         Pod-A → Pod-B (allowed ✅)
  Pod-A ←→ Pod-C                   Pod-A → Pod-C (blocked ❌)
  Everyone talks to everyone       Only allowed traffic gets through
```

## Real world examples
- Frontend pods can ONLY talk to backend pods
- Backend pods can ONLY talk to database pods
- Database pods accept traffic ONLY from backend
- No pod can access the internet except the proxy

## Analogy
```
Without Network Policy = Open office (everyone can walk everywhere)
With Network Policy    = Office with keycards (only authorized access)
```

## Types of rules

```
Ingress (incoming):   "Who can send traffic TO this pod?"
Egress (outgoing):    "Where can this pod send traffic TO?"
```

**Note:** Ingress here means network ingress (incoming traffic to a pod),
NOT the Ingress resource from the previous lesson!

## Files in this folder

| File | What it does |
|------|-------------|
| `setup-pods.yaml` | Creates frontend, backend, database pods |
| `deny-all.yaml` | Blocks ALL traffic to database pods |
| `allow-backend-only.yaml` | Only backend can reach database |

## Commands to Run

### 1. First, install a network plugin (Calico) on minikube
```bash
# Minikube's default network plugin doesn't support Network Policies
# Start minikube with Calico:
minikube stop
minikube start --cni=calico --driver=docker
# Wait a few minutes for Calico to be ready

kubectl get pods -n kube-system | grep calico
```

### 2. Set up test pods
```bash
kubectl apply -f setup-pods.yaml
kubectl get pods

# Verify all pods can talk to each other (before any policy):
kubectl exec frontend -- wget -qO- --timeout=3 http://backend-svc
kubectl exec frontend -- wget -qO- --timeout=3 http://database-svc
kubectl exec backend -- wget -qO- --timeout=3 http://database-svc
# All 3 should work!
```

### 3. Apply deny-all policy to database
```bash
kubectl apply -f deny-all.yaml

# Now test - NOTHING can reach database:
kubectl exec frontend -- wget -qO- --timeout=3 http://database-svc
# TIMEOUT! ❌

kubectl exec backend -- wget -qO- --timeout=3 http://database-svc
# TIMEOUT! ❌

# But frontend to backend still works:
kubectl exec frontend -- wget -qO- --timeout=3 http://backend-svc
# Works! ✅ (policy only applies to database pods)
```

### 4. Allow only backend to reach database
```bash
kubectl apply -f allow-backend-only.yaml

# Backend can reach database now:
kubectl exec backend -- wget -qO- --timeout=3 http://database-svc
# Works! ✅

# Frontend still blocked:
kubectl exec frontend -- wget -qO- --timeout=3 http://database-svc
# TIMEOUT! ❌
```

### 5. Clean up
```bash
kubectl delete -f allow-backend-only.yaml
kubectl delete -f deny-all.yaml
kubectl delete -f setup-pods.yaml
```
