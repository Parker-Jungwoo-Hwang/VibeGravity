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

write_synthesis_packet() {
  local synthesis_file="$run_dir/synthesis.md"
  local generated_at
  generated_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

  {
    echo "---"
    echo "agent_id: hermes-orchestration"
    echo "role: hermes-orchestration"
    echo "phase_id: ${HERMES_PHASE_ID:-phase-05-quality-ops}"
    echo "lane_id: $run_id"
    echo "lane_type: integration_synthesis"
    echo "claimed_files: []"
    echo "reviewed_files: []"
    echo "changed_files: []"
    echo "gates_run: []"
    echo "gates_skipped: []"
    echo "skip_reasons: {}"
    echo "next_owner: leader"
    echo "---"
    echo
    echo "# Hermes Run Synthesis: $run_id"
    echo
    echo "Generated at: $generated_at"
    echo
    echo "This packet records the multi-Hermes run output. It is not final synthesis approval; leader approval is still required."
    echo
    echo "## Profiles"
    echo
    echo "| Profile | Prompt | Exit code | Output | Error log | Result JSON |"
    echo "|---|---|---|---|---|---|"
    for meta in "$run_dir"/*.meta; do
      [[ -e "$meta" ]] || continue
      local profile prompt_file exit_code out_file err_file result_file
      profile="$(basename "$meta" .meta)"
      prompt_file="$(grep '^prompt_file=' "$meta" | cut -d= -f2- || true)"
      exit_code="$(grep '^exit_code=' "$meta" | tail -n 1 | cut -d= -f2- || true)"
      out_file="$(grep '^out_file=' "$meta" | tail -n 1 | cut -d= -f2- || true)"
      err_file="$(grep '^err_file=' "$meta" | tail -n 1 | cut -d= -f2- || true)"
      result_file="$run_dir/$profile.result.json"
      echo "| $profile | $prompt_file | ${exit_code:-pending} | $out_file | $err_file | $result_file |"
    done
    echo
    echo "## Required Leader Review"
    echo
    echo "- Check each profile output for reported changed files."
    echo "- Check each profile output for gates run and skipped."
    echo "- Resolve conflicting recommendations or file ownership."
    echo "- Approve or reject final synthesis explicitly."
  } > "$synthesis_file"
}

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

write_synthesis_packet
echo "run_dir=$run_dir"
echo "synthesis=$run_dir/synthesis.md"
exit "$status"
