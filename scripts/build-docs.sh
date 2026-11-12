#!/usr/bin/env bash
set -euo pipefail
version=${1:-next}
if [[ "$version" != next && ! "$version" =~ ^v0\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$ ]]; then
  echo 'Expected next or an exact v0 release version' >&2; exit 1
fi
# Build a versioned snapshot. The deployment workflow assembles retained snapshots.
output="$PWD/.build/docs/$version"
HUGO_PARAMS_VERSION="$version" .build/tools/hugo --source docs --destination "$output" \
  --baseURL "https://jani123.github.io/dome/$version/" --panicOnWarning
go run scripts/check-docs.go "$output" "/dome/$version/"
