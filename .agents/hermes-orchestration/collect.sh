#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE' >&2
Usage:
  collect.sh <run-id>
  collect.sh --json <run-id>
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

json_bool() {
  if [[ -s "$1" ]]; then
    printf "true"
  else
    printf "false"
  fi
}

json_collect() {
  printf "{\n"
  printf "  \"run_id\": %s,\n" "$(json_string "$run_id")"
  printf "  \"run_dir\": %s,\n" "$(json_string "$run_dir")"
  printf "  \"profiles\": [\n"
  local first=1
  local meta profile out_file err_file result_file exit_code
  for meta in "$run_dir"/*.meta; do
    [[ -e "$meta" ]] || continue
    profile="$(basename "$meta" .meta)"
    out_file="$run_dir/$profile.out.md"
    err_file="$run_dir/$profile.err.log"
    result_file="$run_dir/$profile.result.json"
    exit_code="$(awk -F= '$1 == "exit_code" { print $2 }' "$meta" | tail -n 1)"
    [[ -n "$exit_code" ]] || exit_code=null
    if [[ "$first" -eq 0 ]]; then
      printf ",\n"
    fi
    first=0
    printf "    {\n"
    printf "      \"profile\": %s,\n" "$(json_string "$profile")"
    printf "      \"exit_code\": %s,\n" "$exit_code"
    printf "      \"meta_file\": %s,\n" "$(json_string "$meta")"
    printf "      \"out_file\": %s,\n" "$(json_string "$out_file")"
    printf "      \"err_file\": %s,\n" "$(json_string "$err_file")"
    printf "      \"result_file\": %s,\n" "$(json_string "$result_file")"
    printf "      \"has_output\": %s,\n" "$(json_bool "$out_file")"
    printf "      \"has_stderr\": %s,\n" "$(json_bool "$err_file")"
    printf "      \"has_result\": %s\n" "$(json_bool "$result_file")"
    printf "    }"
  done
  printf "\n  ]\n"
  printf "}\n"
}

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

output_json=0
if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi
if [[ "${1:-}" == "--json" ]]; then
  output_json=1
  shift
fi
if [[ $# -ne 1 ]]; then
  usage
  exit 2
fi

run_id="$1"
repo_root="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
run_dir="$repo_root/.agents/hermes-orchestration/runs/$run_id"

if [[ ! -d "$run_dir" ]]; then
  echo "Run directory not found: $run_dir" >&2
  exit 2
fi

write_synthesis_packet

if [[ "$output_json" -eq 1 ]]; then
  json_collect
  exit 0
fi

echo "synthesis=$run_dir/synthesis.md"
echo

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
