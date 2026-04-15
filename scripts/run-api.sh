#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PROFILE="${LAVOVAL_ENV_PROFILE:-${1:-dev}}"

# shellcheck disable=SC1091
source "${ROOT_DIR}/scripts/load-env.sh"
load_lavoval_env "${ROOT_DIR}" "${PROFILE}"

if [[ "${PROFILE}" == "prod-local" ]]; then
  export APP_ENV=production
fi

cd "${ROOT_DIR}/apps/api"
go run ./cmd/server
