#!/usr/bin/env bash
set -euo pipefail

APP_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
CURRENT_USER="$(whoami)"

echo "======================================================"
echo "    Provisioning xlink App Server & Environment       "
echo "    Project Path: $APP_DIR                            "
echo "    Running User: $CURRENT_USER                       "
echo "======================================================"

# 1. Apply Linux Kernel Network Performance Tuning
echo "Applying Linux kernel network tuning..."
sudo tee /etc/sysctl.d/99-xlink-app.conf > /dev/null << 'EOF'
net.core.somaxconn = 65535
net.ipv4.tcp_max_syn_backlog = 65535
net.core.netdev_max_backlog = 65535
net.ipv4.ip_local_port_range = 1024 65535
net.ipv4.tcp_tw_reuse = 1
net.ipv4.tcp_fin_timeout = 15
net.ipv4.tcp_rmem = 4096 87380 16777216
net.ipv4.tcp_wmem = 4096 65536 16777216
EOF
sudo sysctl --system > /dev/null 2>&1 || true

# 2. Install System Dependencies & Nginx
sudo apt-get update -y
sudo apt-get install -y nginx curl git build-essential ca-certificates jq

# 3. Install Golang (if not already installed)
if ! command -v go &>/dev/null; then
    echo "Installing latest Golang..."
    GO_ARCH="amd64"
    if [ "$(uname -m)" = "aarch64" ]; then
        GO_ARCH="arm64"
    fi
    GO_VERSION=$(curl -s "https://go.dev/VERSION?m=text" | head -n 1)
    curl -fsSL "https://go.dev/dl/${GO_VERSION}.linux-${GO_ARCH}.tar.gz" -o /tmp/go.tar.gz
    sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf /tmp/go.tar.gz
    rm -f /tmp/go.tar.gz
    
    echo "export PATH=\$PATH:/usr/local/go/bin:\$HOME/go/bin" | sudo tee /etc/profile.d/golang.sh > /dev/null
    export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
fi
echo "Go version: $(go version)"

# 4. Tune & Configure Nginx
echo "Configuring high-performance Nginx reverse proxy..."
sudo tee /etc/nginx/nginx.conf > /dev/null << 'EOF'
user www-data;
worker_processes auto;
pid /run/nginx.pid;
worker_rlimit_nofile 65535;

events {
    worker_connections 20000;
    multi_accept on;
    use epoll;
}

http {
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;

    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    access_log off;
    error_log /var/log/nginx/error.log crit;

    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types application/json text/plain text/css application/javascript;

    include /etc/nginx/conf.d/*.conf;
    include /etc/nginx/sites-enabled/*;
}
EOF

if [ -f "$APP_DIR/deploy/nginx/xlink.conf" ]; then
    sudo cp "$APP_DIR/deploy/nginx/xlink.conf" /etc/nginx/sites-available/xlink
    sudo ln -sf /etc/nginx/sites-available/xlink /etc/nginx/sites-enabled/default
    sudo nginx -t
    sudo systemctl restart nginx
fi

# 5. Build the Go binary directly inside project directory
echo "Building xlink binary..."
cd "$APP_DIR"
mkdir -p bin
go build -ldflags="-w -s" -o "$APP_DIR/bin/xlink" ./cmd/api
chmod +x "$APP_DIR/bin/xlink"

# 6. Configure .env file if not present
if [ ! -f "$APP_DIR/.env" ]; then
    echo "Creating production .env configuration..."
    
    DB_IP="${DB_PRIVATE_IP:-}"
    if [ -z "$DB_IP" ]; then
        read -rp "Enter PostgreSQL Private IP (e.g. 10.0.1.16): " DB_IP
    fi

    REDIS_IP="${REDIS_PRIVATE_IP:-}"
    if [ -z "$REDIS_IP" ]; then
        read -rp "Enter Redis Private IP (e.g. 10.0.1.214): " REDIS_IP
    fi

    # Generate a cryptographically secure 64-character secret
    JWT_SECRET=$(head -c 32 /dev/urandom | base64 | tr -dc 'a-zA-Z0-9' | head -c 48)

    cat <<EOF > "$APP_DIR/.env"
PORT=8080
GIN_MODE=release
AUTO_MIGRATE=true
DB_URL=postgresql://postgres:xLinkAdmin9678@${DB_IP}:5432/xlink?sslmode=disable
REDIS_URL=redis://${REDIS_IP}:6379/0
AUTH_JWT_SECRET=${JWT_SECRET}
LOG_LEVEL=info
LOG_FORMAT=json
EOF
    chmod 600 "$APP_DIR/.env"
    echo "Created .env with secure JWT secret."
fi

# 7. Configure & Install Systemd Service pointing directly to this directory
echo "Configuring systemd service..."
sudo tee /etc/systemd/system/xlink.service > /dev/null << EOF
[Unit]
Description=xlink High-Performance URL Shortener API
After=network.target

[Service]
Type=simple
User=${CURRENT_USER}
WorkingDirectory=${APP_DIR}
ExecStart=${APP_DIR}/bin/xlink
Restart=always
RestartSec=5s
LimitNOFILE=65535
EnvironmentFile=${APP_DIR}/.env

# Sandboxing
ProtectSystem=full
NoNewPrivileges=true
PrivateTmp=true

# Structured logs to systemd journal
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
echo " ✅ xlink App Server Setup Complete!"
echo " Service Status:"
sudo systemctl status xlink --no-pager
echo "======================================================"
echo " Test with: curl http://localhost:8080/api/v1/health"
echo " Live logs: sudo journalctl -u xlink -f"
echo "======================================================"
