#!/usr/bin/env bash
set -euo pipefail

SOURCE_PROJECT="${1:-${COMPOSE_PROJECT_NAME:-writing-coach}}"
TARGET_PROJECT="${2:-writing-coach-staging}"

copy_volume() {
  local suffix="$1"
  local source_volume="${SOURCE_PROJECT}_${suffix}"
  local target_volume="${TARGET_PROJECT}_${suffix}"

  docker volume inspect "${source_volume}" >/dev/null
  docker volume create "${target_volume}" >/dev/null

  docker run --rm \
    -v "${target_volume}:/to" \
    alpine:3.20 \
    sh -lc "find /to -mindepth 1 -delete"

  docker run --rm \
    -v "${source_volume}:/from:ro" \
    -v "${target_volume}:/to" \
    alpine:3.20 \
    sh -lc "cd /from && tar cf - . | tar xf - -C /to"

  echo "copied ${source_volume} -> ${target_volume}"
}

copy_volume "writing-coach-data"
copy_volume "kratos-data"
