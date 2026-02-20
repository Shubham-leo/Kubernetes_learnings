# 03 - Volumes

## The Problem
By default, when a pod dies, **all data inside it is lost**.

```
Pod running → writes data to /app/data → Pod crashes → data GONE! 💀
```

## The Solution: Volumes
Volumes give pods **storage that survives restarts**.

```
Pod running → writes data to Volume → Pod crashes → Pod restarts → data still there! ✅
```

## Types of Volumes

| Type | Data survives | Use case |
|------|--------------|----------|
| `emptyDir` | Pod restart, NOT pod deletion | Temp storage, shared between containers |
| `hostPath` | Pod deletion (stored on node) | Dev/testing only, data on the node's disk |
| `PersistentVolume (PV)` | Everything (independent of pod) | Production - databases, uploads |

### How PV/PVC works (the important one)

```
Admin creates:          Developer requests:        Pod uses:
PersistentVolume  ←→  PersistentVolumeClaim  ←→  Volume in Pod
(the actual disk)     ("I need 1GB storage")     (mounted path)
```

Think of it like renting:
- **PV** = available apartments (storage resources)
- **PVC** = rental application ("I need a 1-bedroom")
- **Pod** = tenant that moves in

## Files in this folder

| File | What it does |
|------|-------------|
| `pod-emptydir.yaml` | Two containers sharing temp storage |
| `pod-hostpath.yaml` | Pod using node's disk |
| `pv-and-pvc.yaml` | PersistentVolume + Claim |
| `pod-with-pvc.yaml` | Pod using persistent storage |

## Commands to Run

### 1. emptyDir - Shared temp storage between containers
```bash
kubectl apply -f pod-emptydir.yaml
kubectl get pods

# The writer container writes a file, the reader container reads it:
kubectl logs emptydir-pod -c reader
```

### 2. hostPath - Store on node's disk
```bash
kubectl apply -f pod-hostpath.yaml
kubectl get pods

# Write something inside the pod:
kubectl exec hostpath-pod -- sh -c "echo 'hello from pod' > /app/data/test.txt"
kubectl exec hostpath-pod -- cat /app/data/test.txt

# Delete and recreate - data survives!
kubectl delete pod hostpath-pod
kubectl apply -f pod-hostpath.yaml
kubectl exec hostpath-pod -- cat /app/data/test.txt    # still there!
```

### 3. PersistentVolume + PVC (the production way)
```bash
# Create PV and PVC
kubectl apply -f pv-and-pvc.yaml
kubectl get pv
kubectl get pvc

# Create pod that uses the PVC
kubectl apply -f pod-with-pvc.yaml
kubectl get pods

# Write data
kubectl exec pvc-pod -- sh -c "echo 'persistent data!' > /app/data/test.txt"
kubectl exec pvc-pod -- cat /app/data/test.txt

# Delete pod, recreate, data survives!
kubectl delete pod pvc-pod
kubectl apply -f pod-with-pvc.yaml
kubectl exec pvc-pod -- cat /app/data/test.txt    # still there!
```

### 4. Clean up
```bash
kubectl delete -f pod-emptydir.yaml
kubectl delete -f pod-hostpath.yaml
kubectl delete -f pod-with-pvc.yaml
kubectl delete -f pv-and-pvc.yaml
```
