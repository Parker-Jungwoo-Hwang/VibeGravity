#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage:
  run-agent.sh <profile> <prompt-file> [run-id]

Profiles:
  default
  vuitton
  bottega

Environment overrides:
  HERMES_MODEL       default: gpt-5.5
  HERMES_PROVIDER    default: openai-codex
  HERMES_MAX_TURNS   default: 90
USAGE
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

if [[ $# -lt 2 || $# -gt 3 ]]; then
  usage >&2
  exit 2
fi

profile="$1"
prompt_file="$2"
run_id="${3:-$(date +%Y%m%d_%H%M%S)}"

case "$profile" in
  default)
    hermes_home="/Users/parker/.hermes"
    ;;
  vuitton)
    hermes_home="/Users/parker/.hermes/profiles/vuitton"
    ;;
  bottega)
    hermes_home="/Users/parker/.hermes/profiles/bottega"
    ;;
  *)
    echo "Unknown profile: $profile" >&2
    exit 2
    ;;
esac

if [[ ! -f "$prompt_file" ]]; then
  echo "Prompt file not found: $prompt_file" >&2
  exit 2
fi

repo_root="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
run_dir="$repo_root/.agents/hermes-orchestration/runs/$run_id"
mkdir -p "$run_dir"

out_file="$run_dir/$profile.out.md"
err_file="$run_dir/$profile.err.log"
meta_file="$run_dir/$profile.meta"

model="${HERMES_MODEL:-gpt-5.5}"
provider="${HERMES_PROVIDER:-openai-codex}"
max_turns="${HERMES_MAX_TURNS:-90}"

{
  echo "profile=$profile"
  echo "hermes_home=$hermes_home"
  echo "prompt_file=$prompt_file"
  echo "run_id=$run_id"
  echo "model=$model"
  echo "provider=$provider"
  echo "started_at=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
} > "$meta_file"

set +e
env HERMES_HOME="$hermes_home" hermes chat \
  -Q \
  --source tool \
  --provider "$provider" \
  -m "$model" \
  --max-turns "$max_turns" \
  -q "$(cat "$prompt_file")" \
  > "$out_file" \
  2> "$err_file"
status=$?
set -e

{
  echo "finished_at=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  echo "exit_code=$status"
  echo "out_file=$out_file"
  echo "err_file=$err_file"
} >> "$meta_file"

exit "$status"
