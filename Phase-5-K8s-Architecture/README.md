# Phase 5: Kubernetes Architecture

## How K8s works internally

When you run `kubectl apply -f deployment.yaml`, a LOT happens behind the scenes.
This phase explains what's inside a Kubernetes cluster.

## Cluster = Master Node(s) + Worker Node(s)

```
┌─────────────────────────────────────────────────────────┐
│                   Kubernetes Cluster                     │
│                                                          │
│  ┌──────────────────────┐   ┌────────────────────────┐  │
│  │    Master Node        │   │    Worker Node 1        │  │
│  │    (Control Plane)    │   │                          │  │
│  │                       │   │  ┌─────┐  ┌─────┐      │  │
│  │  ┌─────────────────┐ │   │  │Pod-1│  │Pod-2│      │  │
│  │  │ API Server       │ │   │  └─────┘  └─────┘      │  │
│  │  │ Scheduler        │ │   │  ┌─────┐               │  │
│  │  │ Controller Manager│ │   │  │Pod-3│               │  │
│  │  │ etcd             │ │   │  └─────┘               │  │
│  │  └─────────────────┘ │   │                          │  │
│  └──────────────────────┘   └────────────────────────┘  │
│                                                          │
│                             ┌────────────────────────┐  │
│                             │    Worker Node 2        │  │
│                             │                          │  │
│                             │  ┌─────┐  ┌─────┐      │  │
│                             │  │Pod-4│  │Pod-5│      │  │
│                             │  └─────┘  └─────┘      │  │
│                             └────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

## Master Node (Control Plane) - The Brain

The master node makes ALL the decisions. It has 4 components:

### 1. API Server - The front door
```
You (kubectl) → API Server → rest of the cluster

- EVERY request goes through API Server first
- It's the ONLY component you talk to
- Validates and authenticates requests
```

### 2. Scheduler - The matchmaker
```
New pod needs a home → Scheduler checks:
  Node-1: 80% CPU used → too busy
  Node-2: 20% CPU used → pick this one! ✅

- Decides WHICH node a pod should run on
- Considers resources, constraints, affinity rules
- Does NOT actually start the pod (kubelet does that)
```

### 3. Controller Manager - The supervisor
```
Deployment says: replicas: 3
Controller checks: only 2 pods running
Controller: "Create 1 more pod!"

- Watches the desired state vs actual state
- ReplicaSet Controller, Deployment Controller, Job Controller, etc.
- Makes sure reality matches what you asked for
```

### 4. etcd - The database
```
etcd stores EVERYTHING:
  - Cluster config
  - All resource definitions (pods, services, secrets...)
  - Current state of every object
  - It's a key-value store

⚠️ No application data! Only cluster state.
```

### How they work together
```
You: kubectl apply -f deployment.yaml (replicas: 3)

1. API Server    → receives request, validates it, stores in etcd
2. Controller    → sees "3 replicas needed, 0 exist", creates 3 pod specs
3. Scheduler     → assigns each pod to a node (Node-1, Node-2, Node-1)
4. etcd          → stores all this state
```

---

## Worker Node - The muscle

Worker nodes actually RUN your pods. Each worker has 3 components:

### 1. Kubelet - The node agent
```
Master: "Run this pod on your node"
Kubelet: "OK!" → pulls image → starts container → reports status back

- Runs on every worker node
- Takes instructions from API Server
- Manages containers on its node
- Reports health back to master
```

### 2. Kube-proxy - The network manager
```
Service: "Route traffic to pods with label app=web"
Kube-proxy: sets up network rules on the node

- Handles networking on each node
- Makes Services work (load balancing, routing)
- Maintains network rules (iptables/ipvs)
```

### 3. Container Runtime - The engine
```
Kubelet: "Start this container"
Container Runtime: pulls image, runs it

- Actually runs containers (Docker, containerd, CRI-O)
- Minikube uses Docker as runtime
- Kubernetes doesn't care which runtime, just needs CRI compatible
```

### How they work together
```
Scheduler assigns pod to Node-1:

1. Kubelet (Node-1) → gets instruction from API Server
2. Kubelet           → tells Container Runtime to pull image
3. Container Runtime → pulls nginx:1.25 image, starts container
4. Kubelet           → reports "pod is running" back to API Server
5. Kube-proxy        → updates network rules so traffic can reach the pod
```

---

## Full flow: What happens when you run kubectl apply

```
You: kubectl apply -f deployment.yaml

Step 1: kubectl → API Server
        "Create a deployment with 3 nginx pods"

Step 2: API Server → etcd
        Stores the deployment spec

Step 3: Controller Manager notices
        "Deployment needs 3 pods, 0 exist → create 3"

Step 4: Scheduler decides
        "Pod-1 → Node-1, Pod-2 → Node-2, Pod-3 → Node-1"

Step 5: Kubelet on Node-1
        Pulls nginx image, starts 2 containers
        Reports back: "2 pods running"

Step 6: Kubelet on Node-2
        Pulls nginx image, starts 1 container
        Reports back: "1 pod running"

Step 7: Kube-proxy on both nodes
        Sets up networking so Services can route to pods

Step 8: etcd updated
        Stores current state: "3/3 pods running"

Done! ✅
```

---

## What happens when a pod crashes?

```
1. Kubelet detects: "Pod-2 on Node-2 crashed!"
2. Kubelet reports to API Server
3. Controller Manager sees: "Only 2 pods, need 3"
4. Controller creates new pod spec
5. Scheduler assigns it: "New pod → Node-1"
6. Kubelet on Node-1 starts the new container
7. Back to 3 pods! Self-healing complete ✅
```

---

## What happens when a worker node dies?

```
1. Kubelet on Node-2 stops sending heartbeats
2. API Server: "Node-2 is not responding"
3. After timeout: Node-2 marked as "NotReady"
4. Controller Manager: "Pods on Node-2 are lost, recreate them"
5. Scheduler: assigns pods to healthy nodes (Node-1, Node-3)
6. Kubelets start new pods on healthy nodes
7. All pods recovered on other nodes! ✅
```

---

## Minikube = Everything on 1 machine

```
Real cluster:                    Minikube:
  Master Node (separate)           ┌─────────────────┐
  Worker Node 1                    │ Single Node      │
  Worker Node 2                    │                   │
  Worker Node 3                    │ Master + Worker   │
                                   │ all in one!       │
                                   └─────────────────┘

That's why minikube is only for learning/dev, not production.
```

---

## Quick reference

| Component | Runs on | What it does |
|-----------|---------|-------------|
| API Server | Master | Front door, handles all requests |
| Scheduler | Master | Assigns pods to nodes |
| Controller Manager | Master | Maintains desired state |
| etcd | Master | Stores all cluster data |
| Kubelet | Worker | Manages containers on the node |
| Kube-proxy | Worker | Handles networking on the node |
| Container Runtime | Worker | Actually runs containers |

## Commands to explore architecture

```bash
# See your nodes
kubectl get nodes

# See master components (running as pods in kube-system)
kubectl get pods -n kube-system

# See detailed node info (capacity, conditions, images)
kubectl describe node minikube

# See cluster info
kubectl cluster-info

# See all API resources available
kubectl api-resources
```
