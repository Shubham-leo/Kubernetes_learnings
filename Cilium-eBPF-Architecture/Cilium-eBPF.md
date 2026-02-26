# eBPF / Cilium — Where It Fits

If Kubernetes solves *"how do I scale and self-heal services?"* —
Cilium solves *"how do I control and observe traffic between those services at kernel level?"*

---

## What is eBPF?

Small programs that run **inside the Linux kernel**. The kernel verifies them for safety, then JIT-compiles them to run at near-native speed.

Think of it like this:

```
Normal networking:
  Packet arrives → kernel → iptables rules (hundreds of them) → your app

eBPF networking:
  Packet arrives → kernel runs your eBPF program (one hash lookup) → your app
```

eBPF hooks into specific kernel attach points (XDP, TC ingress/egress, socket-level). In a Kubernetes context with Cilium, this covers ALL pod-to-pod and pod-to-service traffic. No code changes needed in our app.

## What is Cilium?

Kubernetes CNI (networking plugin) built on eBPF. It **replaces** three things:

1. **kube-proxy** — no more iptables chain walking for service routing
2. **iptables-based packet filtering** — eBPF hash map lookups instead
3. **Separate NetworkPolicy engine** — no need for Calico or Azure Network Policy Manager

In AKS, Cilium is native. One flag at cluster creation. kube-proxy does NOT run alongside — it's a **full replacement**.

---

## Where Cilium Sits

```
Your Pods (agent-node, monitor, landing-page)
   │
   ▼
Kubernetes Service
   │
   ▼
Cilium CNI ◄── THIS layer. Replaces kube-proxy + iptables
   │
   ▼
Linux Kernel (eBPF programs)
   │
   ▼
Azure VNet
```

---

## In Our Architecture

We have: `agent-node`, `monitor-node`, `landing-page`, Azure OpenAI, PostgreSQL, Redis, LiveKit.

Here's what Cilium actually does for each.

---

### 1. Network Isolation — Biggest Immediate Win

**Without Cilium:** All pods can talk to everything by default.

**With Cilium:** Rules enforced in the kernel, before the packet reaches userspace.

```
What we define:                           What happens at kernel level:

agent-node  → Azure OpenAI    ALLOW       eBPF program checks source identity
agent-node  → PostgreSQL       DENY        against policy map. Match = forward.
monitor     → PostgreSQL       ALLOW       No match = DROP. Packet never leaves
landing-page → Redis           DENY        kernel space. Never reaches the app.
landing-page → Azure OpenAI    DENY
* → Azure metadata endpoint    DENY
```

**Real scenario:** If `landing-page` gets compromised, attacker runs `curl postgres:5432`. The eBPF program in the kernel checks the source pod identity, finds no matching allow rule, and drops the packet. It never reaches PostgreSQL. No app-level firewall needed.

This is **blast-radius reduction at kernel level**.

> **AKS bonus:** With ACNS (Advanced Container Networking Services) enabled, we also get **L7 policies** — filter by HTTP method and path. Example: allow GET `/api/rooms` but deny DELETE `/api/rooms`. And **FQDN policies** — allow `agent-node` to reach `openai.azure.com` but nothing else on the internet.

---

### 2. Hubble — Live Network Observability

Cilium ships with **Hubble** — a network observability layer that sees every flow in the cluster.

What Hubble gives us:

| Metric | How It Helps |
|---|---|
| Pod-to-pod connection map | See exactly who talks to who, live |
| TCP flow counts | Track connection volume per service |
| DNS query logs | See every DNS lookup our pods make |
| Dropped packets + reason | Know WHY a connection failed, not just that it failed |
| Latency per service | Spot slow backends (requires L7 visibility) |
| HTTP status codes (4xx/5xx) | Monitor error rates (requires L7 visibility) |

**What this means for us:**

Our `monitor-node` tracks business metrics — rooms, users, call quality.
Hubble tracks infrastructure metrics — connection flows, DNS, packet drops, latency.

They complement each other. Right now if a K6 test fails, we're guessing where traffic dropped. Hubble shows us exactly.

```
Example Hubble output:

agent-node-7f8b9 → openai.azure.com:443    200 OK     23ms
agent-node-7f8b9 → redis-service:6379       TCP SYN    DROPPED (policy denied)
monitor-node-4a2  → postgres-service:5432   TCP EST    2ms
landing-page-9c1  → agent-service:8080      TCP SYN    DROPPED (policy denied)
```

> **AKS note:** Basic Hubble flow logs are available with base Cilium. Metrics dashboards, L7 visibility (HTTP status codes, latency histograms), and flow log export require **ACNS** — this is a paid add-on enabled via `--enable-acns`.

---

### 3. Faster Service Routing

Normal Kubernetes with kube-proxy:

```
Service → kube-proxy → iptables chain (rule 1? no. rule 2? no. rule 3? no... rule 847? yes!) → Pod
```

iptables is O(n) — every packet walks the chain linearly. With thousands of services, this gets slow.

Cilium:

```
Service → eBPF hash map lookup (O(1) average case) → Pod
```

One lookup, done. No chain walking.

**Real numbers from Microsoft:** Azure CNI powered by Cilium delivers ~30% higher throughput and ~30% lower service routing latency at 16,000 pods compared to kube-proxy.

For our real-time voice traffic, lower routing overhead = lower jitter. It's not a massive change at small scale, but it's cleaner and scales better as we add agent pods.

---

### 4. Load Balancing

Cilium provides eBPF-based load balancing with Maglev consistent hashing and socket-level rewriting (at `connect()` time, not per-packet NAT). This is more efficient than iptables-based random selection.

For our `monitor-node` WebSocket connections from admin dashboards: Cilium's load balancing distributes new TCP connections more evenly across pods. Once a WebSocket connection is established though, it stays pinned to one pod — that's just how long-lived connections work. The improvement is on **initial connection distribution**, not ongoing WebSocket routing.

---

### 5. What Cilium Does NOT Do

Be honest about the boundaries:

| Thing | Why Cilium Can't Help |
|---|---|
| LiveKit UDP media | If using `hostNetwork`, pods share the node's identity — per-pod network policies don't apply. Service routing still works, but you lose fine-grained isolation. Use node-level `CiliumClusterwideNetworkPolicy` for coarser host rules |
| Windows dev environment | eBPF is Linux-only. It only runs on cluster nodes, not our dev machines |
| Application bugs | Cilium is networking. It doesn't fix our code |
| LLM response time | Azure OpenAI latency is Azure's problem, not our network's |
| Replacing Azure Firewall | Cilium handles east-west (pod-to-pod). Azure Firewall handles north-south (internet ingress/egress). They're different layers |

---

## Should We Use It?

| Feature | Worth It? | Cost |
|---|---|---|
| kube-proxy replacement | YES — free performance upgrade | Free (base Cilium) |
| L3/L4 network policies | YES — real security with zero effort | Free (base Cilium) |
| Hubble flow logs | YES — see who talks to who | Free (base Cilium) |
| L7 policies (HTTP filtering) | YES if we need fine-grained API control | Paid (ACNS) |
| FQDN policies | YES for controlling outbound to Azure OpenAI | Paid (ACNS) |
| Hubble metrics + dashboards | NICE for production monitoring | Paid (ACNS) |

---

## When Does It Really Shine?

When you reach:

- 20+ rooms (multiple agent replicas, HPA active)
- Strict outbound control (LLM API keys must not leak)
- Need to debug "why is this connection failing" without adding logging to every service
- Compliance or security audit requires documented network isolation

That's when Cilium becomes serious infrastructure value.

---

## Bottom Line

**Enable it at AKS cluster creation.** The base features are free and native. No reason not to.

```
K8s gives you:     "Services scale and self-heal"
Cilium gives you:  "Traffic between those services is controlled and observable at kernel level"
```

Easiest path: **AKS with `--network-dataplane cilium`** + Azure managed PostgreSQL/Redis + LiveKit Cloud. Three Deployments, auto-scaling, kernel-level security, full observability.
