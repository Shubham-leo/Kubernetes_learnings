# Phase 3: Advanced Workloads & Networking

## What you'll learn
In Phase 1 & 2 you deployed apps with Deployments. Now you'll learn **other ways to run workloads**
and **advanced networking**.

## Learning Order

| # | Topic | What You'll Learn |
|---|-------|-------------------|
| 01 | Jobs | Run a task once and finish (like a script) |
| 02 | CronJobs | Run tasks on a schedule (like cron) |
| 03 | DaemonSets | Run one pod on EVERY node (like an agent) |
| 04 | StatefulSets | For databases - stable names, ordered startup |
| 05 | Ingress | Expose apps with domain names + routing |
| 06 | Network Policies | Firewall rules between pods |

## When to use what?

```
"I want to run a web app"          → Deployment (Phase 1)
"I want to run a one-time task"    → Job
"I want to run a scheduled task"   → CronJob
"I want to run on every node"      → DaemonSet
"I want to run a database"         → StatefulSet
"I want domain-based routing"      → Ingress
"I want to restrict pod traffic"   → Network Policy
```

## Prerequisites
- Phase 1 & 2 completed
- Minikube running (`minikube status`)

Start with `01-jobs/` and go in order!
