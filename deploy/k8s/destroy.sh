#!/usr/bin/env bash
set -euo pipefail

# ==============================================================================
# 1-Click Teardown of xlink Kubernetes Cluster
# ==============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "[INFO] Tearing down xlink Kubernetes resources..."
kubectl delete -f "$SCRIPT_DIR/07-hpa.yaml" --ignore-not-found
kubectl delete -f "$SCRIPT_DIR/06-app-service.yaml" --ignore-not-found
kubectl delete -f "$SCRIPT_DIR/05-app-deployment.yaml" --ignore-not-found
kubectl delete -f "$SCRIPT_DIR/04-jaeger.yaml" --ignore-not-found
kubectl delete -f "$SCRIPT_DIR/03-redis.yaml" --ignore-not-found
kubectl delete -f "$SCRIPT_DIR/02-postgres.yaml" --ignore-not-found
kubectl delete -f "$SCRIPT_DIR/01-configmap-secrets.yaml" --ignore-not-found
kubectl delete -f "$SCRIPT_DIR/00-namespace.yaml" --ignore-not-found

echo "[SUCCESS] xlink cluster teardown complete."
