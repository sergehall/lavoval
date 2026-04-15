#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PROFILE="${LAVOVAL_ENV_PROFILE:-${1:-dev}}"
MODE="${2:-dev}"

# shellcheck disable=SC1091
source "${ROOT_DIR}/scripts/load-env.sh"
load_lavoval_env "${ROOT_DIR}" "${PROFILE}"

if [[ "${PROFILE}" == "prod-local" || "${MODE}" == "prod" ]]; then
  export APP_ENV=production
fi

cd "${ROOT_DIR}"

if [[ "${MODE}" == "prod" ]]; then
  pnpm --dir apps/web build
  pnpm --dir apps/web start
  exit 0
fi

pnpm --dir apps/web dev
