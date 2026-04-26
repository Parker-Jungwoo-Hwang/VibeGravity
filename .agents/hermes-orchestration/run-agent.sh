#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage:
  run-agent.sh <profile> <prompt-file> [run-id]

Profiles:
  Configured in .agents/hermes-orchestration/profiles.tsv

Environment overrides:
  HERMES_PROFILE_MANIFEST  default: .agents/hermes-orchestration/profiles.tsv
  HERMES_PROFILE_ROOT      default: $HOME/.hermes
  HERMES_MODEL       default: gpt-5.5
  HERMES_PROVIDER    default: openai-codex
  HERMES_MAX_TURNS   default: 90
USAGE
}

json_string() {
  awk -v value="$1" 'BEGIN {
    gsub(/\\/, "\\\\", value)
    gsub(/"/, "\\\"", value)
    gsub(/\r/, "\\r", value)
    gsub(/\n/, "\\n", value)
    gsub(/\t/, "\\t", value)
    printf "\"%s\"", value
  }'
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
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
profile_manifest="${HERMES_PROFILE_MANIFEST:-$script_dir/profiles.tsv}"
profile_root="${HERMES_PROFILE_ROOT:-$HOME/.hermes}"

resolve_hermes_home() {
  local requested_profile="$1"
  local manifest="$2"
  local configured_profile
  local configured_home
  local extra

  if [[ -f "$manifest" ]]; then
    while IFS=$'\t' read -r configured_profile configured_home extra; do
      if [[ -z "${configured_profile:-}" || "${configured_profile:0:1}" == "#" ]]; then
        continue
      fi
      if [[ -n "${extra:-}" ]]; then
        echo "Invalid profile manifest line for $configured_profile: too many fields" >&2
        return 2
      fi
      if [[ "$configured_profile" == "$requested_profile" ]]; then
        configured_home="${configured_home/#\~/$HOME}"
        configured_home="${configured_home/#\$HOME/$HOME}"
        configured_home="${configured_home/#\$\{HOME\}/$HOME}"
        printf "%s\n" "$configured_home"
        return 0
      fi
    done < "$manifest"
  fi

  case "$requested_profile" in
    default)
      printf "%s\n" "$profile_root"
      ;;
    *)
      printf "%s/profiles/%s\n" "$profile_root" "$requested_profile"
      ;;
  esac
}

hermes_home="$(resolve_hermes_home "$profile" "$profile_manifest")"

if [[ ! -f "$prompt_file" ]]; then
  echo "Prompt file not found: $prompt_file" >&2
  exit 2
fi

run_dir="$repo_root/.agents/hermes-orchestration/runs/$run_id"
mkdir -p "$run_dir"

out_file="$run_dir/$profile.out.md"
err_file="$run_dir/$profile.err.log"
meta_file="$run_dir/$profile.meta"
result_file="$run_dir/$profile.result.json"

model="${HERMES_MODEL:-gpt-5.5}"
provider="${HERMES_PROVIDER:-openai-codex}"
max_turns="${HERMES_MAX_TURNS:-90}"
started_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

{
  echo "profile=$profile"
  echo "hermes_home=$hermes_home"
  echo "profile_manifest=$profile_manifest"
  echo "prompt_file=$prompt_file"
  echo "run_id=$run_id"
  echo "model=$model"
  echo "provider=$provider"
  echo "started_at=$started_at"
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

finished_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

{
  echo "finished_at=$finished_at"
  echo "exit_code=$status"
  echo "out_file=$out_file"
  echo "err_file=$err_file"
  echo "result_file=$result_file"
} >> "$meta_file"

{
  printf "{\n"
  printf "  \"profile\": %s,\n" "$(json_string "$profile")"
  printf "  \"run_id\": %s,\n" "$(json_string "$run_id")"
  printf "  \"exit_code\": %s,\n" "$status"
  printf "  \"started_at\": %s,\n" "$(json_string "$started_at")"
  printf "  \"finished_at\": %s,\n" "$(json_string "$finished_at")"
  printf "  \"hermes_home\": %s,\n" "$(json_string "$hermes_home")"
  printf "  \"profile_manifest\": %s,\n" "$(json_string "$profile_manifest")"
  printf "  \"prompt_file\": %s,\n" "$(json_string "$prompt_file")"
  printf "  \"model\": %s,\n" "$(json_string "$model")"
  printf "  \"provider\": %s,\n" "$(json_string "$provider")"
  printf "  \"max_turns\": %s,\n" "$(json_string "$max_turns")"
  printf "  \"out_file\": %s,\n" "$(json_string "$out_file")"
  printf "  \"err_file\": %s,\n" "$(json_string "$err_file")"
  printf "  \"meta_file\": %s,\n" "$(json_string "$meta_file")"
  printf "  \"result_file\": %s\n" "$(json_string "$result_file")"
  printf "}\n"
} > "$result_file"

exit "$status"
