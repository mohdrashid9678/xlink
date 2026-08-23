#!/usr/bin/env bash
set -euo pipefail

echo "======================================================"
echo "         Provisioning Redis 7 Cache Server            "
echo "======================================================"

# 1. Apply Linux Kernel Socket & Network Performance Tuning
echo "Applying Linux kernel network tuning..."
sudo tee /etc/sysctl.d/99-xlink-redis.conf > /dev/null << 'EOF'
net.core.somaxconn = 65535
net.ipv4.tcp_max_syn_backlog = 65535
net.core.netdev_max_backlog = 65535
net.ipv4.ip_local_port_range = 1024 65535
net.ipv4.tcp_tw_reuse = 1
net.ipv4.tcp_fin_timeout = 15
vm.overcommit_memory = 1
EOF
sudo sysctl --system > /dev/null 2>&1 || true

# 2. Install Prerequisites
sudo apt-get update -y
sudo apt-get install -y lsb-release curl gpg ca-certificates

# 3. Add official Redis APT Repository
curl -fsSL https://packages.redis.io/gpg | sudo gpg --dearmor -o /usr/share/keyrings/redis-archive-keyring.gpg --yes
echo "deb [signed-by=/usr/share/keyrings/redis-archive-keyring.gpg] https://packages.redis.io/deb $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/redis.list

sudo apt-get update -y
sudo apt-get install -y redis

REDIS_CONF="/etc/redis/redis.conf"

# 4. Configure Redis Network Binding
sudo sed -i "s/^bind .*/bind 0.0.0.0/g" "$REDIS_CONF"
sudo sed -i "s/^protected-mode yes/protected-mode no/g" "$REDIS_CONF"

# 5. Performance Tuning for Caching (1GB maxmemory with LRU eviction)
if ! grep -q "maxmemory-policy allkeys-lru" "$REDIS_CONF"; then
    echo "maxmemory 1073741824" | sudo tee -a "$REDIS_CONF"
    echo "maxmemory-policy allkeys-lru" | sudo tee -a "$REDIS_CONF"
fi

# 6. Enable and Restart Redis
sudo systemctl enable redis-server
sudo systemctl restart redis-server

PRIVATE_IP=$(ip -4 addr show | grep -oP '(?<=inet\s)\d+(\.\d+){3}' | grep -v '127.0.0.1' | head -n 1)

echo "======================================================"
echo " ✅ Redis 7 Provisioning Complete!"
echo " Connection string:"
echo " redis://$PRIVATE_IP:6379/0"
echo "======================================================"
