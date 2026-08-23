#!/usr/bin/env bash
set -euo pipefail

echo "========================================="
echo "   Provisioning PostgreSQL 16 on Ubuntu   "
echo "========================================="

sudo apt-get update -y
sudo apt-get install -y curl ca-certificates gnupg lsb-release

# Add official PostgreSQL APT repository
sudo install -d /etc/apt/keyrings
curl -fsSL https://www.postgresql.org/media/keys/ACCC4CF8.asc | sudo gpg --dearmor -o /etc/apt/keyrings/postgresql.gpg --yes
echo "deb [signed-by=/etc/apt/keyrings/postgresql.gpg] http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" | sudo tee /etc/apt/sources.list.d/pgdg.list

sudo apt-get update -y
sudo apt-get install -y postgresql-16 postgresql-contrib-16 || sudo apt-get install -y postgresql postgresql-contrib

# Start and enable PostgreSQL
sudo systemctl enable postgresql
sudo systemctl start postgresql

# Read DB Credentials or use defaults
DB_NAME=${DB_NAME:-xlink}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-"xLinkAdmin9678"}

echo "Configuring database '$DB_NAME' and user '$DB_USER'..."

sudo -u postgres psql -c "ALTER USER $DB_USER WITH PASSWORD '$DB_PASSWORD';"
sudo -u postgres psql -tc "SELECT 1 FROM pg_database WHERE datname = '$DB_NAME'" | grep -q 1 || sudo -u postgres psql -c "CREATE DATABASE $DB_NAME OWNER $DB_USER;"

# Configure PostgreSQL to listen on all private network interfaces
PG_CONF=$(sudo -u postgres psql -t -P format=unaligned -c 'SHOW config_file')
PG_HBA=$(sudo -u postgres psql -t -P format=unaligned -c 'SHOW hba_file')

sudo sed -i "s/#listen_addresses = 'localhost'/listen_addresses = '*'/g" "$PG_CONF"
sudo sed -i "s/listen_addresses = 'localhost'/listen_addresses = '*'/g" "$PG_CONF"

# Allow password authentication from VPC CIDR 10.0.0.0/22
if ! grep -q "10.0.0.0/22" "$PG_HBA"; then
    echo "host    all             all             10.0.0.0/22             scram-sha-256" | sudo tee -a "$PG_HBA"
fi

sudo systemctl restart postgresql

echo "========================================="
echo " PostgreSQL setup complete!"
echo " Connection string: postgresql://$DB_USER:$DB_PASSWORD@<PRIVATE_IP>:5432/$DB_NAME?sslmode=disable"
echo "========================================="
