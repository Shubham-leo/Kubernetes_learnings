# 02 - CronJobs

## What is a CronJob?
A CronJob runs a Job **on a schedule**. Like cron on Linux or Task Scheduler on Windows.

```
CronJob: "Run this every 5 minutes"
  ├── 10:00 → Job-1 → done ✅
  ├── 10:05 → Job-2 → done ✅
  ├── 10:10 → Job-3 → done ✅
  └── 10:15 → Job-4 → running...
```

## Real world examples
- Backup database every night at 2 AM
- Send reports every Monday morning
- Clean up temp files every hour
- Check for expired sessions every 5 minutes

## Cron Schedule Syntax

```
┌───────── minute (0 - 59)
│ ┌───────── hour (0 - 23)
│ │ ┌───────── day of month (1 - 31)
│ │ │ ┌───────── month (1 - 12)
│ │ │ │ ┌───────── day of week (0 - 6, 0 = Sunday)
│ │ │ │ │
* * * * *
```

### Common examples
```
*/1 * * * *     → every 1 minute
*/5 * * * *     → every 5 minutes
0 * * * *       → every hour (at minute 0)
0 2 * * *       → every day at 2:00 AM
0 0 * * 0       → every Sunday at midnight
0 9 * * 1       → every Monday at 9:00 AM
```

## Files in this folder

| File | What it does |
|------|-------------|
| `cronjob-basic.yaml` | Runs every minute (for demo) |
| `cronjob-cleanup.yaml` | Simulates a cleanup task every 2 minutes |

## Commands to Run

### 1. Create a CronJob
```bash
kubectl apply -f cronjob-basic.yaml
kubectl get cronjobs       # or: kubectl get cj
```

### 2. Watch it trigger every minute
```bash
kubectl get jobs -w
# Wait 1-2 minutes, you'll see new jobs appear!
# Press Ctrl+C after seeing 2-3 jobs

kubectl get pods
# Each job created a pod
```

### 3. Check the output
```bash
kubectl logs job/<job-name>    # use a job name from 'kubectl get jobs'
```

### 4. Manually trigger a CronJob
```bash
kubectl create job manual-run --from=cronjob/hello-cron
kubectl get pods
kubectl logs job/manual-run
```

### 5. Suspend a CronJob (pause it)
```bash
kubectl patch cronjob hello-cron -p '{"spec":{"suspend":true}}'
kubectl get cj    # SUSPEND shows True

# Resume it
kubectl patch cronjob hello-cron -p '{"spec":{"suspend":false}}'
```

### 6. Clean up
```bash
kubectl delete -f cronjob-basic.yaml
kubectl delete -f cronjob-cleanup.yaml
kubectl delete job manual-run
```
