# 03 - Kustomize Basics

## What is Kustomize?
Kustomize lets you **customize YAML files without editing them**.
You write a base, then apply patches/overlays for each environment.

```
Helm approach:     templates + variables     (like a form with blanks to fill)
Kustomize approach: base YAML + patches      (like a document with sticky notes)
```

## Key advantage: Built into kubectl!
```bash
# No install needed! Already part of kubectl:
kubectl apply -k ./my-folder       # the -k flag = kustomize
```

## How it works
```
base/                           ← your original YAML (unchanged)
  ├── deployment.yaml
  ├── service.yaml
  └── kustomization.yaml        ← tells kustomize which files to use

overlays/dev/                   ← changes for dev environment
  ├── kustomization.yaml        ← "use base + these patches"
  └── replica-patch.yaml        ← "change replicas to 1"

overlays/prod/                  ← changes for production
  ├── kustomization.yaml
  └── replica-patch.yaml        ← "change replicas to 5"
```

### Result:
```
kubectl apply -k overlays/dev/    → deploys with 1 replica
kubectl apply -k overlays/prod/   → deploys with 5 replicas
                                    Same base, different configs!
```

## Files in this folder

```
base/
  ├── deployment.yaml
  ├── service.yaml
  └── kustomization.yaml
```

## Commands to Run

### 1. Look at the base files
```bash
cat base/kustomization.yaml
cat base/deployment.yaml
```

### 2. Preview what Kustomize generates
```bash
kubectl kustomize base/
# Shows the final YAML output
```

### 3. Apply with Kustomize
```bash
kubectl apply -k base/
kubectl get all
```

### 4. Add common labels/annotations
```bash
# Kustomize can add labels to ALL resources at once
# Check kustomization.yaml - it adds "managed-by: kustomize" to everything
kubectl get all --show-labels
```

### 5. Clean up
```bash
kubectl delete -k base/
```
