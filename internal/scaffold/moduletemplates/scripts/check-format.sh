#!/usr/bin/env bash
set -euo pipefail
formatter="$PWD/.build/tools/gofumpt"
[[ -x "$formatter" ]] || { echo 'Run task tools:format first' >&2; exit 1; }
files=$(mktemp)
trap 'rm -f "$files"' EXIT
find . -type d \( -name .git -o -name .build -o -name .devbox -o -name .codex-local -o -name vendor \) -prune -o -type f -name '*.go' -print0 > "$files"
failed=0
while IFS= read -r -d '' file; do
  [[ -L "$file" ]] && { echo "Refusing symlinked Go source: $file" >&2; exit 1; }
  for tool in gofmt "$formatter"; do
    output=$("$tool" -l "$file") || exit 1
    if [[ -n "$output" ]]; then printf '%s\n' "$output" >&2; failed=1; fi
  done
done < "$files"
exit "$failed"
