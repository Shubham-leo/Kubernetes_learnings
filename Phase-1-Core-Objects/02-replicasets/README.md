# 02 - ReplicaSets

## What is a ReplicaSet?
A ReplicaSet ensures that a **specified number of pod replicas** are running at all times.

```
ReplicaSet (replicas: 3)
  ├── Pod-1 (nginx)
  ├── Pod-2 (nginx)
  └── Pod-3 (nginx)

If Pod-2 dies → ReplicaSet auto-creates Pod-2-new!
```

## Key Facts
- ReplicaSet = "keep N copies of this pod running"
- Uses **label selectors** to know which pods belong to it
- If a pod dies, it creates a new one automatically (self-healing!)
- In practice, you rarely create ReplicaSets directly - Deployments manage them
- But understanding them helps you understand Deployments

## Commands to Run

### 1. Create the ReplicaSet
```bash
kubectl apply -f replicaset-basic.yaml
```

### 2. Check it
```bash
kubectl get replicasets          # or: kubectl get rs
kubectl get pods                 # you should see 3 pods!
kubectl get pods --show-labels   # all have app=rs-demo label
```

### 3. Watch self-healing in action!
```bash
# Delete one pod manually
kubectl delete pod <pod-name>    # copy a pod name from 'kubectl get pods'

# Immediately check pods again
kubectl get pods                 # a new pod was created to replace it!
```

### 4. Scale up/down
```bash
kubectl scale replicaset nginx-replicaset --replicas=5
kubectl get pods     # now 5 pods!

kubectl scale replicaset nginx-replicaset --replicas=2
kubectl get pods     # back to 2, extra pods terminated
```

### 5. Describe to see events
```bash
kubectl describe rs nginx-replicaset
```

### 6. Clean up
```bash
kubectl delete -f replicaset-basic.yaml
```
