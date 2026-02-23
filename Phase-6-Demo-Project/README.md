# Phase 6: Demo Project - MongoDB + Mongo Express

## What we're building

A complete app with a database and a web UI to manage it.
This ties together EVERYTHING you learned in Phases 1-5.

```
┌──────────────────────────────────────────────────┐
│                  Kubernetes Cluster                │
│                                                    │
│  ┌─────────────┐         ┌─────────────────────┐ │
│  │ Mongo       │  :8081  │ MongoDB              │ │
│  │ Express     │────────→│                      │ │
│  │ (Web UI)    │         │ (Database)           │ │
│  └──────┬──────┘         └──────────────────────┘ │
│         │                         ↑                │
│         │                    Internal Service      │
│    External Service          (ClusterIP)           │
│    (NodePort :30000)                               │
│         │                                          │
└─────────┼──────────────────────────────────────────┘
          │
     Your Browser
     http://localhost:30000
```

## What K8s resources we'll use

```
1. Secret              → stores MongoDB username/password
2. ConfigMap           → stores MongoDB server URL
3. Deployment (MongoDB)→ runs the database
4. Service (MongoDB)   → internal access (ClusterIP)
5. Deployment (Express)→ runs the web UI
6. Service (Express)   → external access (NodePort)
```

## How everything connects

```
Secret:
  mongo-root-username: admin
  mongo-root-password: password123

ConfigMap:
  mongo-url: mongodb-service    ← DNS name of MongoDB service

MongoDB Deployment:
  reads credentials from → Secret
  exposed internally via → MongoDB Service (ClusterIP)

Mongo Express Deployment:
  reads credentials from → Secret
  reads DB URL from      → ConfigMap
  connects to            → MongoDB Service
  exposed externally via → Express Service (NodePort :30000)
```

## Files in this folder (apply in this order!)

| # | File | What it creates |
|---|------|----------------|
| 1 | `mongo-secret.yaml` | Secret with DB credentials |
| 2 | `mongo-configmap.yaml` | ConfigMap with DB server URL |
| 3 | `mongo-deployment.yaml` | MongoDB Deployment + internal Service |
| 4 | `mongo-express-deployment.yaml` | Mongo Express Deployment + external Service |

## Commands to Run

### 1. Create the Secret first (credentials)
```bash
kubectl apply -f mongo-secret.yaml
kubectl get secret mongo-secret
```

### 2. Create the ConfigMap (DB URL)
```bash
kubectl apply -f mongo-configmap.yaml
kubectl get cm mongo-configmap
```

### 3. Deploy MongoDB
```bash
kubectl apply -f mongo-deployment.yaml
kubectl get pods
kubectl get svc
# Wait until MongoDB pod is Running
```

### 4. Deploy Mongo Express
```bash
kubectl apply -f mongo-express-deployment.yaml
kubectl get pods
kubectl get svc
# Wait until Express pod is Running
```

### 5. Access the Web UI
```bash
# Option 1: minikube service (opens browser automatically)
minikube service mongo-express-service

# Option 2: port-forward
kubectl port-forward service/mongo-express-service 8081:8081
# Open http://localhost:8081
# Login: admin / pass (configured in the deployment)
```

### 6. Test it!
```
1. Open http://localhost:8081
2. You'll see Mongo Express UI
3. Create a new database
4. Add some data
5. Your data is stored in MongoDB!
```

### 7. See how everything connects
```bash
kubectl get all
kubectl get secret
kubectl get cm
# You'll see all 6 resources working together!
```

### 8. Clean up
```bash
kubectl delete -f mongo-express-deployment.yaml
kubectl delete -f mongo-deployment.yaml
kubectl delete -f mongo-configmap.yaml
kubectl delete -f mongo-secret.yaml
```
