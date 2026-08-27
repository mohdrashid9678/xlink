#!/usr/bin/env bash
set -euo pipefail

# ==============================================================================
# Linux Kernel Tuning Script for High-Throughput HTTP (100K+ RPS)
# ==============================================================================

echo "⚙️ Applying High-Throughput Linux Kernel Tuning..."

# 1. Increase System-Wide File Descriptor Limits (1 Million)
cat <<EOF | sudo tee /etc/security/limits.d/99-xlink.conf
*    soft nofile 1048576
*    hard nofile 1048576
root soft nofile 1048576
root hard nofile 1048576
EOF

# 2. Optimize TCP/IP Network Stack
cat <<EOF | sudo tee /etc/sysctl.d/99-xlink.conf
# Maximum number of open files
fs.file-max = 2097152

# Maximum socket receive/send buffer sizes
net.core.rmem_max = 16777216
net.core.wmem_max = 16777216
net.core.rmem_default = 1048576
net.core.wmem_default = 1048576

# Maximum number of packets queued on the input side
net.core.netdev_max_backlog = 100000

# Maximum number of pending TCP connection requests
net.core.somaxconn = 65535

# Expand ephemeral port range to prevent port exhaustion
net.ipv4.ip_local_port_range = 1024 65535

# Enable TCP TIME_WAIT socket reuse for outgoing connections
net.ipv4.tcp_tw_reuse = 1

# Reduce FIN timeout to free up closed sockets quickly
net.ipv4.tcp_fin_timeout = 15

# Disable slow start after idle to maintain high throughput
net.ipv4.tcp_slow_start_after_idle = 0

# Maximum TCP SYN backlog
net.ipv4.tcp_max_syn_backlog = 65535

# Maximum TCP TIME_WAIT buckets
net.ipv4.tcp_max_tw_buckets = 1440000
EOF

# Apply sysctl parameters immediately
sudo sysctl --system

echo "✅ Kernel tuning applied successfully for 100K+ RPS!"
