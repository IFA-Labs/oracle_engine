#!/bin/bash
# ==============================================================================
# Oracle Engine - Database Backup Script
# Run this script to backup the TimescaleDB database
# ==============================================================================

set -euo pipefail

# Configuration
BACKUP_DIR="${BACKUP_DIR:-/var/www/oracle_engine/backups}"
CONTAINER_NAME="${CONTAINER_NAME:-oracle-timescale}"
POSTGRES_USER="${POSTGRES_USER:-oracle}"
POSTGRES_DB="${POSTGRES_DB:-oracle_engine}"
RETENTION_DAYS="${RETENTION_DAYS:-7}"

# Generate backup filename with timestamp
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="${BACKUP_DIR}/oracle_engine_${TIMESTAMP}.sql.gz"

echo "[INFO] Starting database backup..."
echo "[INFO] Backup file: ${BACKUP_FILE}"

# Create backup directory if it doesn't exist
mkdir -p "${BACKUP_DIR}"

# Create backup
docker exec "${CONTAINER_NAME}" pg_dump -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" | gzip > "${BACKUP_FILE}"

# Verify backup was created
if [[ -f "${BACKUP_FILE}" ]] && [[ -s "${BACKUP_FILE}" ]]; then
    BACKUP_SIZE=$(du -h "${BACKUP_FILE}" | cut -f1)
    echo "[INFO] Backup created successfully: ${BACKUP_FILE} (${BACKUP_SIZE})"
else
    echo "[ERROR] Backup failed - file is empty or not created"
    exit 1
fi

# Clean up old backups
echo "[INFO] Cleaning up backups older than ${RETENTION_DAYS} days..."
find "${BACKUP_DIR}" -name "oracle_engine_*.sql.gz" -mtime +${RETENTION_DAYS} -delete

# List remaining backups
echo "[INFO] Current backups:"
ls -lah "${BACKUP_DIR}"/oracle_engine_*.sql.gz 2>/dev/null || echo "No backups found"

echo "[INFO] Backup completed successfully!"
