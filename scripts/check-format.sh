#!/usr/bin/env bash
set -euo pipefail
# Git's NUL-delimited listing includes new files without inspecting ignored caches.
files=$(mktemp)
trap 'rm -f "$files"' EXIT
git ls-files -z --cached --others --exclude-standard -- '*.go' > "$files"
failed=0
while IFS= read -r -d '' file; do
  [[ -L "$file" ]] && { echo "Refusing symlinked Go source: $file" >&2; exit 1; }
  [[ -f "$file" ]] || continue
  for formatter in gofmt gofumpt; do
    output=$("$formatter" -l "./$file") || exit 1
    if [[ -n "$output" ]]; then
      printf '%s: %s\n' "$formatter" "$file" >&2
      failed=1
    fi
  done
done < "$files"
exit "$failed"
