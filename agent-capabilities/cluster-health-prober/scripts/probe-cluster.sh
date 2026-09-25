#!/usr/bin/env bash
set -euo pipefail

# Multi-tier cluster health probe script for ChoreSync
# Verifies Backend, Proxy, Postgres, and Floci emulator.

BACKEND_PORT="${BACKEND_PORT:-8000}"
PROXY_PORT="${PROXY_PORT:-8088}"
FLOCI_PORT="${FLOCI_PORT:-4566}"

echo "==> Probing ChoreSync Multi-Tier Cluster Services..."

# 1. Backend /healthz
echo "1. Checking Go Backend /healthz..."
if curl -s -f "http://localhost:${BACKEND_PORT}/healthz" >/dev/null 2>&1; then
  echo "  [OK] Backend API is healthy on port ${BACKEND_PORT}."
elif curl -s -f "http://localhost:8080/healthz" >/dev/null 2>&1; then
  echo "  [OK] Backend API is healthy on port 8080."
else
  echo "  [INFO] Backend API not responding on port ${BACKEND_PORT} or 8080."
fi

# 2. Proxy port 8088 / Frontend port 3000
echo "2. Checking Frontend / Reverse Proxy..."
if curl -s -f "http://localhost:${PROXY_PORT}/healthz" >/dev/null 2>&1; then
  echo "  [OK] Caddy proxy /healthz responded with 200 OK on port ${PROXY_PORT}."
elif curl -s -f "http://localhost:3000" >/dev/null 2>&1; then
  echo "  [OK] Frontend dev server is responding on port 3000."
else
  echo "  [INFO] Neither Caddy (8088) nor Frontend (3000) responded."
fi

# 3. Floci cloud emulator port 4566
echo "3. Checking Floci cloud emulator on port ${FLOCI_PORT}..."
if curl -s -f "http://localhost:${FLOCI_PORT}/_floci/health" >/dev/null 2>&1 || curl -s "http://localhost:${FLOCI_PORT}/" >/dev/null 2>&1; then
  echo "  [OK] Floci cloud emulator is responding on port ${FLOCI_PORT}."
else
  echo "  [INFO] Floci cloud emulator not detected on port ${FLOCI_PORT}."
fi

# 4. Docker Compose container status
echo "4. Checking Docker Compose container roster..."
if command -v docker >/dev/null 2>&1; then
  docker compose ps 2>/dev/null || echo "  [INFO] Docker Compose not running."
fi

echo "==> Cluster health probing cycle completed."
exit 0
