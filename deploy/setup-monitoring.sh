#!/usr/bin/env bash
set -euo pipefail

echo "======================================================"
echo "    Provisioning Centralized Monitoring Node          "
echo "======================================================"

# 1. Update system & install Docker
sudo apt update && sudo apt install -y git curl ca-certificates
if ! command -v docker &> /dev/null; then
    echo "[INFO] Installing Docker..."
    curl -fsSL https://get.docker.com | sudo sh
    sudo usermod -aG docker "$USER"
fi

# 2. Setup Dedicated Monitoring Directory
MONITOR_DIR="/opt/xlink-monitoring"
sudo mkdir -p "$MONITOR_DIR"
sudo chown -R "$USER:$USER" "$MONITOR_DIR"
cd "$MONITOR_DIR"

# 3. Create Prometheus Configuration
cat <<'EOF' > prometheus.yml
global:
  scrape_interval: 10s
  evaluation_interval: 10s

scrape_configs:
  - job_name: 'xlink_cluster'
    metrics_path: '/metrics'
    static_configs:
      - targets:
          - '10.0.1.10:8080'  # App Pod 1
          - '10.0.1.20:8080'  # App Pod 2
          - '10.0.3.10:8080'  # App Pod 3
        labels:
          app: 'xlink'
          env: 'production'
EOF

# 4. Create Docker Compose for Prometheus, Grafana, Jaeger
cat <<'EOF' > docker-compose.yml
services:
  prometheus:
    image: prom/prometheus:latest
    container_name: xlink-prometheus
    restart: unless-stopped
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.enable-lifecycle'

  grafana:
    image: grafana/grafana:latest
    container_name: xlink-grafana
    restart: unless-stopped
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_USER=admin
      - GF_SECURITY_ADMIN_PASSWORD=admin
      - GF_USERS_ALLOW_SIGN_UP=false
    volumes:
      - grafana_data:/var/lib/grafana
    depends_on:
      - prometheus

  jaeger:
    image: jaegertracing/all-in-one:latest
    container_name: xlink-jaeger
    restart: unless-stopped
    ports:
      - "16686:16686"
      - "4317:4317"
      - "4318:4318"
    environment:
      - COLLECTOR_OTLP_ENABLED=true

volumes:
  prometheus_data:
  grafana_data:
EOF

# 5. Launch the Stack
echo "[INFO] Launching Prometheus, Grafana, and Jaeger..."
docker compose up -d

echo "======================================================"
echo "[SUCCESS] Monitoring Stack is LIVE!"
echo "   - Prometheus: http://localhost:9090"
echo "   - Grafana:    http://localhost:3000 (admin/admin)"
echo "   - Jaeger UI:  http://localhost:16686"
echo "   - OTLP gRPC:  http://localhost:4317"
echo "======================================================"
