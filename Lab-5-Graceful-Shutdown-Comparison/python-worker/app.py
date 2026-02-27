"""
Python Worker — Flask + gunicorn, does the actual processing.

The Go gateway calls this service at /process.
Simulates 200-800ms of work and returns JSON.
"""

import logging
import os
import random
import signal
import sys
import time

from flask import Flask, jsonify

app = Flask(__name__)
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(process)d] %(message)s",
)

shutting_down = False


@app.route("/process")
def process():
    """Simulate backend work with 200-800ms random delay."""
    if shutting_down:
        return jsonify({"error": "shutting down"}), 503

    delay_ms = random.randint(200, 800)
    time.sleep(delay_ms / 1000.0)

    hostname = os.environ.get("HOSTNAME", "unknown")
    return jsonify({
        "service": "python-worker",
        "hostname": hostname,
        "delay_ms": delay_ms,
        "time": time.strftime("%Y-%m-%dT%H:%M:%S%z"),
    })


@app.route("/health")
def health():
    """Readiness/liveness probe. Returns 503 during shutdown."""
    if shutting_down:
        return "shutting down", 503
    return "ok", 200


@app.route("/prestop")
def prestop():
    """
    preStop lifecycle hook.
    Sleep 5 seconds so kube-proxy can remove this pod from Service
    endpoints before we stop accepting requests.
    """
    global shutting_down
    logging.info("preStop: waiting 5s for kube-proxy to update endpoints...")
    time.sleep(5)
    shutting_down = True
    logging.info("preStop: done, rejecting new requests")
    return "ok", 200


def sigterm_handler(signum, frame):
    """
    Handle SIGTERM from Kubernetes.
    Under gunicorn, SIGTERM goes to the master process which then
    gracefully shuts down workers (--graceful-timeout controls the deadline).
    This handler covers the direct-run case (flask run / python app.py).
    """
    global shutting_down
    shutting_down = True
    logging.info("SIGTERM received, shutting down")
    sys.exit(0)


signal.signal(signal.SIGTERM, sigterm_handler)

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=5000)
