#!/usr/bin/env bash
# =============================================================================
# scripts/api/10_health.sh — Health check endpoints (no auth required)
#
# Covers:
#   GET /health/        Overall health
#   GET /health/ready   Readiness probe (DB connections, etc.)
#   GET /health/live    Liveness probe
#   GET /health/startup Startup probe
#
# Usage:
#   bash scripts/api/10_health.sh
# =============================================================================

source "$(dirname "$0")/00_config.sh"

hdr "1. GET /health/ — Overall health"
http "${HTTP_PLAIN[@]}" GET "${BASE_URL}/health/"

hdr "2. GET /health/ready — Readiness probe"
http "${HTTP_PLAIN[@]}" GET "${BASE_URL}/health/ready"

hdr "3. GET /health/live — Liveness probe"
http "${HTTP_PLAIN[@]}" GET "${BASE_URL}/health/live"

hdr "4. GET /health/startup — Startup probe"
http "${HTTP_PLAIN[@]}" GET "${BASE_URL}/health/startup"

hdr "Done"
