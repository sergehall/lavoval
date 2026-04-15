#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PROFILE="${LAVOVAL_ENV_PROFILE:-dev}"
MIGRATIONS_DIR="${ROOT_DIR}/infrastructure/db/migrations"

# shellcheck disable=SC1091
source "${ROOT_DIR}/scripts/load-env.sh"
load_lavoval_env "${ROOT_DIR}" "${PROFILE}"

if [[ -z "${DATABASE_URL:-}" ]]; then
  echo "DATABASE_URL is not set. Check .env/.env.local or your environment." >&2
  exit 1
fi

psql "${DATABASE_URL}" -v ON_ERROR_STOP=1 <<'SQL'
CREATE TABLE IF NOT EXISTS schema_migrations (
  version TEXT PRIMARY KEY,
  applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
SQL

shopt -s nullglob
for file in "${MIGRATIONS_DIR}"/*.sql; do
  version="$(basename "${file}")"

  if psql "${DATABASE_URL}" -tAc "SELECT 1 FROM schema_migrations WHERE version = '${version}'" | grep -q 1; then
    echo "applied ${version}"
  else
    echo "pending ${version}"
  fi
done
