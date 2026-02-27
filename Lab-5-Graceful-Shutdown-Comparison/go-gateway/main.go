package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"
)

var (
	shuttingDown   atomic.Bool
	activeRequests atomic.Int64
)

// Shared HTTP client with connection pooling and explicit timeouts.
// Without timeouts, a dying worker pod causes the gateway to hang forever.
var httpClient = &http.Client{
	Timeout: 12 * time.Second,
	Transport: &http.Transport{
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     30 * time.Second,
	},
}

func main() {
	workerURL := envOrDefault("WORKER_URL", "http://python-worker-svc:5000")
	retryMax, _ := strconv.Atoi(envOrDefault("RETRY_COUNT", "0"))
	retryDelay := 500 * time.Millisecond

	mux := http.NewServeMux()

	// --- Main endpoint ---
	// Calls Python worker, combines response with gateway metadata.
	// If RETRY_COUNT > 0 (v3), retries transient failures.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if shuttingDown.Load() {
			writeJSON(w, 503, map[string]interface{}{"error": "shutting down"})
			return
		}

		activeRequests.Add(1)
		defer activeRequests.Add(-1)

		start := time.Now()

		// Call worker with optional retry logic
		var resp *http.Response
		var lastErr error
		retries := 0

		for attempt := 0; attempt <= retryMax; attempt++ {
			resp, lastErr = httpClient.Get(workerURL + "/process")
			if lastErr == nil && resp.StatusCode == 200 {
				break
			}
			// Close failed response body to avoid connection leak
			if resp != nil {
				resp.Body.Close()
				resp = nil
			}
			retries = attempt + 1
			if attempt < retryMax {
				log.Printf("Retry %d/%d to worker (err=%v)", retries, retryMax, lastErr)
				time.Sleep(retryDelay)
			}
		}

		// Worker unreachable after all attempts
		if lastErr != nil {
			writeJSON(w, 502, map[string]interface{}{
				"error":   "worker_unreachable",
				"detail":  lastErr.Error(),
				"retries": retries,
			})
			return
		}
		defer resp.Body.Close()

		// Worker returned non-200
		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			writeJSON(w, resp.StatusCode, map[string]interface{}{
				"error":   "worker_error",
				"status":  resp.StatusCode,
				"body":    string(body),
				"retries": retries,
			})
			return
		}

		// Success — combine gateway + worker data
		var workerData map[string]interface{}
		body, _ := io.ReadAll(resp.Body)
		json.Unmarshal(body, &workerData)

		hostname, _ := os.Hostname()
		writeJSON(w, 200, map[string]interface{}{
			"status": "ok",
			"gateway": map[string]interface{}{
				"service":    "go-gateway",
				"hostname":   hostname,
				"elapsed_ms": time.Since(start).Milliseconds(),
			},
			"worker":  workerData,
			"retries": retries,
		})
	})

	// --- Health check ---
	// Returns 503 during shutdown so K8s stops routing traffic here.
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if shuttingDown.Load() {
			http.Error(w, "shutting down", 503)
			return
		}
		fmt.Fprint(w, "ok")
	})

	// --- preStop hook endpoint ---
	// Called by K8s BEFORE SIGTERM. The 5s sleep gives kube-proxy time
	// to remove this pod from iptables/IPVS rules — after that, no new
	// traffic arrives, and we can safely shut down.
	mux.HandleFunc("/prestop", func(w http.ResponseWriter, r *http.Request) {
		log.Println("preStop: waiting 5s for kube-proxy to update endpoints...")
		time.Sleep(5 * time.Second)
		shuttingDown.Store(true)
		log.Println("preStop: done, rejecting new requests")
		fmt.Fprint(w, "ok")
	})

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// --- SIGTERM handler ---
	// After preStop finishes, K8s sends SIGTERM.
	// Shutdown() stops accepting new connections and waits for in-flight
	// requests to complete — this is the key to zero dropped requests.
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
		sig := <-sigChan

		log.Printf("%v received — draining %d active requests...", sig, activeRequests.Load())
		shuttingDown.Store(true)

		// 20s timeout: must be less than terminationGracePeriodSeconds (40s) minus preStop (5s)
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Shutdown error: %v", err)
		} else {
			log.Println("All requests drained, server stopped")
		}
	}()

	log.Printf("Go gateway on :8080 (worker=%s, retries=%d)", workerURL, retryMax)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
