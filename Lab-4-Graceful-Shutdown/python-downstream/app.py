"""
Python Downstream -- the leaf service in the request chain.

Traffic flow:
  K6 -> Go Upstream (:8080/call) -> Python Downstream (:5000/process)

This service simulates real backend work (database queries, ML inference,
file processing) with a random delay, then returns a JSON payload.

Graceful shutdown strategy:
  1. preStop hook sleeps 5s so kube-proxy can remove this pod from
     the Service endpoints (iptables / IPVS rules).
  2. gunicorn master receives SIGTERM from kubelet, forwards it to
     worker processes, and waits up to --graceful-timeout seconds
     for them to finish in-flight requests.
  3. The application-level `_draining` flag causes the /health probe
     to return 503, which tells K8s to stop sending new traffic.
"""

import logging
import os
import random
import signal
import sys
import threading
import time

from flask import Flask, jsonify, request

app = Flask(__name__)
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(process)d] %(levelname)s %(message)s",
)
logger = logging.getLogger(__name__)

# --- shutdown coordination ---
# A threading.Event is safer than a bare bool when multiple gunicorn
# workers (separate processes) each import this module independently.
_drain_event = threading.Event()


def _is_draining():
    return _drain_event.is_set()


@app.route("/process")
def process_work():
    """
    Simulate a unit of backend work.

    The delay is randomised between 150ms and 700ms to approximate
    the kind of latency you would see from a real database call or
    an external API round-trip.
    """
    if _is_draining():
        return jsonify({"error": "service is draining"}), 503

    work_ms = random.randint(150, 700)
    time.sleep(work_ms / 1000.0)

    hostname = os.environ.get("HOSTNAME", "unknown")
    return jsonify({
        "service": "python-downstream",
        "hostname": hostname,
        "work_ms": work_ms,
        "processed_at": time.strftime("%Y-%m-%dT%H:%M:%S%z"),
    })


@app.route("/health")
def health():
    """
    Readiness and liveness probe target.

    Returns 503 once the drain flag is set, which makes K8s mark the
    pod as NotReady and stop routing traffic to it.
    """
    if _is_draining():
        return "draining", 503
    return "ok", 200


@app.route("/prestop")
def prestop():
    """
    preStop lifecycle hook handler.

    Called by kubelet before SIGTERM is sent. We sleep 5 seconds to
    give kube-proxy time to converge its routing rules, then flip
    the drain flag so health checks start failing.
    """
    logger.info("prestop hook invoked -- sleeping 5s for kube-proxy convergence")
    time.sleep(5)
    _drain_event.set()
    logger.info("prestop finished -- drain flag is set")
    return "done", 200


def _handle_sigterm(signum, frame):
    """
    Direct-run SIGTERM handler.

    When running under gunicorn, the master process intercepts SIGTERM
    and orchestrates a graceful worker shutdown via --graceful-timeout.
    This handler exists for the `flask run` / `python app.py` case
    where there is no process manager.
    """
    _drain_event.set()
    logger.info("SIGTERM caught -- shutting down")
    sys.exit(0)


signal.signal(signal.SIGTERM, _handle_sigterm)


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=5000, debug=False)
