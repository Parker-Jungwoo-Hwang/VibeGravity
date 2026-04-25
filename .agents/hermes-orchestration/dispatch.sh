#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage:
  dispatch.sh <run-id> <manifest.tsv>

Manifest format:
  <profile><TAB><prompt-file>

Blank lines and lines starting with # are ignored.
USAGE
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

if [[ $# -ne 2 ]]; then
  usage >&2
  exit 2
fi

run_id="$1"
manifest="$2"

if [[ ! -f "$manifest" ]]; then
  echo "Manifest not found: $manifest" >&2
  exit 2
fi

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
run_dir="$repo_root/.agents/hermes-orchestration/runs/$run_id"
mkdir -p "$run_dir"

pids=()
profiles=()

while IFS=$'\t' read -r profile prompt_file rest; do
  if [[ -z "${profile:-}" || "${profile:0:1}" == "#" ]]; then
    continue
  fi
  if [[ -n "${rest:-}" ]]; then
    echo "Invalid manifest line for profile $profile: too many fields" >&2
    exit 2
  fi

  "$script_dir/run-agent.sh" "$profile" "$prompt_file" "$run_id" &
  pids+=("$!")
  profiles+=("$profile")
done < "$manifest"

if [[ ${#pids[@]} -eq 0 ]]; then
  echo "No tasks found in manifest: $manifest" >&2
  exit 2
fi

status=0
for i in "${!pids[@]}"; do
  if wait "${pids[$i]}"; then
    echo "${profiles[$i]}: ok"
  else
    code=$?
    echo "${profiles[$i]}: failed ($code)" >&2
    status=1
  fi
done

echo "run_dir=$run_dir"
exit "$status"
