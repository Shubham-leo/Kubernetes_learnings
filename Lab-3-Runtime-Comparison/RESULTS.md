# Benchmark Results: Rust vs Go vs Python on Kubernetes

**Date:** 2026-02-25
**Environment:** Minikube on Docker Desktop, Windows 11
**Load Testing Tool:** Grafana K6
**Tested by:** Sujit Waghmare

---

## What We Tested

Three HTTP servers computing `fibonacci(30)` recursively — **832,040 function calls per request**. Same algorithm, same K8s config. Only the language changes.

### App Configuration

| | Rust | Go | Python |
|--|------|-----|--------|
| **Language** | Rust 1.85 | Go 1.22 | Python 3.12 |
| **Framework** | Actix-Web 4 (async) | net/http (stdlib) | Flask (WSGI) |
| **Execution** | Compiled native binary | Compiled binary + GC | Interpreter + GIL |
| **Docker image** | debian:bookworm-slim | golang:1.22-alpine | python:3.12-slim |
| **CPU request/limit** | 100m / 500m | 100m / 500m | 100m / 500m |
| **Memory request/limit** | 32Mi / 128Mi | 256Mi / 512Mi | 256Mi / 768Mi |

---

## Test 1: Image Size and Memory

### Docker Image Size

| Rust | Go | Python |
|------|-----|--------|
| **88 MB** | 231 MB | 119 MB |

Go is largest because `golang:1.22-alpine` includes the full compiler. In production with multi-stage build, Go would be ~15-20 MB.

### Memory at Idle (no traffic)

| Rust | Go | Python |
|------|-----|--------|
| **2 Mi** | 104 Mi | 41 Mi |

### Memory After 5 Min Sustained Load

| Rust | Go | Python |
|------|-----|--------|
| **2 Mi** | 107 Mi | 41 Mi |

No memory leaks in any runtime. Rust stayed at 2 Mi throughout — Go uses 52x more, Python 20x more.

---

## Test 2: Single Request — Raw Speed

One `curl` request, no concurrency:

| Rust | Go | Python |
|------|-----|--------|
| **~2ms** | ~4ms | ~159ms |

| vs Rust | Go | Python |
|---------|-----|--------|
| | 2x slower | 80x slower |

---

## Test 3: Load Test — Ramp to 100 Users (60s)

Gradual ramp: 10 → 50 → 100 → 0 users over 60 seconds.

### Rust
| Metric | Value |
|--------|-------|
| Requests/sec | **160.6** |
| Avg response | **224ms** |
| Total requests | 9,639 |
| HTTP errors | **0%** |

### Go
| Metric | Value |
|--------|-------|
| Requests/sec | **70.0** |
| Avg response | **581ms** |
| Total requests | ~4,200 |
| HTTP errors | **0%** |

### Python
| Metric | Value |
|--------|-------|
| Requests/sec | **3.6** |
| Avg response | **~14,000ms** |
| Total requests | ~216 |
| HTTP errors | **0%** |

---

## Test 4: Spike Test — Sudden 100 Users (80s)

5 users → 100 users in 5 seconds. Hold for 1 minute. Simulates a flash sale or viral moment.

### Rust
| Metric | Value |
|--------|-------|
| Requests/sec | **153.4** |
| Avg response | **503ms** |
| p95 response | **617ms** |
| Total requests | 12,279 |
| HTTP errors | **0%** |

### Go
| Metric | Value |
|--------|-------|
| Requests/sec | **68.3** |
| Avg response | **1.19s** |
| p95 response | **2.33s** |
| Total requests | 5,467 |
| HTTP errors | **0%** |

### Python
| Metric | Value |
|--------|-------|
| Requests/sec | **3.16** |
| Avg response | **25.74s** |
| p95 response | **33.8s** |
| Total requests | 315 |
| HTTP errors | **0%** |

When 100 users hit at once, Rust handled 12,279 requests while Python managed 315. Python's average response exploded to **25 seconds** — users would see spinning wheels for half a minute.

---

## Test 5: Soak Test — 30 Users for 5 Minutes

Sustained load to check stability and consistency over time.

### Rust
| Metric | Value |
|--------|-------|
| Requests/sec | **149.6** |
| Avg response | **144ms** |
| p95 response | **181ms** |
| Total requests | 47,884 |
| HTTP errors | **0%** |

### Go
| Metric | Value |
|--------|-------|
| Requests/sec | **67.8** |
| Avg response | **379ms** |
| p95 response | **712ms** |
| Total requests | 21,686 |
| HTTP errors | **0%** |

### Python
| Metric | Value |
|--------|-------|
| Requests/sec | **2.7** |
| Avg response | **10.62s** |
| p95 response | **13.6s** |
| Total requests | 885 |
| HTTP errors | **0%** |

Over 5 minutes: Rust served **47,884** requests, Python served **885**. Rust's p95 was 181ms — Python's was 13.6 seconds. Neither leaked memory.

---

## Test 6: Pod Kill Recovery

Killed each pod during active traffic (30 users soak test running). K8s automatically detected the crash and spun up a replacement.

| | Rust | Go | Python |
|--|------|-----|--------|
| **Recovery time** | **~3s** | ~16s | ~3s* |

*Python pod showed Ready quickly but needs additional time for `pip install flask` before actually serving. Go is slow because `go run` recompiles the source on every pod start.

In production with compiled binaries, Rust would recover in microseconds, Go in milliseconds.

---

## Test 7: Resource Squeeze — 50 Millicores CPU

Redeployed all 3 apps with only **50m CPU** (10x less than normal). Then ran the spike test (100 users).

### Rust (50m CPU)
| Metric | Value |
|--------|-------|
| Requests/sec | **17.5** |
| Avg response | **4.89s** |
| p95 response | **10.1s** |
| Total requests | 1,413 |
| HTTP errors | **5.23%** |

### Go (50m CPU)
| Metric | Value |
|--------|-------|
| Requests/sec | **—** |
| Result | **Connection dropped. Port-forward crashed under load.** |
| HTTP errors | **~100%** |

### Python (50m CPU)
| Metric | Value |
|--------|-------|
| Requests/sec | **0.96** |
| Avg response | **57.94s** |
| p95 response | **60s (timeout)** |
| Total requests | 106 |
| HTTP errors | **91.5%** |

At 50 millicores — barely enough CPU to run a clock:
- **Rust** survived. Slower than normal but still handled traffic with 94.8% success rate.
- **Go** crashed. The connection dropped entirely — couldn't even keep the port-forward alive.
- **Python** was functionally dead. 91.5% of requests timed out at 60 seconds.

This shows the floor — the minimum resources each runtime needs to stay alive under pressure.

---

## Summary — All Tests Side by Side

### Throughput (req/s)

```
                Normal Load    Spike    Soak (5min)   Squeeze (50m CPU)

  Rust            160.6        153.4      149.6           17.5
  Go               70.0         68.3       67.8           crashed
  Python            3.6          3.2        2.7            0.96
```

### Error Rates

| Test | Rust | Go | Python |
|------|------|-----|--------|
| Load (100 users) | 0% | 0% | 0% |
| Spike (100 users) | 0% | 0% | 0% |
| Soak (30 users, 5min) | 0% | 0% | 0% |
| Squeeze (50m CPU) | **5.2%** | **~100%** | **91.5%** |

### Resource Usage

| | Rust | Go | Python |
|--|------|-----|--------|
| Docker image | 88 MB | 231 MB | 119 MB |
| Memory (idle) | 2 Mi | 104 Mi | 41 Mi |
| Memory (under load) | 2 Mi | 109 Mi | 41 Mi |
| Min viable CPU | ~50m | >50m | >>50m |

---

## The Hierarchy

1. **Rust** — Fastest in every test. Lowest memory. Only runtime that survived the resource squeeze. 44x faster than Python, 2.3x faster than Go under normal load.

2. **Go** — Strong second. Consistent 70 req/s across all normal tests. Crashed under extreme resource constraints, but in production with proper limits this wouldn't happen.

3. **Python** — 3 req/s under load. 25-second avg response during spike. 91.5% failure at 50m CPU. Fine for prototyping and ML/AI, but pays a heavy tax on K8s for CPU-bound work.
