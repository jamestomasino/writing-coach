#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${repo_root}"

go test ./cmd/... ./internal/...
go run ./cmd/objective-rule-lint
go run ./cmd/objective-score-eval
./scripts/test-pedagogy-integrity.sh
