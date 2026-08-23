#!/usr/bin/env bash
set -euo pipefail

echo "========================================="
echo "       Provisioning Redis 7 on Ubuntu    "
echo "========================================="

sudo apt-get update -y
sudo apt-get install -y lsb-release curl gpg

curl -fsSL https://packages.redis.io/gpg | sudo gpg --dearmor -o /usr/share/keyrings/redis-archive-keyring.gpg --yes
echo "deb [signed-by=/usr/share/keyrings/redis-archive-keyring.gpg] https://packages.redis.io/deb $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/redis.list

sudo apt-get update -y
sudo apt-get install -y redis

REDIS_CONF="/etc/redis/redis.conf"

# Configure Redis to listen on all private network interfaces
sudo sed -i "s/^bind 127.0.0.1 ::1/bind 0.0.0.0/g" "$REDIS_CONF"
sudo sed -i "s/^protected-mode yes/protected-mode no/g" "$REDIS_CONF"

# Performance Tuning for Caching (1GB maxmemory with LRU eviction)
if ! grep -q "maxmemory-policy allkeys-lru" "$REDIS_CONF"; then
    echo "maxmemory 1073741824" | sudo tee -a "$REDIS_CONF"
    echo "maxmemory-policy allkeys-lru" | sudo tee -a "$REDIS_CONF"
fi

sudo systemctl enable redis-server
sudo systemctl restart redis-server

echo "========================================="
echo " Redis 7 setup complete!"
echo " Connection string: redis://<PRIVATE_IP>:6379/0"
echo "========================================="
