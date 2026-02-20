# 02 - Create Your Own Helm Chart

## Why create your own chart?
Instead of writing 5+ YAML files for your app, create ONE chart that's
reusable across environments.

## Chart structure
```
my-app-chart/
├── Chart.yaml           ← chart metadata (name, version)
├── values.yaml          ← default configuration
├── templates/           ← YAML templates with variables
│   ├── deployment.yaml
│   ├── service.yaml
│   └── configmap.yaml
└── .helmignore          ← files to ignore (like .gitignore)
```

## How templates work
```
values.yaml:                  templates/deployment.yaml:
  appName: my-web-app           name: {{ .Values.appName }}
  replicas: 3                   replicas: {{ .Values.replicas }}
  image: nginx:1.25             image: {{ .Values.image }}

    ↓ helm install combines them ↓

Final output:
  name: my-web-app
  replicas: 3
  image: nginx:1.25
```

## Commands to Run

### 1. Create a chart from scratch
```bash
helm create my-app-chart
# This generates a full chart with templates!
```

### 2. But we'll use our own (simpler) chart
```bash
# Look at the chart I prepared:
ls my-app-chart/
ls my-app-chart/templates/
```

### 3. See what Helm will generate (dry run)
```bash
helm template my-release ./my-app-chart
# Shows the final YAML without installing
```

### 4. Install your chart with default values
```bash
helm install my-app ./my-app-chart
kubectl get all
```

### 5. Install with different values (like for production)
```bash
helm install my-app-prod ./my-app-chart -f prod-values.yaml
kubectl get all
# Different replicas, different config!
```

### 6. Package your chart (to share with others)
```bash
helm package ./my-app-chart
# Creates my-app-chart-0.1.0.tgz
```

### 7. Clean up
```bash
helm uninstall my-app
helm uninstall my-app-prod
```
