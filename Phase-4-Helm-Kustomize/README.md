# Phase 4: Helm & Kustomize

## What you'll learn
In Phases 1-3 you wrote raw YAML files. Now you'll learn tools that make
managing K8s YAML **easier and reusable**.

## The Problem
```
Real app = 10-20 YAML files (deployment, service, configmap, secret, ingress...)
3 environments = 30-60 YAML files with small differences
Updates = editing the same values in multiple files

This gets messy fast!
```

## Two Solutions

| Tool | What it does | Analogy |
|------|-------------|---------|
| **Helm** | Package manager for K8s | Like npm/pip but for K8s apps |
| **Kustomize** | Patch/overlay YAML files | Like CSS overrides for YAML |

### When to use what?
```
Helm       → Installing 3rd party apps (nginx, prometheus, grafana)
             Sharing your app as a reusable package
             Complex templating with variables

Kustomize  → Managing your own app across environments (dev/staging/prod)
             Simple overrides without templates
             Built into kubectl (no install needed!)
```

## Learning Order

| # | Topic | What You'll Learn |
|---|-------|-------------------|
| 01 | Helm Basics | Install apps with one command |
| 02 | Helm Create Chart | Create your own Helm chart |
| 03 | Kustomize Basics | Base + overlay pattern |
| 04 | Kustomize Overlays | Dev vs Prod with one base |

## Prerequisites
- Phase 1-3 completed
- Minikube running
- Helm installed (instructions in 01-helm-basics)

Start with `01-helm-basics/`!
