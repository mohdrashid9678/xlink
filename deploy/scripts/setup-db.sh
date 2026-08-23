#!/usr/bin/env bash
set -euo pipefail

echo "======================================================"
echo "    Provisioning PostgreSQL 16 Database Server        "
echo "======================================================"

# 1. Apply Linux Kernel Socket & Network Performance Tuning
echo "Applying Linux kernel network tuning..."
sudo tee /etc/sysctl.d/99-xlink-db.conf > /dev/null << 'EOF'
net.core.somaxconn = 65535
net.ipv4.tcp_max_syn_backlog = 65535
net.core.netdev_max_backlog = 65535
net.ipv4.ip_local_port_range = 1024 65535
net.ipv4.tcp_tw_reuse = 1
net.ipv4.tcp_fin_timeout = 15
EOF
sudo sysctl --system > /dev/null 2>&1 || true

# 2. Install Prerequisites
sudo apt-get update -y
sudo apt-get install -y curl ca-certificates gnupg lsb-release

# 3. Add official PostgreSQL APT Repository
sudo install -d /etc/apt/keyrings
curl -fsSL https://www.postgresql.org/media/keys/ACCC4CF8.asc | sudo gpg --dearmor -o /etc/apt/keyrings/postgresql.gpg --yes
echo "deb [signed-by=/etc/apt/keyrings/postgresql.gpg] http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" | sudo tee /etc/apt/sources.list.d/pgdg.list

sudo apt-get update -y
sudo apt-get install -y postgresql-16 postgresql-contrib-16 || sudo apt-get install -y postgresql postgresql-contrib

# 4. Start & Enable PostgreSQL Service
sudo systemctl enable postgresql
sudo systemctl start postgresql

# 5. Configure Database and Credentials
DB_NAME=${DB_NAME:-xlink}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-"xLinkAdmin9678"}

echo "Configuring PostgreSQL user '$DB_USER' and database '$DB_NAME'..."
sudo -u postgres psql -c "ALTER USER $DB_USER WITH PASSWORD '$DB_PASSWORD';"
sudo -u postgres psql -tc "SELECT 1 FROM pg_database WHERE datname = '$DB_NAME'" | grep -q 1 || sudo -u postgres psql -c "CREATE DATABASE $DB_NAME OWNER $DB_USER;"

# 6. Configure Networking: Listen on Private VPC Interfaces & Authorize CIDR 10.0.0.0/22
PG_CONF=$(sudo -u postgres psql -t -P format=unaligned -c 'SHOW config_file')
PG_HBA=$(sudo -u postgres psql -t -P format=unaligned -c 'SHOW hba_file')

# Ensure PostgreSQL listens on all interfaces (0.0.0.0 / *)
sudo sed -i "s/#listen_addresses = 'localhost'/listen_addresses = '*'/g" "$PG_CONF"
sudo sed -i "s/listen_addresses = 'localhost'/listen_addresses = '*'/g" "$PG_CONF"

# Authorize password logins from the VPC CIDR (10.0.0.0/22)
if ! grep -q "10.0.0.0/22" "$PG_HBA"; then
    echo "host    all             all             10.0.0.0/22             scram-sha-256" | sudo tee -a "$PG_HBA"
fi

# 7. Restart PostgreSQL to apply network changes
sudo systemctl restart postgresql

PRIVATE_IP=$(ip -4 addr show | grep -oP '(?<=inet\s)\d+(\.\d+){3}' | grep -v '127.0.0.1' | head -n 1)

echo "======================================================"
echo " PostgreSQL 16 Provisioning Complete!"
echo " Connection string:"
echo " postgresql://$DB_USER:$DB_PASSWORD@$PRIVATE_IP:5432/$DB_NAME?sslmode=disable"
echo "======================================================"
