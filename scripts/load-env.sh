#!/usr/bin/env bash

load_lavoval_env() {
  local root_dir="${1:?root directory is required}"
  local profile="${2:-dev}"
  local base_env="${root_dir}/.env"
  local local_env="${root_dir}/.env.local"

  if [[ -f "${base_env}" ]]; then
    set -a
    # shellcheck disable=SC1090
    source "${base_env}"
    set +a
  fi

  case "${profile}" in
    dev|local|prod-local)
      if [[ -f "${local_env}" ]]; then
        set -a
        # shellcheck disable=SC1090
        source "${local_env}"
        set +a
      fi
      ;;
  esac
}
