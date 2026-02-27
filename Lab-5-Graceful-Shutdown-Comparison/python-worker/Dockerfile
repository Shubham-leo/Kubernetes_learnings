FROM python:3.12-slim

WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY app.py .

EXPOSE 5000

# gunicorn for production serving:
#   --workers 2: concurrent request handling (1 worker = 1 process)
#   --graceful-timeout 30: workers get 30s to finish in-flight requests after SIGTERM
#   --timeout 30: kill workers that don't respond within 30s (prevents zombies)
#   --access-logfile -: pipe access logs to stdout (visible in kubectl logs)
CMD ["gunicorn", \
     "--workers", "2", \
     "--graceful-timeout", "30", \
     "--timeout", "30", \
     "--access-logfile", "-", \
     "--bind", "0.0.0.0:5000", \
     "app:app"]
