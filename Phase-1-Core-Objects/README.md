# Phase 1: Core Kubernetes Objects

## Prerequisites - Set Up Local Cluster

You have Docker and kubectl. Now install **Minikube** (easiest local cluster):

### Step 1: Install Minikube
```powershell
# Open PowerShell as Administrator and run:
winget install Kubernetes.minikube
```
OR download from: https://minikube.sigs.k8s.io/docs/start/

### Step 2: Start your cluster
```bash
minikube start --driver=docker
```

### Step 3: Verify
```bash
kubectl cluster-info
kubectl get nodes
```

You should see one node in "Ready" state. You're good to go!

---

## Learning Order

| # | Topic | What You'll Learn |
|---|-------|-------------------|
| 01 | Pods | Smallest deployable unit in K8s |
| 02 | ReplicaSets | Ensures N copies of a pod run |
| 03 | Deployments | Manages ReplicaSets + rolling updates |
| 04 | Services | Networking - expose pods to traffic |
| 05 | Namespaces | Logical isolation within a cluster |

Each folder has:
- `README.md` - Concepts explained simply
- YAML files - Ready to apply
- Commands to test and verify

Start with `01-pods/` and go in order!
