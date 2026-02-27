# Python slim image — ~150MB vs ~900MB for full python image
FROM python:3.12-slim

WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY app.py .

EXPOSE 5000

# Run with gunicorn for production-grade serving:
#   --workers 2: handle concurrent requests (1 worker = 1 process)
#   --graceful-timeout 30: after SIGTERM, workers have 30s to finish in-flight requests
#   --timeout 30: kill workers that don't respond within 30s (prevents zombie workers)
#   --bind 0.0.0.0:5000: listen on all interfaces
#
# IMPORTANT: gunicorn properly handles SIGTERM — it:
#   1. Stops accepting new connections
#   2. Sends SIGTERM to each worker
#   3. Waits up to --graceful-timeout for workers to finish
#   4. Force-kills workers that exceed the timeout
CMD ["gunicorn", "--workers", "2", "--graceful-timeout", "30", "--timeout", "30", "--bind", "0.0.0.0:5000", "app:app"]
