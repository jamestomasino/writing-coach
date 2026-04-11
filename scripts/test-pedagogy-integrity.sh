#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${repo_root}"

echo "Running pedagogy integrity gates..."
go test ./internal/api -run 'ReviewSubmissionEmitsDecisionEvents|ReviewSubmissionClearsHoldAndEmitsTransitionEvent|PromptNext.*ProgressionHold|EmitProgressionHoldTransitionEventWritesActivation' -count=1
go test ./internal/db -run 'UpdateProgressionHoldStateActivatesAndClears|UpdateProgressionHoldStateIsEnrollmentScoped|SaveAndLoadDecisionEventsByReview' -count=1
