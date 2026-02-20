# 03 - Deployments

## What is a Deployment?
A Deployment is the **most common way to run apps** in Kubernetes.
It manages ReplicaSets and adds rolling updates + rollback.

```
Deployment
  └── ReplicaSet (managed automatically)
        ├── Pod-1
        ├── Pod-2
        └── Pod-3
```

## Pod vs ReplicaSet vs Deployment

Think of it like an army:
- **Pod** = one soldier. Dies? Gone forever.
- **ReplicaSet** = a squad that replaces fallen soldiers. Always keeps N alive.
- **Deployment** = a commander that can also swap old soldiers for upgraded ones.

| kind | Self-healing | Rolling updates | Use in production? |
|------|-------------|----------------|-------------------|
| Pod | No | No | Never alone |
| ReplicaSet | Yes | No | Rarely directly |
| Deployment | Yes | Yes | **Yes, always** |

## Why Deployment > ReplicaSet?
- Rolling updates (zero-downtime deployments)
- Rollback to previous version
- Pause/resume updates
- Deployment history

## Commands to Run

### 1. Create a Deployment
```bash
kubectl apply -f deployment-basic.yaml
```

### 2. Check everything it created
```bash
kubectl get deployments        # or: kubectl get deploy
kubectl get replicasets        # deployment created a RS for you
kubectl get pods               # RS created 3 pods
```

### 3. See deployment details
```bash
kubectl describe deployment nginx-deployment
```

### 4. Rolling Update (change nginx version)
```bash
kubectl apply -f deployment-update.yaml
# OR do it with a command:
# kubectl set image deployment/nginx-deployment nginx=nginx:1.25

# Watch the rollout happen in real-time:
kubectl rollout status deployment/nginx-deployment
kubectl get pods    # notice pods being replaced one by one
```

### 5. Check rollout history
```bash
kubectl rollout history deployment/nginx-deployment
```

### 6. Rollback!
```bash
kubectl rollout undo deployment/nginx-deployment
kubectl get pods    # pods rolling back to previous version
kubectl rollout status deployment/nginx-deployment
```

### 7. Scale
```bash
kubectl scale deployment nginx-deployment --replicas=5
kubectl get pods
```

### 8. Clean up
```bash
kubectl delete -f deployment-basic.yaml
```
