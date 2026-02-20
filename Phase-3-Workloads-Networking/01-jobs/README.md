# 01 - Jobs

## What is a Job?
A Job runs a task **once and exits**. Unlike Deployments that keep pods running forever,
Jobs are for one-time tasks.

```
Deployment:  Start → Run forever → never stops
Job:         Start → Do work → Complete ✅ → done!
```

## Real world examples
- Database migration
- Sending a batch of emails
- Processing a video
- Running a backup
- Data import/export

## Job vs Deployment

| | Deployment | Job |
|---|-----------|-----|
| Goal | Keep app running forever | Run task and finish |
| Pod restarts? | Yes, always | Only if it fails |
| Completion | Never "done" | Completes when task finishes |
| Use case | Web servers, APIs | Scripts, batch tasks |

## Types of Jobs

```
Simple Job:        1 pod, runs once
  [Pod] → done ✅

Parallel Job:      multiple pods run at the same time
  [Pod-1] → done ✅
  [Pod-2] → done ✅
  [Pod-3] → done ✅

Completions Job:   runs N times total (can be parallel)
  Run 1 → done ✅
  Run 2 → done ✅
  Run 3 → done ✅
  Run 4 → done ✅
  Run 5 → done ✅   (completions: 5)
```

## Files in this folder

| File | What it does |
|------|-------------|
| `job-basic.yaml` | Simple one-time job |
| `job-parallel.yaml` | Job that runs multiple pods in parallel |
| `job-failure.yaml` | Job that fails (to see retry behavior) |

## Commands to Run

### 1. Create a simple Job
```bash
kubectl apply -f job-basic.yaml
kubectl get jobs
kubectl get pods
```

### 2. Check the job output
```bash
# Wait until STATUS shows "Completed"
kubectl logs job/math-job
# You'll see: "The answer to 6 x 7 is 42"
```

### 3. See job details
```bash
kubectl describe job math-job
# Look for "Succeeded: 1" in the output
```

### 4. Parallel Job
```bash
kubectl apply -f job-parallel.yaml
kubectl get pods -w
# Watch multiple pods run at the same time!
# Press Ctrl+C after they complete

kubectl get jobs
# COMPLETIONS should show 5/5
```

### 5. Job that fails and retries
```bash
kubectl apply -f job-failure.yaml
kubectl get pods -w
# Watch it fail and retry (backoffLimit: 3)
# After 3 failures, the job gives up

kubectl describe job failing-job
# Look at Events - you'll see the retries
```

### 6. Clean up
```bash
kubectl delete -f job-basic.yaml
kubectl delete -f job-parallel.yaml
kubectl delete -f job-failure.yaml
```
