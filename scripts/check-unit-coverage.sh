#!/usr/bin/env bash
set -euo pipefail

minimum=80

check_coverage() {
  local directory="$1"
  local package="$2"
  local result coverage

  result="$(cd "$directory" && go test -cover "$package")"
  printf '%s\n' "$result"
  coverage="$(printf '%s\n' "$result" | sed -n 's/.*coverage: \([0-9.]*\)%.*/\1/p' | tail -n 1)"
  if [[ -z "$coverage" ]] || ! awk -v got="$coverage" -v required="$minimum" 'BEGIN { exit !(got >= required) }'; then
    echo "Coverage for $directory $package is ${coverage:-unknown}%; required: ${minimum}%" >&2
    exit 1
  fi
}

# Entry points compose databases, queues, object storage and HTTP servers. Their
# behavior is exercised by the compose smoke build; this gate measures unit-test
# coverage for the deterministic domain code in each service.
check_coverage services/video-api ./internal/auth
check_coverage services/video-api ./internal/config
check_coverage services/video-api ./internal/metrics
check_coverage services/video-api ./internal/middleware
check_coverage services/video-api ./internal/video
check_coverage services/video-worker ./internal/processor
check_coverage services/notification-service ./internal/notification
