#!/usr/bin/env bash
set -euo pipefail

# ==============================================================================
# Setup Dedicated k6 Load Tester Node
# ==============================================================================

echo "⚡ Setting up Dedicated k6 Load Tester Node..."

# 1. Update system & install dependencies
sudo apt update && sudo apt install -y git curl gnupg

# 2. Apply high-throughput kernel tuning (prevents client socket exhaustion)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
chmod +x "$SCRIPT_DIR/tune-kernel.sh"
sudo "$SCRIPT_DIR/tune-kernel.sh"

# 3. Install k6 official binary
if ! command -v k6 &> /dev/null; then
    echo "📦 Installing k6..."
    sudo gpg -k || true
    sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
    echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
    sudo apt update && sudo apt install -y k6
fi

echo "✅ k6 Load Tester Node is ready!"
echo "   Run a test: BASE_URL=\"http://<ALB_OR_APP_IP>:8080\" k6 run loadtests/e2e_flow.js"
