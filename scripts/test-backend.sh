#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${repo_root}"

go test ./cmd/... ./internal/...
go run ./cmd/objective-rule-lint
./scripts/test-pedagogy-integrity.sh
