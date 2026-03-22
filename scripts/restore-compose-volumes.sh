#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 1 ]]; then
  echo "usage: $0 BACKUP_DIR [PROJECT_NAME]" >&2
  exit 1
fi

BACKUP_DIR="$1"
PROJECT_NAME="${2:-${COMPOSE_PROJECT_NAME:-writing-coach}}"

restore_volume() {
  local suffix="$1"
  local archive="${BACKUP_DIR%/}/${suffix}.tgz"
  local volume="${PROJECT_NAME}_${suffix}"

  test -f "${archive}"
  docker volume create "${volume}" >/dev/null

  docker run --rm \
    -v "${volume}:/to" \
    alpine:3.20 \
    sh -lc "find /to -mindepth 1 -delete"

  docker run --rm \
    -v "${volume}:/to" \
    -v "$(cd "$(dirname "${archive}")" && pwd):/backup:ro" \
    alpine:3.20 \
    sh -lc "cd /to && tar xzf '/backup/$(basename "${archive}")'"

  echo "restored ${archive} -> ${volume}"
}

restore_volume "writing-coach-data"
restore_volume "kratos-data"
