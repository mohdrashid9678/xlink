# AWS Production Deployment Guide (Zero Manual Setup)

This guide walks through deploying **xlink** on AWS EC2 instances inside your VPC (`10.0.0.0/22`).

---

## 1. AWS Security Groups (AWS Console)

Create 3 Security Groups inside your VPC:

1. **`sg-xlink-app`** (Attached to App Server EC2):
   - Inbound: `HTTP (80)` from `0.0.0.0/0`
   - Inbound: `HTTPS (443)` from `0.0.0.0/0`
   - Inbound: `SSH (22)` from `Your-IP/32` (or tester EC2)
2. **`sg-xlink-db`** (Attached to PostgreSQL EC2):
   - Inbound: `5432` from `sg-xlink-app` (Reference Security Group ID)
   - Inbound: `22` from `sg-xlink-app`
3. **`sg-xlink-redis`** (Attached to Redis EC2):
   - Inbound: `6379` from `sg-xlink-app` (Reference Security Group ID)
   - Inbound: `22` from `sg-xlink-app`

---

## 2. Automated Deployment Steps

### Machine 1: PostgreSQL 16 (Private Subnet)

SSH into your DB instance (via App Server as bastion jump host):
```bash
ssh -J ubuntu@<APP_PUBLIC_IP> ubuntu@<DB_PRIVATE_IP>
```

Run the automated DB provisioner:
```bash
git clone https://github.com/mohdrashid9678/xlink.git
cd xlink
git checkout feature/aws-self-hosted-deployment
chmod +x deploy/scripts/setup-db.sh
./deploy/scripts/setup-db.sh
exit
```
*(Copy the printed PostgreSQL connection string/IP).*

---

### Machine 2: Redis 7 (Private Subnet)

SSH into your Redis instance:
```bash
ssh -J ubuntu@<APP_PUBLIC_IP> ubuntu@<REDIS_PRIVATE_IP>
```

Run the automated Redis provisioner:
```bash
git clone https://github.com/mohdrashid9678/xlink.git
cd xlink
git checkout feature/aws-self-hosted-deployment
chmod +x deploy/scripts/setup-redis.sh
./deploy/scripts/setup-redis.sh
exit
```
*(Copy the printed Redis IP).*

---

### Machine 3: xlink App Server (Public Subnet)

SSH into your App Server:
```bash
ssh ubuntu@<APP_PUBLIC_IP>
```

Run the automated App setup script:
```bash
git clone https://github.com/mohdrashid9678/xlink.git
cd xlink
git checkout feature/aws-self-hosted-deployment
chmod +x deploy/scripts/setup-app.sh
./deploy/scripts/setup-app.sh
```

The script will automatically:
1. Apply Linux kernel socket tuning (`sysctl`).
2. Install Golang & compile the binary into `bin/xlink`.
3. Configure high-concurrency Nginx reverse proxy on port 80.
4. Prompt for your DB & Redis private IPs (or use environment variables) and generate a secure `.env`.
5. Register and start the `xlink.service` systemd service running in-place.

---

## 3. Verify Live API

```bash
# 1. Health Probe
curl -i http://<APP_PUBLIC_IP>/api/v1/health

# 2. Check Service Logs
sudo journalctl -u xlink -f
```
