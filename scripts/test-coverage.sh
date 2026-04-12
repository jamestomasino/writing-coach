#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${repo_root}"

mapfile -t pkgs < <(go list -f '{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}' ./cmd/... ./internal/... | sed '/^$/d')
if [[ ${#pkgs[@]} -eq 0 ]]; then
  echo "No Go packages with tests found under ./cmd or ./internal"
  exit 1
fi

go test -cover "${pkgs[@]}"
