#!/usr/bin/env bash
set -euo pipefail

# ==============================================================================
# 1-Click Cloud-Native Kubernetes Deployment for xlink (100K RPS Cluster)
# ==============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$SCRIPT_DIR/../.."

echo "[INFO] Deploying xlink to Kubernetes..."

# 1. Ensure Namespace exists
echo "[STEP 1/5] Creating xlink namespace..."
kubectl apply -f "$SCRIPT_DIR/00-namespace.yaml"

# 2. Apply ConfigMap and Secrets
echo "[STEP 2/5] Applying ConfigMap and Secrets..."
kubectl apply -f "$SCRIPT_DIR/01-configmap-secrets.yaml"

# 3. Apply PostgreSQL and Redis StatefulSets
echo "[STEP 3/5] Deploying PostgreSQL 16 and Redis 7 StatefulSets..."
kubectl apply -f "$SCRIPT_DIR/02-postgres.yaml"
kubectl apply -f "$SCRIPT_DIR/03-redis.yaml"

# 4. Deploy Jaeger Distributed Tracing
echo "[STEP 4/5] Deploying Jaeger OpenTelemetry collector..."
kubectl apply -f "$SCRIPT_DIR/04-jaeger.yaml"

# 5. Wait for Databases to be Ready
echo "[INFO] Waiting for PostgreSQL and Redis to be ready..."
kubectl rollout status statefulset/postgres -n xlink --timeout=120s
kubectl rollout status statefulset/redis -n xlink --timeout=120s

# 6. Apply App Deployment & Service
echo "[STEP 5/5] Deploying xlink-api (3 replicas) and Service..."
kubectl apply -f "$SCRIPT_DIR/05-app-deployment.yaml"
kubectl apply -f "$SCRIPT_DIR/06-app-service.yaml"
kubectl apply -f "$SCRIPT_DIR/07-hpa.yaml"

# 7. Wait for App Pods Rollout
echo "[INFO] Waiting for xlink-api pods to become healthy..."
kubectl rollout status deployment/xlink-api -n xlink --timeout=120s

echo ""
echo "================================================================================"
echo "[SUCCESS] xlink Cloud-Native Kubernetes Cluster is 100% LIVE and HEALTHY!"
echo "================================================================================"
echo ""
kubectl get pods,svc,hpa -n xlink
echo ""
echo "To port-forward and test locally:"
echo "   kubectl port-forward svc/xlink-service 8080:8080 -n xlink"
echo "   kubectl port-forward svc/jaeger 16686:16686 -n xlink"
echo ""
echo "To run k6 saturation test:"
echo "   BASE_URL=\"http://localhost:8080\" k6 run loadtests/saturation_test.js"
echo "================================================================================"
