# 01 - Pods

## What is a Pod?
A Pod is the **smallest thing you can deploy** in Kubernetes.
Think of it as a wrapper around one or more containers.

```
+------ Pod ------+
|  [Container 1]  |
|  (nginx:latest) |
+-----------------+
```

## Key Facts
- A Pod = 1 or more containers that share the same network and storage
- Every Pod gets its own IP address inside the cluster
- Pods are **ephemeral** - they can die and be replaced
- You rarely create Pods directly (you use Deployments), but we start here to learn

## Files in this folder

| File | What it does |
|------|-------------|
| `pod-basic.yaml` | Simplest possible pod - runs nginx |
| `pod-with-labels.yaml` | Pod with labels (used for organizing/selecting) |
| `pod-multi-container.yaml` | Pod with 2 containers (sidecar pattern) |

## Commands to Run (in order)

### 1. Create your first pod
```bash
kubectl apply -f pod-basic.yaml
```

### 2. Check if it's running
```bash
kubectl get pods
kubectl get pods -o wide    # shows IP and node
```

### 3. See pod details
```bash
kubectl describe pod my-first-pod
```

### 4. See pod logs
```bash
kubectl logs my-first-pod
```

### 5. Exec into the pod (like SSH)
```bash
kubectl exec -it my-first-pod -- /bin/sh
```

This opens a **shell inside the running container** - like SSH-ing into a server.

| Part | Meaning |
|------|---------|
| `exec` | execute a command inside a pod |
| `-i` | interactive - keep input open |
| `-t` | allocate a terminal (so you see a prompt) |
| `my-first-pod` | which pod to enter |
| `--` | separator between kubectl args and the command |
| `/bin/sh` | the command to run - opens a shell |

Once inside, you're **inside the nginx container**. Try:
```sh
ls /usr/share/nginx/html     # see the nginx welcome page file
hostname                      # shows the pod name
exit                          # come back out
```
This is super useful for **debugging** - checking files, testing network, verifying config inside a running container.

### 6. Port-forward to access from browser
```bash
kubectl port-forward my-first-pod 8080:80
# Now open http://localhost:8080 in your browser!
# Press Ctrl+C to stop port-forwarding
```

### 7. Try the labeled pod
```bash
kubectl apply -f pod-with-labels.yaml
kubectl get pods --show-labels
kubectl get pods -l app=web       # filter by label
```

### 8. Try multi-container pod
```bash
kubectl apply -f pod-multi-container.yaml
kubectl logs multi-container-pod -c nginx     # logs from specific container
kubectl logs multi-container-pod -c sidecar   # logs from sidecar
```

### 9. Clean up
```bash
kubectl delete -f pod-basic.yaml
kubectl delete -f pod-with-labels.yaml
kubectl delete -f pod-multi-container.yaml
# OR delete all at once:
# kubectl delete pod --all
```
