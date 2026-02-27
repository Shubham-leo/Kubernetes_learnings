package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

// draining flips to true once we begin the shutdown sequence.
// After this, the /health probe returns 503 and new inbound
// requests are turned away with a "service draining" message.
var draining atomic.Bool

// inflight keeps a running count of requests that have entered
// the handler but haven't returned yet. During shutdown we poll
// this counter to decide when it's safe to close the listener.
var inflight atomic.Int64

// downstreamBase points at the Python downstream service inside
// the cluster. Override with DOWNSTREAM_URL env var if you need
// a different address (e.g. during local development).
var downstreamBase string

func init() {
	downstreamBase = os.Getenv("DOWNSTREAM_URL")
	if downstreamBase == "" {
		downstreamBase = "http://downstream-svc:5000"
	}
}

// httpClient is shared across all requests. We set explicit
// timeouts so a dying downstream pod can't hold our goroutines
// open forever. Keeping MaxIdleConnsPerHost > 1 lets us reuse
// TCP connections to the same downstream pod across requests.
var httpClient = &http.Client{
	Timeout: 12 * time.Second,
	Transport: &http.Transport{
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     30 * time.Second,
	},
}

func main() {
	router := http.NewServeMux()

	// ---- /call ----
	// The main endpoint. External traffic (K6, curl, other services)
	// hits this route. We forward the request to the Python downstream
	// service and stitch both responses together before replying.
	router.HandleFunc("/call", handleCall)

	// ---- /health ----
	// Kubernetes readiness and liveness probes point here.
	// Returns 200 while the pod is ready, 503 once draining begins.
	router.HandleFunc("/health", handleHealth)

	// ---- /prestop ----
	// Kubernetes preStop lifecycle hook target.
	//
	// When K8s terminates a pod, two things happen concurrently:
	//   a) kube-proxy updates iptables/IPVS to stop routing new traffic here
	//   b) this preStop hook runs
	//
	// We sleep for 5 seconds to give kube-proxy enough time to finish
	// its iptables update. Without this pause, new requests can still
	// arrive for a couple of seconds after we mark ourselves as draining,
	// and those would receive an unnecessary 503.
	router.HandleFunc("/prestop", handlePreStop)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// ---- Signal listener ----
	// After the preStop hook returns, kubelet sends SIGTERM.
	// We catch it and trigger http.Server.Shutdown, which:
	//   1. Closes the listener so no new TCP connections come in
	//   2. Waits for every in-flight HTTP handler to return
	//   3. Returns nil (or an error if the context deadline fires first)
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
		caught := <-quit

		log.Printf("signal=%v received, beginning graceful shutdown", caught)
		draining.Store(true)

		// Log how many requests are still being processed.
		log.Printf("in-flight requests: %d", inflight.Load())

		// Allow up to 30 seconds for handlers to finish.
		// This must be shorter than terminationGracePeriodSeconds (60s)
		// minus the preStop sleep (5s) = 55s budget.
		ctx, stop := context.WithTimeout(context.Background(), 30*time.Second)
		defer stop()

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("forced shutdown: %v", err)
		} else {
			log.Println("all handlers returned, server closed cleanly")
		}
	}()

	log.Printf("go-upstream listening on :8080, downstream=%s", downstreamBase)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("listen error: %v", err)
	}
	log.Println("process exiting")
}

// --- handlers ---

func handleCall(w http.ResponseWriter, r *http.Request) {
	if draining.Load() {
		http.Error(w, `{"error":"service draining"}`, http.StatusServiceUnavailable)
		return
	}

	inflight.Add(1)
	defer inflight.Add(-1)

	start := time.Now()

	// Call the Python downstream service.
	resp, err := httpClient.Get(downstreamBase + "/process")
	if err != nil {
		log.Printf("downstream error: %v", err)
		writeJSON(w, http.StatusBadGateway, map[string]interface{}{
			"error":  "downstream_unreachable",
			"detail": err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	// If the downstream returned a non-200, relay its status.
	if resp.StatusCode != http.StatusOK {
		writeJSON(w, resp.StatusCode, map[string]interface{}{
			"error":             "downstream_error",
			"downstream_status": resp.StatusCode,
		})
		return
	}

	var downstream map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&downstream); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"error":  "bad_downstream_json",
			"detail": err.Error(),
		})
		return
	}

	hostname, _ := os.Hostname()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"upstream": map[string]interface{}{
			"service":    "go-upstream",
			"hostname":   hostname,
			"elapsed_ms": time.Since(start).Milliseconds(),
		},
		"downstream": downstream,
	})
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	if draining.Load() {
		http.Error(w, "draining", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "ok")
}

func handlePreStop(w http.ResponseWriter, _ *http.Request) {
	log.Println("prestop hook fired -- sleeping 5s for kube-proxy convergence")
	time.Sleep(5 * time.Second)
	draining.Store(true)
	log.Println("prestop complete -- draining flag set")
	fmt.Fprint(w, "done")
}

// writeJSON is a small helper that sets Content-Type, writes the
// status code, and marshals the payload as JSON.
func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}
