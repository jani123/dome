#!/usr/bin/env bash
set -euo pipefail
# Exercise the actual Hugo template with a published list and a local preview.
mkdir -p .build
output=$(mktemp -d "$PWD/.build/docs-versions.XXXXXX")
trap 'rm -rf "$output"' EXIT
HUGO_PARAMS_VERSION=next HUGO_PARAMS_DOCVERSIONS=next,v0.2.0,v0.1.0 \
  .build/tools/hugo --source docs --destination "$output/published" \
  --baseURL https://jani123.github.io/dome/next/ --panicOnWarning
page=$(cat "$output/published/index.html")
[[ "$page" == *'<select id="version-select"'* ]]
[[ "$page" == *'value="/dome/next/" selected'* ]]
[[ "$page" == *'value="https://jani123.github.io/dome/v0.2.0/"'* ]]
[[ "$page" == *'value="https://jani123.github.io/dome/v0.1.0/"'* ]]
HUGO_PARAMS_VERSION=next HUGO_PARAMS_DOCVERSIONS='' \
  .build/tools/hugo --source docs --destination "$output/local" \
  --baseURL http://localhost/ --panicOnWarning
page=$(cat "$output/local/index.html")
[[ "$page" == *'value="/" selected'* ]]
[[ "$page" == *'All published versions'* ]]
[[ "$page" != *'v0.2.0'* ]]
echo 'Documentation version picker checks passed'

# Bootstrap commands must match the selected guide source.
for version in next v0.2.0 v0.2.0-rc.1; do
  HUGO_PARAMS_VERSION="$version" HUGO_PARAMS_DOCVERSIONS='' \
    .build/tools/hugo --source docs --destination "$output/$version" \
    --baseURL "https://jani123.github.io/dome/$version/" --panicOnWarning
  quickstart=$(cat "$output/$version/quickstart/index.html")
  [[ "$quickstart" == *'id="bootstrap-go-tab" role="tab" aria-selected="true"'* ]]
  [[ "$quickstart" == *'id="bootstrap-docker-tab" role="tab" aria-selected="false"'* ]]
  [[ "$quickstart" == *'id="bootstrap-docker-panel" role="tabpanel" aria-labelledby="bootstrap-docker-tab" tabindex="0" hidden'* ]]
  expected=$version
  [[ "$version" != next ]] || expected=main
  page=$(cat "$output/$version/quickstart/index.html")
  [[ "$page" == *"github.com/jani123/dome/cmd/dome@$expected"* ]]
  if [[ "$version" == next ]]; then
    [[ "$page" == *'--project-module example.com/my-machine'* ]]
    [[ "$page" != *'github.com/jani123/dome/cmd/dome@v0.1.0'* ]]
  else
    [[ "$page" == *'--module example.com/my-machine'* ]]
  fi
done
echo 'Documentation bootstrap version checks passed'
