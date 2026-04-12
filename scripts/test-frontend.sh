#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${repo_root}/web"

if [[ ! -d node_modules ]]; then
  npm ci
fi

npm run lint
npm run check:i18n-timezone
npm run check:skill-details
npm run build
