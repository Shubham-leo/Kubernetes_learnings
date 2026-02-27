# Slim variant cuts the image from ~900MB (full) to ~150MB
FROM python:3.13-slim

WORKDIR /app

# Install dependencies first so this layer is cached when only
# app code changes (Docker layer caching optimisation).
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY app.py .

EXPOSE 5000

# gunicorn is the production process manager:
#   --workers 2         two worker processes for concurrency
#   --graceful-timeout 30  after SIGTERM, each worker gets 30s to finish
#   --timeout 30        kill any worker stuck longer than 30s on one request
#   --access-logfile -  print access logs to stdout (picked up by kubectl logs)
#   --bind 0.0.0.0:5000 listen on all interfaces inside the container
CMD ["gunicorn", \
     "--workers", "2", \
     "--graceful-timeout", "30", \
     "--timeout", "30", \
     "--access-logfile", "-", \
     "--bind", "0.0.0.0:5000", \
     "app:app"]
