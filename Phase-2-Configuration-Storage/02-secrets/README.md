# 02 - Secrets

## What is a Secret?
A Secret is like a ConfigMap but for **sensitive data** - passwords, API keys, tokens.

```
ConfigMap:                    Secret:
  DB_HOST=10.0.0.5             DB_PASSWORD=supersecret123
  APP_MODE=dev                 API_KEY=abc-xyz-789
  (non-sensitive)              (sensitive - base64 encoded)
```

## ConfigMap vs Secret

| | ConfigMap | Secret |
|---|----------|--------|
| Data type | Non-sensitive config | Passwords, keys, tokens |
| Storage | Plain text | Base64 encoded |
| Secure? | No | Slightly (not encrypted by default!) |

**Important:** Secrets are base64 **encoded**, NOT encrypted. Anyone with cluster access can decode them. For real security, use tools like Vault or enable encryption at rest.

## Files in this folder

| File | What it does |
|------|-------------|
| `secret-basic.yaml` | Secret with username/password |
| `pod-env-secret.yaml` | Pod that reads Secret as env vars |
| `pod-volume-secret.yaml` | Pod that mounts Secret as a file |

## Commands to Run

### 1. Create a Secret from YAML
```bash
kubectl apply -f secret-basic.yaml
kubectl get secrets
kubectl describe secret db-credentials
```

### 2. Decode the secret (base64)
```bash
# Secrets are base64 encoded, not encrypted!
kubectl get secret db-credentials -o jsonpath='{.data.DB_PASSWORD}' | base64 --decode
```

### 3. Create a Secret from command line (easier)
```bash
kubectl create secret generic my-secret \
  --from-literal=username=admin \
  --from-literal=password=mypassword123
```

### 4. Pod that reads Secret as env vars
```bash
kubectl apply -f pod-env-secret.yaml
kubectl get pods

# Verify the secret is inside the pod:
kubectl exec secret-env-pod -- env | grep -E "DB_USER|DB_PASSWORD"
```

### 5. Pod that mounts Secret as files
```bash
kubectl apply -f pod-volume-secret.yaml
kubectl get pods

# Each key becomes a file:
kubectl exec secret-volume-pod -- ls /etc/secrets
kubectl exec secret-volume-pod -- cat /etc/secrets/DB_PASSWORD
```

### 6. Clean up
```bash
kubectl delete -f pod-env-secret.yaml
kubectl delete -f pod-volume-secret.yaml
kubectl delete -f secret-basic.yaml
kubectl delete secret my-secret
```
