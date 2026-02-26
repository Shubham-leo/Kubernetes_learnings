# Kubernetes — Where It Fits

Your problem: everything runs on one VM. If it dies, all calls drop. Scaling means buying a bigger VM.

---

## Now vs. With Kubernetes

```
┌───────────────────┬───────────────────────────────────┬───────────────────────────────────────────────┐
│      Problem      │       Now (Docker Compose)        │                   With K8s                    │
├───────────────────┼───────────────────────────────────┼───────────────────────────────────────────────┤
│ Scaling           │ Resize VM manually                │ Auto-add agent pods per room                  │
├───────────────────┼───────────────────────────────────┼───────────────────────────────────────────────┤
│ VM crash          │ Everything dies                   │ Services restart on another node              │
├───────────────────┼───────────────────────────────────┼───────────────────────────────────────────────┤
│ Deploying updates │ Brief downtime                    │ Zero-downtime rolling updates                 │
├───────────────────┼───────────────────────────────────┼───────────────────────────────────────────────┤
│ Resource fights   │ Services compete for CPU/RAM      │ Guaranteed limits per service                 │
├───────────────────┼───────────────────────────────────┼───────────────────────────────────────────────┤
│ Secrets           │ Plain .env on disk                │ Encrypted, per-pod access via Azure Key Vault │
├───────────────────┼───────────────────────────────────┼───────────────────────────────────────────────┤
│ Network isolation │ Everything can talk to everything │ Firewall rules per service                    │
└───────────────────┴───────────────────────────────────┴───────────────────────────────────────────────┘
```

---

## Best Candidates for K8s

### agent-node — Biggest Win

One pod per room, auto-scales from 2 to 50. This is the whole point of moving to Kubernetes. HPA (Horizontal Pod Autoscaler) watches room count and spins up new agent pods automatically. When rooms close, pods scale back down. You only pay for what you use.

### landing-page — Easy Win

Stateless, trivial to scale. Put it behind a Deployment with 2-3 replicas and you get zero-downtime deploys for free.

### monitor-node — Easy Win

Stateless, same as landing-page. Replicas handle WebSocket connections from admin dashboards.

### PostgreSQL / Redis — Don't Run in K8s

Use **Azure managed services** (Azure Database for PostgreSQL, Azure Cache for Redis). You don't want to manage database failover, backups, and persistent storage inside K8s. Let Azure handle that.

### LiveKit — Hardest

Needs UDP port range 50000-60000 for media. Two options:
- `hostNetwork` DaemonSet (pin LiveKit to dedicated nodes, expose UDP directly)
- **LiveKit Cloud** (recommended — offload the hardest part entirely)

---

## Architecture With K8s

```
                        ┌─────────────────────────────────┐
                        │          AKS Cluster             │
                        │                                  │
  Users ──────────────▶ │  ┌─────────────┐  ┌──────────┐  │
                        │  │ landing-page │  │ monitor  │  │
                        │  │  (2 pods)    │  │ (2 pods) │  │
                        │  └──────┬───────┘  └────┬─────┘  │
                        │         │               │        │
                        │         ▼               │        │
                        │  ┌─────────────────┐    │        │
                        │  │   agent-node    │    │        │      ┌──────────────────┐
                        │  │  (2-50 pods)    │◄───┘        │      │  Azure Managed   │
                        │  │   HPA scaling   │             │      │                  │
                        │  └───────┬─────────┘             │      │  - PostgreSQL    │
                        │          │                       │      │  - Redis         │
                        │          ▼                       │      │  - Key Vault     │
                        │  ┌──────────────┐                │      └──────────────────┘
                        │  │   LiveKit     │                │
                        │  │ (hostNetwork  │                │      ┌──────────────────┐
                        │  │  or Cloud)    │                │      │  Azure OpenAI    │
                        │  └──────────────┘                │      └──────────────────┘
                        └─────────────────────────────────┘
```

---

## Bottom Line

**Yes, worth it when you need >15 concurrent rooms.** The agent-node auto-scaling alone justifies K8s.

Easiest path:
- AKS cluster
- Azure managed PostgreSQL + Redis
- LiveKit Cloud
- You manage three Deployments: `agent-node`, `landing-page`, `monitor-node`
- You get: auto-scaling, self-healing, rolling updates, network isolation
