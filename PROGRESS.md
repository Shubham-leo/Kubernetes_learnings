# Kubernetes Learning Progress

## Phase 1 - Core Objects ✅
- **Pods** - Smallest deployable unit, runs containers
- **ReplicaSets** - Self-healing, keeps N pods alive
- **Deployments** - Rolling updates, rollback (used 99% of the time)
- **Services** - ClusterIP (internal), NodePort (external)
- **Namespaces** - Logical isolation within cluster

## Phase 2 - Configuration & Storage ✅
- **ConfigMaps** - Pass non-sensitive config to pods (env vars or files)
- **Secrets** - Pass sensitive data like passwords (base64 encoded)
- **Volumes** - Persistent storage (PV/PVC survives pod deletion)
- **Probes** - Health checks (liveness, readiness, startup)
- **Resource Limits** - CPU/memory limits (OOMKilled if exceeded)

## Phase 3 - Workloads & Networking ✅
- **Jobs** - Run a task once and finish
- **CronJobs** - Run tasks on a schedule
- **DaemonSets** - Run one pod on every node
- **StatefulSets** - For databases (stable names, ordered startup)
- **Ingress** - Domain-based routing (myapp.com → service)
- **Network Policies** - Firewall rules between pods (skipped)

## Phase 4 - Helm & Kustomize ⏭️ Skipped
- Created but not tested

## Phase 5 - K8s Architecture ✅
- Master Node (API Server, Scheduler, Controller Manager, etcd)
- Worker Node (Kubelet, Kube-proxy, Container Runtime)
- Full flow of kubectl apply explained

## Phase 6 - Demo Project 🔜
- MongoDB + MongoExpress (created, not tested)
- Ties everything together (Secret, ConfigMap, Deployment, Service)

---

## Nana's Course Mapping

| Nana's Topic | Status | Phase |
|--------------|--------|-------|
| What is Kubernetes | ✅ | Phase 5 |
| Main Components | ✅ | Phase 1 + 2 |
| K8s Architecture | ✅ | Phase 5 |
| Minikube & kubectl setup | ✅ | Phase 1 |
| Main kubectl commands | ✅ | All phases |
| YAML Configuration files | ✅ | All phases |
| Namespaces | ✅ | Phase 1 |
| Ingress | ✅ | Phase 3 |
| Volumes | ✅ | Phase 2 |
| StatefulSets | ✅ | Phase 3 |
| Services | ✅ | Phase 1 |
| Helm | ⏭️ | Phase 4 |

## Setup
- **Tools**: kubectl, Docker Desktop, Minikube
- **Cluster**: minikube (docker driver)
- **MINIKUBE_HOME**: T:\Kubernetes tests\.minikube-data
- **Repo**: https://github.com/Shubham-leo/Kubernetes_learnings
