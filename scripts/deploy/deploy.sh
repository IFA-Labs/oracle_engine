#!/bin/bash
# ==============================================================================
# Oracle Engine - Deployment Script
# This script deploys the application to a server
# ==============================================================================

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Configuration - Override these via environment variables
DEPLOY_HOST="${DEPLOY_HOST:-}"
DEPLOY_USER="${DEPLOY_USER:-oracle}"
DEPLOY_DIR="${DEPLOY_DIR:-/var/www/oracle_engine}"
DOCKER_IMAGE="${DOCKER_IMAGE:-ghcr.io/ifa-labs/oracle_engine:latest}"
COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.prod.yml}"

# Validate required variables
if [[ -z "$DEPLOY_HOST" ]]; then
    log_error "DEPLOY_HOST is not set"
    exit 1
fi

log_info "Starting deployment to ${DEPLOY_HOST}..."

# ==============================================================================
# Pre-deployment checks
# ==============================================================================
log_info "Running pre-deployment checks..."

# Check SSH connection
if ! ssh -o ConnectTimeout=10 "${DEPLOY_USER}@${DEPLOY_HOST}" "echo 'SSH connection successful'" &>/dev/null; then
    log_error "Cannot connect to ${DEPLOY_HOST} via SSH"
    exit 1
fi

log_info "SSH connection verified"

# ==============================================================================
# Deploy
# ==============================================================================
log_info "Deploying to ${DEPLOY_HOST}..."

ssh "${DEPLOY_USER}@${DEPLOY_HOST}" << ENDSSH
    set -euo pipefail
    
    cd "${DEPLOY_DIR}"
    
    echo "Pulling latest images..."
    docker compose -f ${COMPOSE_FILE} pull
    
    echo "Stopping current containers..."
    docker compose -f ${COMPOSE_FILE} down --remove-orphans || true
    
    echo "Starting new containers..."
    docker compose -f ${COMPOSE_FILE} up -d
    
    echo "Cleaning up old images..."
    docker image prune -f
    
    echo "Waiting for health checks..."
    sleep 10
    
    echo "Checking container status..."
    docker compose -f ${COMPOSE_FILE} ps
    
    echo "Checking health endpoint..."
    curl -sf http://localhost:8000/health || echo "Health check failed (may need more time)"
ENDSSH

log_info "Deployment completed successfully!"

# ==============================================================================
# Post-deployment verification
# ==============================================================================
log_info "Running post-deployment verification..."

# Wait for services to be fully ready
sleep 5

# Check if the service is responding
if ssh "${DEPLOY_USER}@${DEPLOY_HOST}" "curl -sf http://localhost:8000/health" &>/dev/null; then
    log_info "Health check passed!"
else
    log_warn "Health check failed - service may still be starting"
fi

log_info "=========================================="
log_info "Deployment to ${DEPLOY_HOST} completed!"
log_info "=========================================="
