#!/usr/bin/env bash
set -euo pipefail

BACKUP_DIR="${BACKUP_DIR:-./backups}"
RETENTION_DAYS="${RETENTION_DAYS:-7}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-wtrlab}"
DB_NAME="${DB_NAME:-wtrlab}"
DB_PASSWORD="${DB_PASSWORD:-wtrlab_secret}"

mkdir -p "$BACKUP_DIR"

export PGPASSWORD="$DB_PASSWORD"

pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
  --format=custom \
  --compress=9 \
  --file="$BACKUP_DIR/${DB_NAME}_${TIMESTAMP}.dump"

echo "Backup created: $BACKUP_DIR/${DB_NAME}_${TIMESTAMP}.dump"

find "$BACKUP_DIR" -name "${DB_NAME}_*.dump" -mtime +$RETENTION_DAYS -delete
