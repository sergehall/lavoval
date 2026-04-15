#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

export LAVOVAL_ENV_PROFILE=production

exec bash "${ROOT_DIR}/scripts/db-migrate.sh"
