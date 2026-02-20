# 01 - ConfigMaps

## What is a ConfigMap?
A ConfigMap stores **non-sensitive configuration** as key-value pairs.
Instead of hardcoding config inside your container, you pass it in from outside.

```
Without ConfigMap:                With ConfigMap:
+------------------+              +------------------+
| Container        |              | Container        |
| DB_HOST=10.0.0.5 | (hardcoded) | DB_HOST=???      | ← reads from ConfigMap
| DB_PORT=5432     |              | DB_PORT=???      |
+------------------+              +------------------+
                                          ↑
                                   ConfigMap:
                                   DB_HOST=10.0.0.5
                                   DB_PORT=5432
```

## Why use ConfigMaps?
- **Same container image, different configs** - dev vs production
- Change config WITHOUT rebuilding the image
- Keep config separate from code (12-factor app principle)

## How it works visually

### 1. ConfigMap is stored in K8s (not in any pod)
```
+--- Kubernetes Cluster Memory ---+
|                                  |
|   ConfigMap: "app-config"        |
|   ┌──────────────────────┐       |
|   │ DB_HOST  = 10.0.0.5  │       |
|   │ DB_PORT  = 5432       │       |
|   │ APP_MODE = development│       |
|   └──────────────────────┘       |
|                                  |
+----------------------------------+
```

### 2. Pod references the ConfigMap, K8s injects values at startup
```
ConfigMap                          Pod
┌──────────────────┐    inject    ┌─────────────────────────┐
│ DB_HOST =10.0.0.5│ ──────────→ │ Container               │
│ DB_PORT =5432    │ ──────────→ │                          │
│ APP_MODE=dev     │ ──────────→ │ env vars:               │
└──────────────────┘             │   DB_HOST  = 10.0.0.5   │
                                 │   DB_PORT  = 5432        │
                                 │   APP_MODE = development │
                                 │                          │
                                 │ Your app reads:          │
                                 │   process.env.DB_HOST    │
                                 │   os.getenv("DB_PORT")   │
                                 └─────────────────────────┘
```

### 3. Real world: Same image, different config per environment
```
Same Docker image → deployed to 3 environments:

┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│ Dev          │    │ Staging     │    │ Production   │
│              │    │             │    │              │
│ ConfigMap:   │    │ ConfigMap:  │    │ ConfigMap:   │
│ DB=localhost │    │ DB=staging  │    │ DB=prod-db   │
│ LOG=debug    │    │ LOG=info    │    │ LOG=error    │
└─────────────┘    └─────────────┘    └─────────────┘
      ↑                  ↑                  ↑
      └──────── same image everywhere ──────┘
```

## Two ways to use ConfigMaps

### 1. As Environment Variables
```
ConfigMap → injected as ENV vars → your app reads process.env.DB_HOST
```

### 2. As Files (mounted volume)
```
ConfigMap → mounted as a file → your app reads /etc/config/app.conf
```

## Files in this folder

| File | What it does |
|------|-------------|
| `configmap-basic.yaml` | Simple key-value ConfigMap |
| `configmap-from-file.yaml` | ConfigMap with a full config file |
| `pod-env-configmap.yaml` | Pod that reads ConfigMap as env vars |
| `pod-volume-configmap.yaml` | Pod that mounts ConfigMap as a file |

## Commands to Run

### 1. Create the ConfigMap
```bash
kubectl apply -f configmap-basic.yaml
kubectl get configmaps       # or: kubectl get cm
kubectl describe cm app-config
```

### 2. Create a pod that uses it as ENV vars
```bash
kubectl apply -f pod-env-configmap.yaml
kubectl get pods

# Check if the env vars are inside the pod:
kubectl exec configmap-env-pod -- env | grep -E "DB_HOST|DB_PORT|APP_MODE"
```

### 3. Create ConfigMap with a file
```bash
kubectl apply -f configmap-from-file.yaml
```

### 4. Create a pod that mounts it as a file
```bash
kubectl apply -f pod-volume-configmap.yaml
kubectl get pods

# Check the mounted file inside the pod:
kubectl exec configmap-volume-pod -- cat /etc/config/app.properties
```

### 5. Update ConfigMap and see changes
```bash
# Edit the ConfigMap
kubectl edit cm app-config
# Change DB_PORT to 3306, save and exit

# For volume-mounted ConfigMaps, changes reflect automatically (within ~1 min)
# For env var ConfigMaps, you need to restart the pod
```

### 6. Clean up
```bash
kubectl delete -f pod-env-configmap.yaml
kubectl delete -f pod-volume-configmap.yaml
kubectl delete -f configmap-basic.yaml
kubectl delete -f configmap-from-file.yaml
```
