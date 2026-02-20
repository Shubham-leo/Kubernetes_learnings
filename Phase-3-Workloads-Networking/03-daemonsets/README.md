# 03 - DaemonSets

## What is a DaemonSet?
A DaemonSet ensures that **one copy of a pod runs on every node** in the cluster.

```
Deployment (replicas: 3):         DaemonSet:
  Node-1: Pod, Pod                  Node-1: Pod  (exactly 1)
  Node-2: Pod                       Node-2: Pod  (exactly 1)
  Node-3: (nothing)                 Node-3: Pod  (exactly 1)
  ↑ random placement                ↑ one per node, always

Add Node-4:                       Add Node-4:
  Node-4: (nothing)                 Node-4: Pod  (auto-added!)
```

## Real world examples
- **Log collector** (Fluentd, Filebeat) - collect logs from every node
- **Monitoring agent** (Prometheus node-exporter) - monitor every node
- **Network plugin** (Calico, Cilium) - networking on every node
- **Storage daemon** - storage driver on every node

## DaemonSet vs Deployment

| | Deployment | DaemonSet |
|---|-----------|-----------|
| How many pods? | You choose (replicas: N) | One per node (automatic) |
| Where do they run? | K8s decides | Every node |
| New node added? | Nothing happens | Pod auto-added |
| Use case | Apps, APIs | Agents, monitoring, logging |

## Files in this folder

| File | What it does |
|------|-------------|
| `daemonset-basic.yaml` | Runs a monitoring agent on every node |

## Commands to Run

### 1. Create the DaemonSet
```bash
kubectl apply -f daemonset-basic.yaml
kubectl get daemonsets       # or: kubectl get ds
kubectl get pods -o wide     # see which node each pod is on
```

### 2. Check details
```bash
kubectl describe ds node-monitor
# Look for "Number of Nodes Scheduled" - should match node count
```

### 3. See it's on every node
```bash
kubectl get nodes
kubectl get pods -l app=node-monitor -o wide
# Each node has exactly 1 pod
```

### 4. Check the logs
```bash
kubectl logs -l app=node-monitor
# Shows node monitoring info from each pod
```

### 5. Note: minikube only has 1 node
```bash
# Since minikube has only 1 node, you'll see only 1 DaemonSet pod
# In a real cluster with 10 nodes, you'd see 10 pods automatically!
```

### 6. Clean up
```bash
kubectl delete -f daemonset-basic.yaml
```
