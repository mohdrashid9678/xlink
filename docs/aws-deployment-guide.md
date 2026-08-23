# AWS Self-Hosted Production Deployment Guide

This guide walks through deploying **xlink** on self-hosted AWS EC2 instances inside your custom VPC (`10.0.0.0/22`).

---

## 1. Network & VPC Layout

| Component | CIDR / Subnet | Route Table Destination | Purpose |
| :--- | :--- | :--- | :--- |
| **VPC** | `10.0.0.0/22` | Local | Private VPC Network |
| **Public Subnet (AZ-A)** | `10.0.0.0/24` | `0.0.0.0/0` -> Internet Gateway (IGW) | **App Server EC2** (xlink API + Nginx) |
| **Private Subnet (AZ-A)** | `10.0.1.0/24` | Internal only (or NAT if outbound needed) | **PostgreSQL EC2 & Redis EC2** |

---

## 2. Security Groups

### A. App Server Security Group (`sg-xlink-app`)
Attach to the EC2 in the **Public Subnet**:
- **Inbound**:
  - `HTTP (80)`: Source `0.0.0.0/0`
  - `HTTPS (443)`: Source `0.0.0.0/0`
  - `SSH (22)`: Source `<Your-Local-IP>/32`
- **Outbound**:
  - All traffic to `0.0.0.0/0` (or restricted to VPC CIDR)

### B. Database Security Group (`sg-xlink-db`)
Attach to the PostgreSQL EC2 in the **Private Subnet**:
- **Inbound**:
  - `PostgreSQL (5432)`: Source `sg-xlink-app` (Reference by SG ID)
  - `SSH (22)`: Source `sg-xlink-app` (Use App server as Bastion Jump host)
- **Outbound**: All traffic

### C. Redis Security Group (`sg-xlink-redis`)
Attach to the Redis EC2 in the **Private Subnet**:
- **Inbound**:
  - `Redis (6379)`: Source `sg-xlink-app` (Reference by SG ID)
  - `SSH (22)`: Source `sg-xlink-app`
- **Outbound**: All traffic

---

## 3. Machine Setup Instructions

### Machine 1: PostgreSQL 16 (Private Subnet)
1. SSH to App Server, then SSH to DB private IP:
   ```bash
   ssh -J ubuntu@<APP_PUBLIC_IP> ubuntu@<DB_PRIVATE_IP>
   ```
2. Run database provisioner:
   ```bash
   git clone https://github.com/mohdrashid9678/xlink.git
   cd xlink
   chmod +x deploy/scripts/setup-db.sh
   DB_PASSWORD="YourStrongSecurePassword123!" ./deploy/scripts/setup-db.sh
   ```

---

### Machine 2: Redis 7 (Private Subnet)
1. SSH into Redis instance via bastion:
   ```bash
   ssh -J ubuntu@<APP_PUBLIC_IP> ubuntu@<REDIS_PRIVATE_IP>
   ```
2. Run Redis provisioner:
   ```bash
   git clone https://github.com/mohdrashid9678/xlink.git
   cd xlink
   chmod +x deploy/scripts/setup-redis.sh
   ./deploy/scripts/setup-redis.sh
   ```

---

### Machine 3: xlink App Server (Public Subnet)
1. SSH into the App Server:
   ```bash
   ssh ubuntu@<APP_PUBLIC_IP>
   ```
2. Clone repository & run App setup:
   ```bash
   git clone https://github.com/mohdrashid9678/xlink.git
   cd xlink
   chmod +x deploy/scripts/setup-app.sh
   ./deploy/scripts/setup-app.sh
   ```
3. Copy Linux binary into place:
   ```bash
   sudo cp bin/xlink-linux-arm64 /opt/xlink/bin/xlink   # For Graviton t4g/c7g
   # OR: sudo cp bin/xlink-linux-amd64 /opt/xlink/bin/xlink # For Intel/AMD
   sudo chmod +x /opt/xlink/bin/xlink
   ```
4. Create `/opt/xlink/.env`:
   ```bash
   sudo cp deploy/env.production.example /opt/xlink/.env
   sudo nano /opt/xlink/.env
   ```
   *Fill in `<DB_PRIVATE_IP>`, `<REDIS_PRIVATE_IP>`, database password, and `AUTH_JWT_SECRET`.*
   ```bash
   sudo chown xlink:xlink /opt/xlink/.env
   sudo chmod 600 /opt/xlink/.env
   ```
5. Start & Verify Service:
   ```bash
   sudo systemctl restart xlink
   sudo systemctl status xlink
   ```
6. Check Live Logs:
   ```bash
   journalctl -u xlink -f
   ```

---

## 4. End-to-End Verification

From your local terminal, test the live API via the App Server's Public IP or Domain:

```bash
# 1. Health Probe
curl -i http://<APP_PUBLIC_IP>/api/v1/health

# 2. Register Account
curl -i -X POST http://<APP_PUBLIC_IP>/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"Password123!","name":"Admin"}'

# 3. Create Short URL
curl -i -X POST http://<APP_PUBLIC_IP>/api/v1/urls \
  -H "Authorization: Bearer <ACCESS_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"long_url":"https://google.com","custom_alias":"google"}'

# 4. Test Redirect Resolution (Fast Cache Hit)
curl -i http://<APP_PUBLIC_IP>/google
```
