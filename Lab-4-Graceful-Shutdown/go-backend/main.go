package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

// isShuttingDown is an atomic flag — once set, new requests get rejected.
// This prevents accepting work we can't finish.
var isShuttingDown atomic.Bool

// activeRequests tracks how many requests are currently in-flight.
// We use this to drain gracefully before exiting.
var activeRequests atomic.Int64

func main() {
	mux := http.NewServeMux()

	// --- /api endpoint ---
	// Simulates real work with a random 200-800ms delay.
	// This is what the Python frontend calls.
	mux.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		// If we're shutting down, reject new requests immediately.
		// The client will get a 503 and retry on another pod.
		if isShuttingDown.Load() {
			http.Error(w, "shutting down", http.StatusServiceUnavailable)
			return
		}

		activeRequests.Add(1)
		defer activeRequests.Add(-1)

		// Simulate work: 200-800ms random delay
		// In real apps, this would be DB queries, API calls, etc.
		delay := time.Duration(200+rand.Intn(600)) * time.Millisecond
		time.Sleep(delay)

		hostname, _ := os.Hostname()
		resp := map[string]interface{}{
			"service":  "go-backend",
			"hostname": hostname,
			"delay_ms": delay.Milliseconds(),
			"time":     time.Now().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// --- /health endpoint ---
	// Used by Kubernetes readiness & liveness probes.
	// Returns 503 during shutdown so K8s stops sending traffic.
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if isShuttingDown.Load() {
			http.Error(w, "shutting down", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	// --- /prestop endpoint ---
	// Called by Kubernetes preStop lifecycle hook BEFORE SIGTERM.
	//
	// WHY THE 5-SECOND SLEEP?
	// When K8s decides to kill a pod, two things happen IN PARALLEL:
	//   1. kube-proxy starts removing the pod from iptables/IPVS rules
	//   2. The preStop hook runs
	//
	// The iptables update takes 1-3 seconds. If we start shutting down
	// immediately, we'll reject requests that are already in-flight
	// (routed before iptables updated). The 5s sleep gives kube-proxy
	// time to finish, so no new traffic arrives at this pod.
	mux.HandleFunc("/prestop", func(w http.ResponseWriter, r *http.Request) {
		log.Println("preStop hook called — waiting 5s for kube-proxy to update...")
		time.Sleep(5 * time.Second)

		log.Println("preStop done — marking as shutting down")
		isShuttingDown.Store(true)

		fmt.Fprint(w, "ok")
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
		// ReadTimeout/WriteTimeout protect against slow clients
		// holding connections open forever.
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// --- SIGTERM Handler ---
	// After preStop hook finishes, K8s sends SIGTERM.
	// We use http.Server.Shutdown() which:
	//   1. Stops accepting NEW connections
	//   2. Waits for IN-FLIGHT requests to complete
	//   3. Then returns
	// This is the key to zero dropped requests.
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
		sig := <-sigChan

		log.Printf("Received %v — starting graceful shutdown...", sig)
		isShuttingDown.Store(true)

		log.Printf("Active requests: %d — waiting for them to finish...", activeRequests.Load())

		// Give in-flight requests up to 25 seconds to complete.
		// This must be LESS than terminationGracePeriodSeconds (60s)
		// minus the preStop sleep (5s) = 55s available.
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Shutdown error: %v", err)
		} else {
			log.Println("Server shut down gracefully — all requests completed")
		}
	}()

	log.Println("Go backend starting on :8080")
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
	log.Println("Server exited")
}
