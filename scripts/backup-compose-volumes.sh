#!/usr/bin/env bash
set -euo pipefail

PROJECT_NAME="${1:-${COMPOSE_PROJECT_NAME:-writing-coach}}"
BACKUP_ROOT="${2:-./backups}"
STAMP="$(date +%Y%m%d-%H%M%S)"
DEST_DIR="${BACKUP_ROOT%/}/${PROJECT_NAME}-${STAMP}"

mkdir -p "${DEST_DIR}"

backup_volume() {
  local suffix="$1"
  local volume="${PROJECT_NAME}_${suffix}"
  local archive="${DEST_DIR}/${suffix}.tgz"

  docker volume inspect "${volume}" >/dev/null
  docker run --rm \
    -v "${volume}:/from:ro" \
    -v "$(pwd)/${DEST_DIR}:/backup" \
    alpine:3.20 \
    sh -lc "cd /from && tar czf '/backup/${suffix}.tgz' ."

  echo "backed up ${volume} -> ${archive}"
}

backup_volume "writing-coach-data"
backup_volume "kratos-data"

echo "backup directory: ${DEST_DIR}"
