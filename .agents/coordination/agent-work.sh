#!/bin/sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
STATE_DIR="$ROOT_DIR/.agents/coordination"
PROGRESS_FILE="$STATE_DIR/WORK_PROGRESS.md"
CLAIMS_FILE="$STATE_DIR/claims.tsv"
LOG_FILE="$STATE_DIR/activity.log"
LOCK_DIR="$STATE_DIR/.lock"
STALE_CLAIM_SECONDS="${AGENT_WORK_STALE_SECONDS:-7200}"

usage() {
	cat <<'EOF'
Usage:
  .agents/coordination/agent-work.sh init
  .agents/coordination/agent-work.sh status [--json] [--no-render]
  .agents/coordination/agent-work.sh claim <agent-id> <task> <file> [<file> ...]
  .agents/coordination/agent-work.sh heartbeat <agent-id> <note>
  .agents/coordination/agent-work.sh release <agent-id> <file> [<file> ...]
  .agents/coordination/agent-work.sh done <agent-id> <note>

Claim exact files before editing them. Release each file as soon as that file is
finished, before moving to another write surface.
EOF
}

timestamp() {
	date -u "+%Y-%m-%dT%H:%M:%SZ"
}

sanitize() {
	printf "%s" "$1" | tr '\t\r\n' '   '
}

repo_path() {
	path=$1
	case "$path" in
		"$ROOT_DIR"/*)
			path=${path#"$ROOT_DIR"/}
			;;
	esac
	printf "%s" "$path"
}

date_to_epoch() {
	ts=$1
	epoch=$(date -u -j -f "%Y-%m-%dT%H:%M:%SZ" "$ts" "+%s" 2>/dev/null || true)
	if [ -z "$epoch" ]; then
		epoch=$(date -u -d "$ts" "+%s" 2>/dev/null || true)
	fi
	printf "%s" "$epoch"
}

validate_claim_path() {
	raw_path=$1
	path=$(repo_path "$raw_path")
	case "$path" in
		""|"-"|"--"|".")
			echo "Invalid claim path '$raw_path': claim exact repo file paths only." >&2
			return 1
			;;
		-*)
			echo "Invalid claim path '$raw_path': paths that look like flags are not allowed." >&2
			return 1
			;;
		/*)
			echo "Invalid claim path '$raw_path': path is outside this repo." >&2
			return 1
			;;
		*[[:space:]]*)
			echo "Invalid claim path '$raw_path': whitespace is not allowed in claim paths." >&2
			return 1
			;;
		*".."*|*"*"*|*"?"*|*"["*|*"]"*|*"**"*)
			echo "Invalid claim path '$raw_path': parent traversal and globs are not allowed." >&2
			return 1
			;;
		*/)
			echo "Invalid claim path '$raw_path': claim files, not directories." >&2
			return 1
			;;
	esac
	return 0
}

acquire_lock() {
	mkdir -p "$STATE_DIR"
	attempts=0
	while ! mkdir "$LOCK_DIR" 2>/dev/null; do
		attempts=$((attempts + 1))
		if [ "$attempts" -ge 30 ]; then
			echo "Timed out waiting for coordination lock: $LOCK_DIR" >&2
			exit 3
		fi
		sleep 1
	done
	trap 'rm -rf "$LOCK_DIR"' EXIT INT TERM
}

ensure_state() {
	[ -f "$CLAIMS_FILE" ] || : >"$CLAIMS_FILE"
	[ -f "$LOG_FILE" ] || : >"$LOG_FILE"
}

append_log() {
	action=$(sanitize "$1")
	agent=$(sanitize "$2")
	files=$(sanitize "$3")
	note=$(sanitize "$4")
	printf "%s\t%s\t%s\t%s\t%s\n" "$(timestamp)" "$action" "$agent" "$files" "$note" >>"$LOG_FILE"
}

join_files() {
	out=
	for raw_path in "$@"; do
		path=$(sanitize "$(repo_path "$raw_path")")
		if [ -z "$out" ]; then
			out=$path
		else
			out="$out, $path"
		fi
	done
	printf "%s" "$out"
}

status_json() {
	now=$(timestamp)
	printf "{\n"
	printf "  \"generated_at\": \"%s\",\n" "$now"
	printf "  \"progress_file\": \"%s\",\n" "$PROGRESS_FILE"
	printf "  \"claims_file\": \"%s\",\n" "$CLAIMS_FILE"
	printf "  \"log_file\": \"%s\",\n" "$LOG_FILE"
	printf "  \"active_claims\": [\n"
	if [ -s "$CLAIMS_FILE" ]; then
		awk -F '\t' '
			function json(s) {
				gsub(/\\/, "\\\\", s)
				gsub(/"/, "\\\"", s)
				gsub(/\r/, "\\r", s)
				gsub(/\n/, "\\n", s)
				gsub(/\t/, "\\t", s)
				return "\"" s "\""
			}
			NF >= 5 {
				note = ""
				if (NF >= 6) {
					note = $6
				}
				if (seen > 0) {
					printf ",\n"
				}
				printf "    {\"file\": %s, \"agent\": %s, \"task\": %s, \"claimed_at\": %s, \"last_update\": %s, \"note\": %s}", json($1), json($2), json($3), json($4), json($5), json(note)
				seen++
			}
		' "$CLAIMS_FILE"
	fi
	printf "\n  ],\n"
	printf "  \"recent_activity\": [\n"
	if [ -s "$LOG_FILE" ]; then
		tail -n 30 "$LOG_FILE" | awk -F '\t' '
			function json(s) {
				gsub(/\\/, "\\\\", s)
				gsub(/"/, "\\\"", s)
				gsub(/\r/, "\\r", s)
				gsub(/\n/, "\\n", s)
				gsub(/\t/, "\\t", s)
				return "\"" s "\""
			}
			NF >= 5 {
				if (seen > 0) {
					printf ",\n"
				}
				printf "    {\"time\": %s, \"action\": %s, \"agent\": %s, \"files\": %s, \"note\": %s}", json($1), json($2), json($3), json($4), json($5)
				seen++
			}
		'
	fi
	printf "\n  ]\n"
	printf "}\n"
}

status_cmd() {
	output_json=0
	render=1
	while [ "$#" -gt 0 ]; do
		case "$1" in
			--json)
				output_json=1
				;;
			--no-render)
				render=0
				;;
			*)
				usage >&2
				exit 2
				;;
		esac
		shift
	done
	if [ "$render" -eq 1 ]; then
		render_progress
	fi
	if [ "$output_json" -eq 1 ]; then
		status_json
	elif [ -f "$PROGRESS_FILE" ]; then
		cat "$PROGRESS_FILE"
	else
		echo "No rendered progress file. Run: .agents/coordination/agent-work.sh status" >&2
		exit 2
	fi
}

render_progress() {
	now=$(timestamp)
	now_epoch=$(date -u "+%s")
	tmp="$STATE_DIR/WORK_PROGRESS.md.tmp.$$"
	{
		echo "# Agent Work Progress"
		echo
		echo "Live shared progress board for concurrent VibeGravity agents."
		echo
		echo "Generated by \`.agents/coordination/agent-work.sh\`. Prefer the script over manual edits."
		echo
		echo "Last rendered: $now"
		echo
		echo "## Active Claims"
		echo
		if [ -s "$CLAIMS_FILE" ]; then
			echo "| File | Agent | Task | Claimed at | Last update | Note |"
			echo "|---|---|---|---|---|---|"
			awk -F '\t' '
				function esc(s) {
					gsub(/\|/, "\\|", s)
					return s
				}
				NF >= 5 {
					note = ""
					if (NF >= 6) {
						note = $6
					}
					printf "| `%s` | %s | %s | %s | %s | %s |\n", esc($1), esc($2), esc($3), esc($4), esc($5), esc(note)
				}
			' "$CLAIMS_FILE"
		else
			echo "No active claims."
		fi
		echo
		echo "## Stale Claim Warnings"
		echo
		warnings=0
		if [ -s "$CLAIMS_FILE" ]; then
			tab=$(printf '\t')
			while IFS="$tab" read -r file agent task claimed_at last_update note; do
				last_epoch=$(date_to_epoch "$last_update")
				if [ -n "$last_epoch" ]; then
					age=$((now_epoch - last_epoch))
					if [ "$age" -ge "$STALE_CLAIM_SECONDS" ]; then
						warnings=$((warnings + 1))
						age_minutes=$((age / 60))
						printf -- '- `%s` owned by %s has no heartbeat for %s minutes. Last note: %s\n' "$file" "$agent" "$age_minutes" "$note"
					fi
				fi
			done <"$CLAIMS_FILE"
		fi
		if [ "$warnings" -eq 0 ]; then
			echo "No stale claims."
		fi
		echo
		echo "## Recent Activity"
		echo
		if [ -s "$LOG_FILE" ]; then
			echo "| Time | Action | Agent | Files | Note |"
			echo "|---|---|---|---|---|"
			tail -n 30 "$LOG_FILE" | awk -F '\t' '
				function esc(s) {
					gsub(/\|/, "\\|", s)
					return s
				}
				NF >= 5 {
					printf "| %s | %s | %s | %s | %s |\n", esc($1), esc($2), esc($3), esc($4), esc($5)
				}
			'
		else
			echo "No activity yet."
		fi
		echo
		echo "## Protocol"
		echo
		echo "- Read this file before starting and before widening scope."
		echo "- Claim exact files before editing."
		echo "- Do not edit files claimed by another active agent."
		echo "- Heartbeat during long work."
		echo "- Release files immediately when finished."
	} >"$tmp"
	mv "$tmp" "$PROGRESS_FILE"
}

claim_cmd() {
	if [ "$#" -lt 3 ]; then
		usage >&2
		exit 2
	fi
	agent=$(sanitize "$1")
	task=$(sanitize "$2")
	shift 2
	now=$(timestamp)
	conflicts=
	for raw_path in "$@"; do
		validate_claim_path "$raw_path"
	done
	for raw_path in "$@"; do
		path=$(sanitize "$(repo_path "$raw_path")")
		owner=$(awk -F '\t' -v file="$path" -v agent="$agent" '$1 == file && $2 != agent { print $2; exit }' "$CLAIMS_FILE")
		if [ -n "$owner" ]; then
			if [ -z "$conflicts" ]; then
				conflicts="$path is already claimed by $owner"
			else
				conflicts="$conflicts
$path is already claimed by $owner"
			fi
		fi
	done
	if [ -n "$conflicts" ]; then
		echo "Claim rejected:" >&2
		printf "%s\n" "$conflicts" >&2
		exit 2
	fi

	tmp="$STATE_DIR/claims.tsv.tmp.$$"
	cp "$CLAIMS_FILE" "$tmp"
	for raw_path in "$@"; do
		path=$(sanitize "$(repo_path "$raw_path")")
		next="$tmp.next"
		awk -F '\t' -v file="$path" '$1 != file { print }' "$tmp" >"$next"
		mv "$next" "$tmp"
		printf "%s\t%s\t%s\t%s\t%s\t%s\n" "$path" "$agent" "$task" "$now" "$now" "$task" >>"$tmp"
	done
	mv "$tmp" "$CLAIMS_FILE"
	append_log "claim" "$agent" "$(join_files "$@")" "$task"
}

heartbeat_cmd() {
	if [ "$#" -ne 2 ]; then
		usage >&2
		exit 2
	fi
	agent=$(sanitize "$1")
	note=$(sanitize "$2")
	now=$(timestamp)
	tmp="$STATE_DIR/claims.tsv.tmp.$$"
	awk -F '\t' -v OFS='\t' -v agent="$agent" -v now="$now" -v note="$note" '
		$2 == agent {
			$5 = now
			$6 = note
		}
		{ print }
	' "$CLAIMS_FILE" >"$tmp"
	mv "$tmp" "$CLAIMS_FILE"
	append_log "heartbeat" "$agent" "$(awk -F '\t' -v agent="$agent" '$2 == agent { if (out == "") out = $1; else out = out ", " $1 } END { print out }' "$CLAIMS_FILE")" "$note"
}

release_cmd() {
	if [ "$#" -lt 2 ]; then
		usage >&2
		exit 2
	fi
	agent=$(sanitize "$1")
	shift
	conflicts=
	for raw_path in "$@"; do
		path=$(sanitize "$(repo_path "$raw_path")")
		owner=$(awk -F '\t' -v file="$path" '$1 == file { print $2; exit }' "$CLAIMS_FILE")
		if [ -n "$owner" ] && [ "$owner" != "$agent" ]; then
			if [ -z "$conflicts" ]; then
				conflicts="$path is claimed by $owner, not $agent"
			else
				conflicts="$conflicts
$path is claimed by $owner, not $agent"
			fi
		fi
	done
	if [ -n "$conflicts" ]; then
		echo "Release rejected:" >&2
		printf "%s\n" "$conflicts" >&2
		exit 2
	fi

	tmp="$STATE_DIR/claims.tsv.tmp.$$"
	cp "$CLAIMS_FILE" "$tmp"
	for raw_path in "$@"; do
		path=$(sanitize "$(repo_path "$raw_path")")
		next="$tmp.next"
		awk -F '\t' -v file="$path" -v agent="$agent" '!(($1 == file) && ($2 == agent)) { print }' "$tmp" >"$next"
		mv "$next" "$tmp"
	done
	mv "$tmp" "$CLAIMS_FILE"
	append_log "release" "$agent" "$(join_files "$@")" "released claimed files"
}

done_cmd() {
	if [ "$#" -ne 2 ]; then
		usage >&2
		exit 2
	fi
	agent=$(sanitize "$1")
	note=$(sanitize "$2")
	files=$(awk -F '\t' -v agent="$agent" '$2 == agent { if (out == "") out = $1; else out = out ", " $1 } END { print out }' "$CLAIMS_FILE")
	tmp="$STATE_DIR/claims.tsv.tmp.$$"
	awk -F '\t' -v agent="$agent" '$2 != agent { print }' "$CLAIMS_FILE" >"$tmp"
	mv "$tmp" "$CLAIMS_FILE"
	append_log "done" "$agent" "$files" "$note"
}

if [ "$#" -lt 1 ]; then
	usage >&2
	exit 2
fi

cmd=$1
shift

case "$cmd" in
	init)
		acquire_lock
		ensure_state
		render_progress
		echo "$PROGRESS_FILE"
		;;
	status)
		acquire_lock
		ensure_state
		status_cmd "$@"
		;;
	claim)
		acquire_lock
		ensure_state
		claim_cmd "$@"
		render_progress
		echo "Claim recorded. See $PROGRESS_FILE"
		;;
	heartbeat)
		acquire_lock
		ensure_state
		heartbeat_cmd "$@"
		render_progress
		echo "Heartbeat recorded. See $PROGRESS_FILE"
		;;
	release)
		acquire_lock
		ensure_state
		release_cmd "$@"
		render_progress
		echo "Release recorded. See $PROGRESS_FILE"
		;;
	done)
		acquire_lock
		ensure_state
		done_cmd "$@"
		render_progress
		echo "Done recorded. See $PROGRESS_FILE"
		;;
	*)
		usage >&2
		exit 2
		;;
esac
