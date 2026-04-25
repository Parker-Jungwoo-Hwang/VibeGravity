#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 || "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  echo "Usage: collect.sh <run-id>" >&2
  exit 2
fi

run_id="$1"
repo_root="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
run_dir="$repo_root/.agents/hermes-orchestration/runs/$run_id"

if [[ ! -d "$run_dir" ]]; then
  echo "Run directory not found: $run_dir" >&2
  exit 2
fi

for meta in "$run_dir"/*.meta; do
  [[ -e "$meta" ]] || continue
  profile="$(basename "$meta" .meta)"
  out_file="$run_dir/$profile.out.md"
  err_file="$run_dir/$profile.err.log"

  echo "===== $profile meta ====="
  sed -n '1,120p' "$meta"
  echo
  echo "===== $profile output ====="
  if [[ -s "$out_file" ]]; then
    sed -n '1,220p' "$out_file"
  else
    echo "(no output)"
  fi
  if [[ -s "$err_file" ]]; then
    echo
    echo "===== $profile stderr ====="
    sed -n '1,120p' "$err_file"
  fi
  echo
done
