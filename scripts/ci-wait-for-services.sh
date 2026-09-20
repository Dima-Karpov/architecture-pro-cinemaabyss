#!/usr/bin/env bash
set -euo pipefail

TIMEOUT_SECONDS="${CI_WAIT_TIMEOUT:-300}"
DEADLINE=$((SECONDS + TIMEOUT_SECONDS))

wait_for() {
  local url=$1
  echo "Waiting for ${url}..."

  until curl -sf "${url}" > /dev/null; do
    if (( SECONDS >= DEADLINE )); then
      echo "Timeout after ${TIMEOUT_SECONDS}s waiting for ${url}"
      docker compose ps
      docker compose logs monolith movies-service events-service proxy-service kafka --tail=100
      exit 1
    fi
    sleep 5
  done

  echo "Ready: ${url}"
}

wait_for "http://localhost:8080/health"
wait_for "http://localhost:8081/api/movies/health"
wait_for "http://localhost:8082/api/events/health"
wait_for "http://localhost:8000/health"

echo "All services are ready."
