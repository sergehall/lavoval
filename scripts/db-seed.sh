#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="${ROOT_DIR}/.env.local"
SEEDS_DIR="${ROOT_DIR}/infrastructure/db/seeds"

if [[ -f "${ENV_FILE}" ]]; then
  set -a
  # shellcheck disable=SC1090
  source "${ENV_FILE}"
  set +a
fi

if [[ -z "${DATABASE_URL:-}" ]]; then
  echo "DATABASE_URL is not set. Check ${ENV_FILE} or your environment." >&2
  exit 1
fi

shopt -s nullglob
for file in "${SEEDS_DIR}"/*.sql; do
  echo "seed $(basename "${file}")"
  psql "${DATABASE_URL}" -v ON_ERROR_STOP=1 -f "${file}"
done

echo "seed complete"
