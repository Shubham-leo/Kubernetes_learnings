# Lab 4: Exercises — Test Your Understanding

Work through these exercises in order. Each one builds on the previous.

---

## Exercise 1: Understand the Pod Lifecycle

### Setup
```bash
eval $(minikube docker-env)
docker build -t go-backend:v1 ./go-backend/
docker build -t python-frontend:v1 ./python-frontend/
```

### Task 1.1: Watch a pod die
```bash
# Deploy just the backend (problem version)
kubectl apply -f k8s-problem/backend-deployment.yaml

# Watch pods in real-time
kubectl get pods -w
```

In another terminal:
```bash
# Delete one pod manually and watch what happens
kubectl delete pod $(kubectl get pods -l app=go-backend -o jsonpath='{.items[0].metadata.name}')
```

**Q: How fast did the pod go from Running → Terminating → Gone?**
**Q: Did the replacement pod start before or after the old one died?**

### Task 1.2: Check what signals the pod receives
```bash
# Watch logs of a backend pod
kubectl logs -f $(kubectl get pods -l app=go-backend -o jsonpath='{.items[0].metadata.name}')
```

In another terminal, delete that pod. Watch the logs.

**Q: Did you see the "Received SIGTERM" log message?**
**Q: Did you see "Server shut down gracefully"? Why or why not?**

### Cleanup
```bash
kubectl delete -f k8s-problem/backend-deployment.yaml
```

---

## Exercise 2: See the 502/504 Problem

### Task 2.1: Deploy the problem version
```bash
kubectl apply -f k8s-problem/
kubectl get pods -w
# Wait until all pods show Running
```

### Task 2.2: Test manually first
```bash
# Get the frontend URL
FRONTEND_URL=http://$(minikube ip):30500

# Send a single request — should return JSON with frontend + backend data
curl $FRONTEND_URL
```

**Q: What fields do you see in the response?**
**Q: Which hostname is the backend? Which is the frontend?**

### Task 2.3: Trigger errors during rolling update
Terminal 1 — send continuous requests:
```bash
FRONTEND_URL=http://$(minikube ip):30500
while true; do
  STATUS=$(curl -s -o /dev/null -w "%{http_code}" $FRONTEND_URL)
  if [ "$STATUS" != "200" ]; then
    echo "ERROR: Got status $STATUS at $(date +%T)"
  else
    echo "OK: 200 at $(date +%T)"
  fi
  sleep 0.3
done
```

Terminal 2 — trigger a rolling update:
```bash
kubectl rollout restart deployment/go-backend
```

**Q: Did you see any non-200 status codes? What were they?**
**Q: How many errors appeared? Over what time window?**

### Task 2.4: Now restart the frontend too
Keep Terminal 1 running. In Terminal 2:
```bash
kubectl rollout restart deployment/python-frontend
```

**Q: Did you get different error codes compared to just restarting the backend?**
**Q: Why might restarting the frontend cause different errors than restarting the backend?**

### Cleanup
```bash
# Stop the curl loop in Terminal 1 (Ctrl+C)
kubectl delete -f k8s-problem/
```

---

## Exercise 3: See the Fix in Action

### Task 3.1: Deploy the solution version
```bash
kubectl apply -f k8s-solution/
kubectl get pods -w
# Wait for all pods to show Running AND Ready (1/1)
```

### Task 3.2: Compare the deployment specs
```bash
# Look at the rolling update strategy
kubectl get deployment go-backend -o jsonpath='{.spec.strategy}' | python -m json.tool
```

**Q: What is maxUnavailable set to? What does that mean?**
**Q: What is maxSurge set to? What does that mean?**

### Task 3.3: Check the probes
```bash
kubectl describe deployment go-backend | grep -A5 "Readiness\|Liveness\|Lifecycle"
```

**Q: What endpoint does the readiness probe check?**
**Q: What does the preStop hook do?**

### Task 3.4: Trigger rolling update under load
Terminal 1 — continuous requests:
```bash
FRONTEND_URL=http://$(minikube ip):30500
COUNT=0
ERRORS=0
while true; do
  STATUS=$(curl -s -o /dev/null -w "%{http_code}" $FRONTEND_URL)
  COUNT=$((COUNT+1))
  if [ "$STATUS" != "200" ]; then
    ERRORS=$((ERRORS+1))
    echo "ERROR #$ERRORS: Got status $STATUS (total: $COUNT requests)"
  fi
  sleep 0.3
done
```

Terminal 2 — trigger rolling update:
```bash
kubectl rollout restart deployment/go-backend && \
kubectl rollout restart deployment/python-frontend
```

Wait for the rollout to complete:
```bash
kubectl rollout status deployment/go-backend
kubectl rollout status deployment/python-frontend
```

Then stop the curl loop (Ctrl+C).

**Q: How many total errors did you get?**
**Q: Compare this to Exercise 2. What's the difference?**

### Cleanup
```bash
kubectl delete -f k8s-solution/
```

---

## Exercise 4: Understand the Timing

### Task 4.1: Watch the shutdown sequence
```bash
kubectl apply -f k8s-solution/
kubectl get pods -w
# Wait for Ready
```

Follow logs of one backend pod:
```bash
POD=$(kubectl get pods -l app=go-backend -o jsonpath='{.items[0].metadata.name}')
kubectl logs -f $POD
```

In another terminal, delete that specific pod:
```bash
kubectl delete pod $POD
```

**Q: What's the order of log messages you see?**
**Q: How many seconds between "preStop hook called" and "Received SIGTERM"?**
**Q: Why is that delay important?**

### Task 4.2: What if terminationGracePeriodSeconds is too short?

Think about this scenario (don't run it, just reason):
```yaml
terminationGracePeriodSeconds: 3   # Only 3 seconds!
preStop:
  httpGet:
    path: /prestop              # Sleeps 5 seconds
```

**Q: What happens if preStop takes 5s but the grace period is only 3s?**
**Q: What signal does K8s send when the grace period expires?**
**Q: Can you catch or handle that signal?**

### Cleanup
```bash
kubectl delete -f k8s-solution/
```

---

## Exercise 5: K6 Load Testing

### Task 5.1: Run K6 against the problem version
```bash
kubectl apply -f k8s-problem/
kubectl get pods -w
# Wait for Running
```

Terminal 1:
```bash
k6 run -e FRONTEND_URL=http://$(minikube ip):30500 k6-test.js
```

Terminal 2 (about 15 seconds after K6 starts):
```bash
kubectl rollout restart deployment/go-backend && \
kubectl rollout restart deployment/python-frontend
```

**Q: What was the final error_rate?**
**Q: How many 502s vs 504s did you get?**
**Q: Did the test pass or fail the threshold (< 1% errors)?**

### Task 5.2: Run K6 against the solution version
```bash
kubectl delete -f k8s-problem/
kubectl apply -f k8s-solution/
kubectl get pods -w
# Wait for Ready (1/1)
```

Repeat the same K6 test + rolling update.

**Q: What's the error_rate now?**
**Q: Did the test pass?**

### Cleanup
```bash
kubectl delete -f k8s-solution/
```

---

## Exercise 6: Thought Experiments

No cluster needed — just think through these.

### 6.1: The timeout chain
```
K6 (15s timeout) → Python (2s connect, 10s read) → Go (10s read/write)
```

**Q: If the Go backend takes 12 seconds to respond, what happens?**
**Q: If the Go backend is unreachable (pod deleted), how long until Python gets an error?**
**Q: Why should the outer timeout always be larger than the inner timeout?**

### 6.2: preStop vs readiness probe
Both can prevent traffic from reaching a dying pod. How?

**Q: How does the preStop hook prevent traffic?** (hint: kube-proxy)
**Q: How does a failing readiness probe prevent traffic?** (hint: endpoints)
**Q: Why do we use BOTH?** (hint: what if one is slow?)

### 6.3: Why not just sleep in the SIGTERM handler?
Instead of a preStop hook, you could do:
```go
signal.Notify(sigChan, syscall.SIGTERM)
<-sigChan
time.Sleep(5 * time.Second)  // sleep here instead of preStop
server.Shutdown(ctx)
```

**Q: Would this work the same? Why or why not?**
**Q: What's the difference between preStop and SIGTERM in the K8s lifecycle?**

### 6.4: Docker CMD format matters
```dockerfile
# Format A
CMD ./server

# Format B
CMD ["./server"]
```

**Q: Which format receives SIGTERM directly?**
**Q: What does Format A actually run? (hint: /bin/sh -c)**
**Q: Why does this break graceful shutdown?**

---

## Answer Key

<details>
<summary>Exercise 4.2: terminationGracePeriodSeconds too short</summary>

- K8s sends SIGKILL after 3 seconds, killing the pod mid-preStop
- SIGKILL cannot be caught, handled, or ignored
- The pod dies immediately with no graceful shutdown
- Rule: `terminationGracePeriodSeconds` must be > preStop time + drain time

</details>

<details>
<summary>Exercise 6.1: The timeout chain</summary>

- Go takes 12s but has a 10s WriteTimeout → Go returns an error or closes connection
- Python's read timeout is 10s → Python gets a timeout error after 10s
- If Go is unreachable, Python's connect timeout fires after 2s
- Outer timeout must be larger so it doesn't mask the specific inner error. If K6's 15s timeout fires, you just see "timeout". If Python's 10s fires, you see "backend_timeout" — more useful for debugging.

</details>

<details>
<summary>Exercise 6.2: preStop vs readiness probe</summary>

- preStop hook: gives kube-proxy time to remove pod from iptables BEFORE the app starts shutting down
- Failing readiness probe: K8s removes the pod from the Service endpoints list, so new requests aren't routed to it
- We use both because: kube-proxy updates can lag (preStop covers the gap), and readiness probes catch cases where the app is unhealthy but hasn't received SIGTERM yet. Belt and suspenders.

</details>

<details>
<summary>Exercise 6.3: Sleep in SIGTERM handler vs preStop</summary>

- Not the same. K8s lifecycle: preStop runs FIRST, then SIGTERM is sent AFTER preStop completes.
- If you sleep in the SIGTERM handler, kube-proxy may not have finished updating by the time SIGTERM arrives (no guaranteed delay).
- preStop runs in parallel with kube-proxy updates. SIGTERM runs after preStop. The preStop sleep is specifically designed to cover the kube-proxy race.
- Also: preStop is managed by K8s, so it works regardless of the application language or signal handling.

</details>

<details>
<summary>Exercise 6.4: Docker CMD format</summary>

- Format B (`CMD ["./server"]`) — exec form, process is PID 1, receives SIGTERM directly.
- Format A (`CMD ./server`) — shell form, runs as `/bin/sh -c ./server`. The shell is PID 1. SIGTERM goes to the shell, which doesn't forward it to the Go process.
- The Go process never sees SIGTERM → no graceful shutdown → requests are dropped.

</details>
