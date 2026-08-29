#!/usr/bin/env bash
set -euo pipefail

APP_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CURRENT_USER="$(whoami)"

echo "======================================================"
echo "    Provisioning Native xlink App Server              "
echo "    Project Path: $APP_DIR                            "
echo "    User:         $CURRENT_USER                       "
echo "======================================================"

# 1. Ensure .env file is present
if [ ! -f "$APP_DIR/.env" ]; then
    echo "[INFO] No .env found. Initializing default production .env template..."
    cat <<EOF > "$APP_DIR/.env"
SERVER_PORT=8080
SERVER_BASE_URL=http://localhost:8080
DB_URL=postgres://postgres:xLinkAdmin9678@10.0.1.30:5432/xlink?sslmode=disable
REDIS_URL=redis://10.0.1.30:6379/0
AUTH_JWT_SECRET=your-production-jwt-secret-key-32-chars-long
LOG_LEVEL=info
LOG_FORMAT=json
CACHE_L1_MAX_COST_MB=512
CACHE_L1_NUM_COUNTERS=10000000
PPROF_ENABLED=true
OTEL_EXPORTER_OTLP_ENDPOINT=10.0.3.50:4317
OTEL_SERVICE_NAME=xlink-api
OTEL_ENVIRONMENT=production
EOF
    echo "[WARN] Created $APP_DIR/.env. Please verify database and Redis IP addresses."
else
    echo "[INFO] Using existing .env file from $APP_DIR/.env"
fi

# 2. Apply Linux Kernel Socket & Network Performance Tuning (100K RPS)
echo "[INFO] Applying High-Throughput Linux Kernel Tuning..."
sudo tee /etc/sysctl.d/99-xlink.conf > /dev/null << 'EOF'
fs.file-max = 2097152
net.core.somaxconn = 65535
net.ipv4.tcp_max_syn_backlog = 65535
net.core.netdev_max_backlog = 100000
net.ipv4.ip_local_port_range = 1024 65535
net.ipv4.tcp_tw_reuse = 1
net.ipv4.tcp_fin_timeout = 15
net.ipv4.tcp_slow_start_after_idle = 0
net.core.rmem_max = 16777216
net.core.wmem_max = 16777216
EOF
sudo sysctl --system > /dev/null 2>&1 || true

# 3. Increase File Descriptor Limits to 1 Million
sudo tee /etc/security/limits.d/99-xlink.conf > /dev/null << EOF
*    soft nofile 1048576
*    hard nofile 1048576
root soft nofile 1048576
root hard nofile 1048576
$CURRENT_USER soft nofile 1048576
$CURRENT_USER hard nofile 1048576
EOF

# 4. Install Dependencies & Nginx
echo "[INFO] Installing build dependencies and Nginx..."
sudo apt-get update -y
sudo apt-get install -y nginx curl git build-essential ca-certificates jq

# 5. Install Golang if not present
if ! command -v go &>/dev/null; then
    echo "[INFO] Installing latest Golang..."
    GO_ARCH="amd64"
    if [ "$(uname -m)" = "aarch64" ] || [ "$(uname -m)" = "arm64" ]; then
        GO_ARCH="arm64"
    fi
    GO_VERSION=$(curl -s "https://go.dev/VERSION?m=text" | head -n 1)
    curl -fsSL "https://go.dev/dl/${GO_VERSION}.linux-${GO_ARCH}.tar.gz" -o /tmp/go.tar.gz
    sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf /tmp/go.tar.gz
    rm -f /tmp/go.tar.gz
    
    echo "export PATH=\$PATH:/usr/local/go/bin:\$HOME/go/bin" | sudo tee /etc/profile.d/golang.sh > /dev/null
    export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
fi
echo "[INFO] $(go version)"

# 6. Build the Native Go Binary
echo "[INFO] Building xlink binary..."
cd "$APP_DIR"
mkdir -p bin
go build -ldflags="-w -s" -o "$APP_DIR/bin/xlink" ./cmd/api
chmod +x "$APP_DIR/bin/xlink"

# 7. Configure Nginx (65,535 Worker Connections + 1,000 Upstream Keepalives)
echo "[INFO] Configuring High-Performance Nginx reverse proxy..."
sudo tee /etc/nginx/nginx.conf > /dev/null << 'EOF'
user www-data;
worker_processes auto;
worker_rlimit_nofile 1048576;
pid /run/nginx.pid;

events {
    worker_connections 65535;
    multi_accept on;
    use epoll;
}

http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    keepalive_requests 100000;
    types_hash_max_size 2048;
    server_tokens off;

    client_body_buffer_size 128k;
    client_max_body_size 10m;
    client_header_buffer_size 1k;
    large_client_header_buffers 4 4k;

    access_log /var/log/nginx/access.log combined buffer=64k flush=5s;
    error_log /var/log/nginx/error.log warn;

    upstream xlink_backend {
        server 127.0.0.1:8080 max_fails=3 fail_timeout=10s;
        keepalive 1000;
    }

    server {
        listen 80 default_server reuseport;
        listen [::]:80 default_server reuseport;
        server_name _;

        location / {
            proxy_pass http://xlink_backend;
            proxy_http_version 1.1;
            proxy_set_header Connection "";
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;

            proxy_connect_timeout 3s;
            proxy_read_timeout 10s;
            proxy_send_timeout 10s;
        }

        location /readyz {
            proxy_pass http://xlink_backend/readyz;
            proxy_http_version 1.1;
            proxy_set_header Connection "";
        }
    }
}
EOF

sudo nginx -t
sudo systemctl restart nginx
sudo systemctl enable nginx

# 8. Configure & Start Systemd Service for xlink
echo "[INFO] Configuring xlink systemd service..."
sudo tee /etc/systemd/system/xlink.service > /dev/null << EOF
[Unit]
Description=xlink High-Performance URL Shortener API
After=network.target

[Service]
Type=simple
User=$CURRENT_USER
WorkingDirectory=$APP_DIR
ExecStart=$APP_DIR/bin/xlink
Restart=always
RestartSec=3s
LimitNOFILE=1048576
EnvironmentFile=$APP_DIR/.env

StandardOutput=journal
StandardError=journal
SyslogIdentifier=xlink-api

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable xlink
sudo systemctl restart xlink

echo "======================================================"
echo "[SUCCESS] xlink Server is LIVE and RUNNING!"
echo "======================================================"
sudo systemctl status xlink --no-pager -n 5 || true
echo ""
echo "Verify health:"
echo "   curl http://localhost/readyz"
echo "======================================================"
