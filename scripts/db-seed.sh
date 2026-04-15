#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PROFILE="${LAVOVAL_ENV_PROFILE:-dev}"
SEEDS_DIR="${ROOT_DIR}/infrastructure/db/seeds"

# shellcheck disable=SC1091
source "${ROOT_DIR}/scripts/load-env.sh"
load_lavoval_env "${ROOT_DIR}" "${PROFILE}"

if [[ -z "${DATABASE_URL:-}" ]]; then
  echo "DATABASE_URL is not set. Check .env/.env.local or your environment." >&2
  exit 1
fi

shopt -s nullglob
for file in "${SEEDS_DIR}"/*.sql; do
  echo "seed $(basename "${file}")"
  psql "${DATABASE_URL}" -v ON_ERROR_STOP=1 -f "${file}"
done

echo "seed complete"
