#!/usr/bin/env bash
set -euo pipefail

# ==============================================================================
# Setup Centralized Monitoring Node (Prometheus, Jaeger, Grafana)
# ==============================================================================

echo "[INFO] Setting up Centralized Monitoring Node..."

# 1. Update system & install Docker
sudo apt update && sudo apt install -y git curl ufw
if ! command -v docker &> /dev/null; then
    echo "[INFO] Installing Docker..."
    curl -fsSL https://get.docker.com | sudo sh
    sudo usermod -aG docker "$USER"
fi

# 2. Navigate to monitoring directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MONITORING_DIR="$SCRIPT_DIR/../monitoring"
cd "$MONITORING_DIR"

# 3. Launch Prometheus, Grafana, Jaeger
echo "[INFO] Launching Prometheus, Grafana, and Jaeger..."
docker compose -f docker-compose.monitoring.yml up -d

echo "[SUCCESS] Monitoring stack deployed successfully!"
echo "   - Prometheus: http://localhost:9090"
echo "   - Grafana:    http://localhost:3000 (admin/admin)"
echo "   - Jaeger UI:  http://localhost:16686"
echo "   - OTLP gRPC:  http://localhost:4317"
