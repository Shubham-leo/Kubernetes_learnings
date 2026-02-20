# 04 - Services

## What is a Service?
Pods get random IPs that change when they restart. A Service gives your pods a
**stable endpoint** (fixed IP + DNS name) so other things can find them.

```
                    Service (stable IP)
                    ClusterIP: 10.96.0.1
                         |
            +------------+------------+
            |            |            |
         Pod-1        Pod-2        Pod-3
       (10.0.0.5)   (10.0.0.6)   (10.0.0.7)
                    ↑ these IPs change
```

## 3 Main Types of Services

| Type | Who can access | Use case |
|------|---------------|----------|
| **ClusterIP** | Only inside cluster | Default. Pod-to-pod communication |
| **NodePort** | Outside cluster via node IP:port | Dev/testing. Opens a port (30000-32767) on every node |
| **LoadBalancer** | Outside via cloud load balancer | Production (AWS/GCP/Azure) |

## ClusterIP vs NodePort Explained

### ClusterIP - Internal phone line
```
Inside the cluster:

  Pod-A (frontend) → http://nginx-clusterip:80 → Service → Pod-1 (backend)
                                                          → Pod-2
                                                          → Pod-3

  Outside (your browser) → ❌ CANNOT reach it
```
- Only pods inside the cluster can use it
- Gets a stable internal IP + DNS name
- Real use: frontend pod calling backend pod, app calling database

### NodePort - Opens a door to the outside
```
Your Browser → http://localhost:30080 → Node → Service → Pod-1
                                                        → Pod-2
                                                        → Pod-3

Other pods   → http://nginx-nodeport:80 → same thing (also works internally)
```
- Opens a specific port (30080) on the machine
- Anyone outside can access it
- Real use: dev/testing, quick access from browser

### Simple analogy
```
ClusterIP  = office internal phone (extension 101)
             → only employees can call it

NodePort   = office public phone number
             → anyone from outside can call it
```

### When to use what?
| Type | When to use |
|------|------------|
| ClusterIP | Always, for internal communication |
| NodePort | Dev/testing only |
| LoadBalancer | Production (uses cloud provider's load balancer) |

---

## Commands to Run

### 1. First, create a Deployment (we need pods to expose)
```bash
kubectl apply -f ../03-deployments/deployment-basic.yaml
```

### 2. Create a ClusterIP Service
```bash
kubectl apply -f service-clusterip.yaml
kubectl get services    # or: kubectl get svc
```

### 3. Test ClusterIP from inside the cluster
```bash
# ClusterIP is only accessible from within the cluster
# Use port-forward to test from your machine:
kubectl port-forward service/nginx-clusterip 8080:80
# Open http://localhost:8080
```

### 4. Create a NodePort Service
```bash
kubectl apply -f service-nodeport.yaml
kubectl get svc

# With minikube, get the URL:
minikube service nginx-nodeport --url
# Open that URL in your browser!
```

### 5. See endpoints (which pods the service routes to)
```bash
kubectl get endpoints nginx-clusterip
kubectl describe svc nginx-clusterip
```

### 6. Clean up
```bash
kubectl delete -f service-clusterip.yaml
kubectl delete -f service-nodeport.yaml
kubectl delete -f ../03-deployments/deployment-basic.yaml
```
