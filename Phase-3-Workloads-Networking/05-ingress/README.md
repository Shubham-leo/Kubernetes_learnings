# 05 - Ingress

## What is Ingress?
Ingress manages **external HTTP/HTTPS access** to your services.
It's like a smart reverse proxy that routes traffic based on domain names and paths.

## The Problem
```
Without Ingress (NodePort):
  http://node-ip:30080  → Service A
  http://node-ip:30081  → Service B
  http://node-ip:30082  → Service C
  ↑ ugly ports, hard to remember

With Ingress:
  http://myapp.com       → Service A
  http://myapp.com/api   → Service B
  http://blog.myapp.com  → Service C
  ↑ clean URLs, domain-based routing!
```

## How Ingress works
```
Internet → Ingress Controller (nginx) → Ingress Rules → Services → Pods
                                              |
                                    ┌─────────┴─────────┐
                                    │ myapp.com/         │ → frontend-svc
                                    │ myapp.com/api      │ → backend-svc
                                    │ blog.myapp.com     │ → blog-svc
                                    └───────────────────┘
```

## Key concepts
- **Ingress Controller**: The actual proxy (nginx, traefik). Must be installed separately!
- **Ingress Resource**: Your routing rules (YAML)
- Think: Ingress Controller = the traffic cop, Ingress Resource = the rules they follow

## Files in this folder

| File | What it does |
|------|-------------|
| `app-deployments.yaml` | Two apps (frontend + backend) with services |
| `ingress-path-based.yaml` | Route by path (/frontend, /backend) |
| `ingress-host-based.yaml` | Route by domain name |

## Commands to Run

### 1. Enable Ingress on Minikube
```bash
minikube addons enable ingress
# Wait 1 minute for the ingress controller to start

kubectl get pods -n ingress-nginx
# Wait until the controller pod is Running
```

### 2. Deploy two apps
```bash
kubectl apply -f app-deployments.yaml
kubectl get pods
kubectl get svc
```

### 3. Path-based routing
```bash
kubectl apply -f ingress-path-based.yaml
kubectl get ingress

# Get minikube IP
minikube ip

# Test the routes (replace <minikube-ip> with actual IP):
curl http://<minikube-ip>/frontend
curl http://<minikube-ip>/backend

# Or add to hosts file and use domain name (see step 5)
```

### 4. Host-based routing
```bash
kubectl apply -f ingress-host-based.yaml
kubectl get ingress
```

### 5. Set up local DNS (to test domain names)
```bash
# Get minikube IP
minikube ip

# Add to your hosts file (run as Administrator):
# Windows: C:\Windows\System32\drivers\etc\hosts
# Add this line (replace IP):
# <minikube-ip>  myapp.local api.local

# Then test:
curl http://myapp.local
curl http://api.local
```

### 6. Clean up
```bash
kubectl delete -f ingress-path-based.yaml
kubectl delete -f ingress-host-based.yaml
kubectl delete -f app-deployments.yaml
```
