#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="${ROOT_DIR}/.env.local"
MIGRATIONS_DIR="${ROOT_DIR}/infrastructure/db/migrations"

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

echo "Waiting for database to be ready..."
for i in $(seq 1 30); do
  if psql "${DATABASE_URL}" -c '\q' 2>/dev/null; then
    echo "Database is ready."
    break
  fi
  echo "  attempt ${i}/30 — not ready yet, retrying in 1s..."
  sleep 1
done

if ! psql "${DATABASE_URL}" -c '\q' 2>/dev/null; then
  echo "Database did not become ready in time." >&2
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
    echo "skip ${version}"
    continue
  fi

  echo "apply ${version}"
  psql "${DATABASE_URL}" -v ON_ERROR_STOP=1 -f "${file}"
  psql "${DATABASE_URL}" -v ON_ERROR_STOP=1 -c "INSERT INTO schema_migrations (version) VALUES ('${version}');"
done

echo "migrations complete"
