# 05 - Namespaces

## What is a Namespace?
Namespaces are like **folders inside your cluster**. They let you organize
and isolate resources.

```
Cluster
  ├── namespace: default         ← your stuff goes here by default
  ├── namespace: kube-system     ← K8s internal components
  ├── namespace: dev             ← you can create custom ones
  └── namespace: production      ← for logical separation
```

## Key Facts
- Resources in different namespaces are isolated by default
- Names must be unique WITHIN a namespace, not across
- Some resources are cluster-wide (Nodes, PersistentVolumes) - not namespaced
- Default namespace is "default" if you don't specify one

## How it works - Step by step

### 1. Create a namespace (namespace-dev.yaml)
```yaml
kind: Namespace
metadata:
  name: dev          ← creates a new "folder" called dev
```

### 2. Deploy into that namespace (deployment-in-namespace.yaml)
```yaml
kind: Deployment
metadata:
  name: nginx-in-dev
  namespace: dev     ← this one line puts it in dev, not default
```

### 3. Why `kubectl get pods` shows nothing
```
kubectl get pods           → looks in "default" namespace
                             your pods are NOT here

kubectl get pods -n dev    → looks in "dev" namespace
                             your pods ARE here ✅
```

It's like folders on your computer:
```
C:\Users\default\          ← kubectl get pods (looks here)
   (empty)

C:\Users\dev\              ← kubectl get pods -n dev (looks here)
   nginx-pod-1
   nginx-pod-2
```

### Why is this useful?
```
Cluster
  ├── namespace: dev         ← developers test here
  ├── namespace: staging     ← QA tests here
  └── namespace: production  ← real users here

Same app, same cluster, but completely isolated!
```
Each team/environment gets its own namespace. They can't accidentally mess with each other.

---

## Commands to Run

### 1. See existing namespaces
```bash
kubectl get namespaces    # or: kubectl get ns
```

### 2. See what's running in kube-system (K8s internals)
```bash
kubectl get pods -n kube-system
```

### 3. Create a namespace
```bash
kubectl apply -f namespace-dev.yaml
kubectl get ns
```

### 4. Deploy something into the new namespace
```bash
kubectl apply -f deployment-in-namespace.yaml
kubectl get pods                  # nothing in default namespace
kubectl get pods -n dev           # there it is!
```

### 5. See resources across ALL namespaces
```bash
kubectl get pods --all-namespaces    # or: kubectl get pods -A
```

### 6. Set a default namespace (so you don't type -n every time)
```bash
kubectl config set-context --current --namespace=dev
kubectl get pods     # now shows dev namespace by default

# Switch back to default
kubectl config set-context --current --namespace=default
```

### 7. Clean up
```bash
kubectl delete namespace dev
# This deletes the namespace AND everything inside it!
```
