# 01 - Helm Basics

## What is Helm?
Helm is the **package manager for Kubernetes**. Like npm for Node.js or pip for Python.

```
Without Helm:                          With Helm:
  kubectl apply -f deployment.yaml       helm install myapp nginx-chart
  kubectl apply -f service.yaml          (one command installs everything!)
  kubectl apply -f configmap.yaml
  kubectl apply -f secret.yaml
  kubectl apply -f ingress.yaml
  kubectl apply -f pvc.yaml
  ↑ 6 commands, 6 files                  ↑ 1 command
```

## Key Concepts

```
Chart      = a package (folder with templates + values)
Release    = an installed instance of a chart
Repository = where charts are stored (like npm registry)
Values     = configuration variables you can customize
```

### Analogy
```
Chart      = recipe (instructions to make the app)
Values     = ingredients list (you can change quantities)
Release    = the actual dish you made (running in your cluster)
Repository = cookbook (collection of recipes)
```

## How Helm works
```
Chart (template):                  Values (your config):
  deployment.yaml                    replicas: 3
    replicas: {{ .Values.replicas }} image: nginx:1.25
    image: {{ .Values.image }}       port: 80

         ↓ helm combines them ↓

Final YAML (applied to cluster):
  deployment.yaml
    replicas: 3
    image: nginx:1.25
```

## Install Helm

### Windows
```powershell
# Option 1: Download installer
# Go to https://github.com/helm/helm/releases
# Download helm-vX.X.X-windows-amd64.zip
# Extract helm.exe to a folder in your PATH

# Option 2: Using chocolatey (if installed)
choco install kubernetes-helm

# Option 3: Using scoop (if installed)
scoop install helm
```

### Verify
```bash
helm version
```

## Commands to Run

### 1. Add a chart repository
```bash
# Add the official Bitnami repo (has lots of popular apps)
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
```

### 2. Search for charts
```bash
helm search repo nginx
helm search repo postgresql
helm search repo redis
# Hundreds of apps available!
```

### 3. Install nginx using Helm (one command!)
```bash
helm install my-nginx bitnami/nginx
# This created: Deployment + Service + ConfigMap + more!

kubectl get all
# See everything Helm created
```

### 4. See what's installed
```bash
helm list                    # or: helm ls
helm status my-nginx         # detailed status
```

### 5. Customize with values
```bash
# See what values you can configure:
helm show values bitnami/nginx | head -50

# Install with custom values:
helm install my-nginx-custom bitnami/nginx \
  --set replicaCount=3 \
  --set service.type=NodePort
```

### 6. Install from a values file
```bash
helm install my-nginx-values bitnami/nginx -f custom-values.yaml
kubectl get pods    # should see 2 replicas
kubectl get svc     # should see NodePort type
```

### 7. Upgrade a release
```bash
helm upgrade my-nginx bitnami/nginx --set replicaCount=5
kubectl get pods    # now 5 replicas
```

### 8. Rollback
```bash
helm history my-nginx
helm rollback my-nginx 1     # rollback to revision 1
```

### 9. Uninstall
```bash
helm uninstall my-nginx
helm uninstall my-nginx-custom
helm uninstall my-nginx-values
```
