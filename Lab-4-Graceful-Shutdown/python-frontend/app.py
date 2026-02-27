"""
Python Frontend — Flask app that calls the Go backend.

This is the "upstream" side of the chain:
  K6 → Python Frontend → Go Backend

Key points:
  - Uses explicit timeouts on requests to the backend (connect=2s, read=10s)
  - Without timeouts, a dying backend causes Python to hang indefinitely
  - SIGTERM handler ensures gunicorn workers finish in-flight requests
"""

import json
import logging
import os
import signal
import sys
import time

import requests
from flask import Flask, jsonify

app = Flask(__name__)
logging.basicConfig(level=logging.INFO, format="%(asctime)s %(message)s")

# Where the Go backend lives inside the cluster
BACKEND_URL = os.environ.get("BACKEND_URL", "http://backend-svc:8080")

# Explicit timeouts prevent hanging when the backend pod is terminating.
# connect_timeout: how long to wait for TCP handshake (2s is generous inside a cluster)
# read_timeout: how long to wait for the response body (10s covers the 200-800ms work + buffer)
CONNECT_TIMEOUT = float(os.environ.get("CONNECT_TIMEOUT", "2"))
READ_TIMEOUT = float(os.environ.get("READ_TIMEOUT", "10"))


@app.route("/")
def index():
    """Call Go backend and return combined response."""
    hostname = os.environ.get("HOSTNAME", "unknown")
    start = time.time()

    try:
        resp = requests.get(
            f"{BACKEND_URL}/api",
            timeout=(CONNECT_TIMEOUT, READ_TIMEOUT),
        )
        resp.raise_for_status()
        backend_data = resp.json()
        elapsed = round((time.time() - start) * 1000, 1)

        return jsonify({
            "frontend": {
                "service": "python-frontend",
                "hostname": hostname,
                "elapsed_ms": elapsed,
            },
            "backend": backend_data,
        })

    except requests.exceptions.ConnectionError as e:
        # This is the 502 scenario: backend pod is gone, TCP connection refused
        logging.error(f"Connection error to backend: {e}")
        return jsonify({"error": "backend_connection_failed", "detail": str(e)}), 502

    except requests.exceptions.Timeout as e:
        # This is the 504 scenario: backend accepted connection but didn't respond
        # (it's shutting down mid-request)
        logging.error(f"Timeout calling backend: {e}")
        return jsonify({"error": "backend_timeout", "detail": str(e)}), 504

    except requests.exceptions.RequestException as e:
        logging.error(f"Request error: {e}")
        return jsonify({"error": "backend_error", "detail": str(e)}), 500


@app.route("/health")
def health():
    """Readiness/liveness probe endpoint."""
    return "ok", 200


@app.route("/prestop")
def prestop():
    """
    preStop lifecycle hook.

    Same logic as the Go backend: sleep 5 seconds to let kube-proxy
    remove this pod from the Service endpoints before we start shutting down.
    """
    logging.info("preStop hook called — waiting 5s for kube-proxy to update...")
    time.sleep(5)
    logging.info("preStop done")
    return "ok", 200


def sigterm_handler(signum, frame):
    """
    Handle SIGTERM from Kubernetes.

    When running under gunicorn, SIGTERM goes to the master process,
    which then gracefully shuts down workers. But if running directly
    (e.g., flask run), we need to handle it ourselves.
    """
    logging.info("Received SIGTERM — shutting down gracefully")
    sys.exit(0)


signal.signal(signal.SIGTERM, sigterm_handler)

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=5000)
