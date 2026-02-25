# 04 - StatefulSets

## What is a StatefulSet?
A StatefulSet is like a Deployment but for **apps that need stable identity** - mainly databases.

## Why identity matters

```
Deployment pods = temporary workers (no name badge)
  "Hey you, serve this request" → any pod can handle it
  Pod dies → new random worker replaces it, nobody notices

StatefulSet pods = permanent employees (with name badge + own desk)
  "Hey mysql-0, you're the master" → specific identity matters
  Pod dies → same person comes back, same name, same desk (storage)
```

### Without vs With stable identity
```
Without stable identity:
  Pod-abc12 (master) → dies → Pod-xyz89 (new pod)
  Replicas: "Who is the master now? abc12 is gone!" ❌

With StatefulSet:
  mysql-0 (master) → dies → mysql-0 (same name is back!)
  Replicas: "mysql-0 is back, I know who to sync with" ✅
```

## StatefulSet gives 3 guarantees

**1. Stable names** - always web-0, web-1, web-2 (not random)
**2. Ordered startup** - starts one by one (master before replicas)
**3. Own storage per pod** - each pod gets its own private disk

### What retains what
```
Stable name     → StatefulSet guarantees this
Stable storage  → PVC (PersistentVolumeClaim) per pod
Stable DNS      → Headless Service gives each pod a unique address
```

## The Problem with Deployments for Databases
```
Deployment pods:
  nginx-deployment-5cf6467d9-abc12    ← random name
  nginx-deployment-5cf6467d9-xyz89    ← random name
  nginx-deployment-5cf6467d9-def45    ← random name
  ↑ names are random, pods are interchangeable

Database needs:
  mysql-0    ← always "mysql-0" (master)
  mysql-1    ← always "mysql-1" (replica)
  mysql-2    ← always "mysql-2" (replica)
  ↑ stable names, each has its own storage
```

## Deployment vs StatefulSet

| | Deployment | StatefulSet |
|---|-----------|------------|
| Pod names | Random (abc12, xyz89) | Ordered (app-0, app-1, app-2) |
| Startup order | All at once | One by one (0 → 1 → 2) |
| Shutdown order | Random | Reverse (2 → 1 → 0) |
| Storage | Shared or none | Each pod gets its own volume |
| Network identity | Random IP | Stable DNS name per pod |
| Use case | Stateless apps (web, API) | Databases (MySQL, Postgres, Redis) |

## Deployment replicas vs StatefulSet replicas
```
Deployment (replicas: 3):          StatefulSet (replicas: 3):
  Created ALL at once                Created ONE BY ONE in order
  nginx-abc12  ← random             web-0  ← first (waits until ready)
  nginx-xyz89  ← random             web-1  ← second (waits until web-0 ready)
  nginx-def45  ← random             web-2  ← third (waits until web-1 ready)
  All identical                      Each has its own identity + storage
```

### Each pod gets its OWN storage
```
web-0 → PVC: web-data-web-0 (100MB disk)
web-1 → PVC: web-data-web-1 (100MB disk)
web-2 → PVC: web-data-web-2 (100MB disk)
         ↑ NOT shared, each pod has private data
```

### In real database use
```
web-0 (master)   → writes go here
web-1 (replica)  → reads, copies from web-0
web-2 (replica)  → reads, copies from web-0
```

## Headless Service - Why StatefulSets need it

```
Normal Service:    gives one IP for all pods (load balanced)
Headless Service:  gives each pod its own DNS name
```

## How StatefulSet pods get DNS names
```
StatefulSet name: mysql
Headless Service: mysql-svc

Pod-0 DNS: mysql-0.mysql-svc.default.svc.cluster.local
Pod-1 DNS: mysql-1.mysql-svc.default.svc.cluster.local
Pod-2 DNS: mysql-2.mysql-svc.default.svc.cluster.local

Other pods can connect to a SPECIFIC database replica!
```

## Files in this folder

| File | What it does |
|------|-------------|
| `statefulset-basic.yaml` | StatefulSet with headless service + per-pod storage |

## Commands to Run

### 1. Create the StatefulSet
```bash
kubectl apply -f statefulset-basic.yaml
kubectl get statefulsets      # or: kubectl get sts
```

### 2. Watch pods come up IN ORDER
```bash
kubectl get pods -w
# web-0 starts first
# web-1 starts AFTER web-0 is ready
# web-2 starts AFTER web-1 is ready
# This ordered startup is the key feature!
```

### 3. Notice the stable names
```bash
kubectl get pods
# Names are web-0, web-1, web-2 (not random!)
```

### 4. Check each pod has its own storage
```bash
kubectl get pvc
# Each pod got its own PersistentVolumeClaim automatically!
# web-data-web-0, web-data-web-1, web-data-web-2
```

### 5. Write data to each pod
```bash
kubectl exec web-0 -- sh -c "echo 'I am web-0' > /usr/share/nginx/html/index.html"
kubectl exec web-1 -- sh -c "echo 'I am web-1' > /usr/share/nginx/html/index.html"

# Each pod has its OWN data (not shared)
kubectl exec web-0 -- cat /usr/share/nginx/html/index.html
kubectl exec web-1 -- cat /usr/share/nginx/html/index.html
```

### 6. Delete a pod - it comes back with the SAME name and data
```bash
kubectl delete pod web-0
kubectl get pods -w
# web-0 comes back (same name, same storage!)

kubectl exec web-0 -- cat /usr/share/nginx/html/index.html
# Data is still there!
```

### 7. Check DNS names (headless service)
```bash
kubectl run -it --rm dns-test --image=busybox --restart=Never -- nslookup web-0.nginx-headless
```

### 8. Clean up
```bash
kubectl delete -f statefulset-basic.yaml
kubectl delete pvc --all       # StatefulSet PVCs are not auto-deleted
```
