#!/bin/bash
# ==============================================================================
# Oracle Engine - Rollback Script
# This script rolls back to a previous version
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

# Configuration
DEPLOY_HOST="${DEPLOY_HOST:-}"
DEPLOY_USER="${DEPLOY_USER:-oracle}"
DEPLOY_DIR="${DEPLOY_DIR:-/opt/oracle_engine}"
ROLLBACK_VERSION="${1:-}"
COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.prod.yml}"

if [[ -z "$DEPLOY_HOST" ]]; then
    log_error "DEPLOY_HOST is not set"
    exit 1
fi

if [[ -z "$ROLLBACK_VERSION" ]]; then
    log_error "Usage: $0 <version>"
    log_error "Example: $0 v1.2.3"
    exit 1
fi

log_warn "Rolling back to version: ${ROLLBACK_VERSION}"
log_warn "This will stop the current deployment and start the previous version"
read -p "Are you sure you want to continue? (y/N) " -n 1 -r
echo

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    log_info "Rollback cancelled"
    exit 0
fi

log_info "Starting rollback to ${ROLLBACK_VERSION}..."

ssh "${DEPLOY_USER}@${DEPLOY_HOST}" << ENDSSH
    set -euo pipefail
    
    cd "${DEPLOY_DIR}"
    
    echo "Stopping current containers..."
    docker compose -f ${COMPOSE_FILE} down
    
    echo "Updating image tag to ${ROLLBACK_VERSION}..."
    # Update the image tag in the compose file or pull specific version
    export IMAGE_TAG="${ROLLBACK_VERSION}"
    
    echo "Pulling version ${ROLLBACK_VERSION}..."
    docker pull ghcr.io/ifa-labs/oracle_engine:${ROLLBACK_VERSION}
    
    # Tag the rollback version as latest for the compose file
    docker tag ghcr.io/ifa-labs/oracle_engine:${ROLLBACK_VERSION} ghcr.io/ifa-labs/oracle_engine:latest
    
    echo "Starting containers with rolled back version..."
    docker compose -f ${COMPOSE_FILE} up -d
    
    echo "Waiting for health checks..."
    sleep 10
    
    echo "Checking container status..."
    docker compose -f ${COMPOSE_FILE} ps
ENDSSH

log_info "Rollback to ${ROLLBACK_VERSION} completed!"
