#!/usr/bin/env bash
# UserPromptSubmit hook: detect active plan and show progress for session recovery.
# If no plan is active, outputs nothing.

set -euo pipefail

# Find git root by walking up from $PWD
find_git_root() {
  local dir="$PWD"
  while [ "$dir" != "/" ]; do
    if [ -d "$dir/.git" ]; then
      echo "$dir"
      return 0
    fi
    dir="$(dirname "$dir")"
  done
  return 1
}

git_root="$(find_git_root)" || exit 0

current_file="$git_root/plans/.current"
[ -f "$current_file" ] || exit 0

plan_name="$(cat "$current_file")"
[ -n "$plan_name" ] || exit 0

# Convert plan name to filename (lowercase, spaces/slashes to dashes)
plan_file="$git_root/plans/$(echo "$plan_name" | tr '[:upper:]' '[:lower:]' | tr ' /' '--').md"
[ -f "$plan_file" ] || exit 0

# If trail binary is available, use it for richer output
if command -v trail &>/dev/null; then
  output="$(trail status 2>/dev/null)" || output=""
  if [ -n "$output" ]; then
    echo "[trail] $output"
    echo "[trail] Read $plan_file to resume work."
    exit 0
  fi
fi

# Fallback: parse plan file directly
# Count top-level checkboxes (no leading whitespace) under ## Tasks
in_tasks=false
total=0
done=0

while IFS= read -r line; do
  if [[ "$line" =~ ^##\  ]]; then
    if [[ "$line" =~ ^##\ Tasks ]]; then
      in_tasks=true
    else
      in_tasks=false
    fi
    continue
  fi

  if $in_tasks; then
    if [[ "$line" =~ ^-\ \[x\]\  ]]; then
      total=$((total + 1))
      done=$((done + 1))
    elif [[ "$line" =~ ^-\ \[\ \]\  ]]; then
      total=$((total + 1))
    fi
  fi
done < "$plan_file"

if [ "$total" -gt 0 ]; then
  echo "[trail] Active plan: $plan_name ($done/$total tasks done)"
else
  echo "[trail] Active plan: $plan_name"
fi
echo "[trail] Read $plan_file to resume work."
