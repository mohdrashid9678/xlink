#!/usr/bin/env bash
set -euo pipefail

echo "========================================="
echo "   Provisioning xlink App Server & Nginx "
echo "========================================="

# 1. Install Nginx
sudo apt-get update -y
sudo apt-get install -y nginx

# 2. Create system user and directory structure
sudo id -u xlink &>/dev/null || sudo useradd -r -s /bin/false -d /opt/xlink xlink
sudo mkdir -p /opt/xlink/bin /opt/xlink/logs
sudo chown -R xlink:xlink /opt/xlink

# 3. Configure Nginx
if [ -f "deploy/nginx/xlink.conf" ]; then
    sudo cp deploy/nginx/xlink.conf /etc/nginx/sites-available/xlink
    sudo ln -sf /etc/nginx/sites-available/xlink /etc/nginx/sites-enabled/default
    sudo nginx -t
    sudo systemctl restart nginx
fi

# 4. Install systemd service
if [ -f "deploy/systemd/xlink.service" ]; then
    sudo cp deploy/systemd/xlink.service /etc/systemd/system/xlink.service
    sudo systemctl daemon-reload
    sudo systemctl enable xlink
fi

echo "========================================="
echo " App server provisioning complete!"
echo " Copy binary to /opt/xlink/bin/xlink and environment to /opt/xlink/.env"
echo " Start service with: sudo systemctl start xlink"
echo "========================================="
