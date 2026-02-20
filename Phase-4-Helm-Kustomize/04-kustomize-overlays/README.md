# 04 - Kustomize Overlays

## What are Overlays?
Overlays let you use the **same base YAML** but customize it per environment.
This is the real power of Kustomize.

```
       base/ (shared)
      /     \
     /       \
overlays/    overlays/
  dev/         prod/
  1 replica    5 replicas
  debug logs   error logs
  ClusterIP    LoadBalancer
```

## Folder structure
```
base/                        ← shared YAML (never edited per environment)
  ├── deployment.yaml
  ├── service.yaml
  └── kustomization.yaml

overlays/
  ├── dev/                   ← dev-specific changes
  │   ├── kustomization.yaml
  │   └── replica-patch.yaml
  └── prod/                  ← prod-specific changes
      ├── kustomization.yaml
      └── replica-patch.yaml
```

## How patches work
```
Base deployment:            Dev patch:              Result (dev):
  replicas: 2          +     replicas: 1        =    replicas: 1
  image: nginx:1.25          (only changes this)     image: nginx:1.25

Base deployment:            Prod patch:             Result (prod):
  replicas: 2          +     replicas: 5        =    replicas: 5
  image: nginx:1.25          (only changes this)     image: nginx:1.25
```

## Commands to Run

### 1. Preview dev environment
```bash
kubectl kustomize overlays/dev/
# Shows final YAML for dev (1 replica, dev namespace)
```

### 2. Preview prod environment
```bash
kubectl kustomize overlays/prod/
# Shows final YAML for prod (5 replicas, prod namespace)
```

### 3. Deploy dev
```bash
kubectl create namespace dev
kubectl apply -k overlays/dev/
kubectl get all -n dev
# 1 replica!
```

### 4. Deploy prod
```bash
kubectl create namespace prod
kubectl apply -k overlays/prod/
kubectl get all -n prod
# 5 replicas!
```

### 5. Compare both
```bash
kubectl get deployments -A
# dev:  1 replica
# prod: 5 replicas
# Same base, different overlays!
```

### 6. Clean up
```bash
kubectl delete -k overlays/dev/
kubectl delete -k overlays/prod/
kubectl delete namespace dev prod
```
