#!/usr/bin/env bash
set -euo pipefail
version=$(cat .hugo-version)
if [[ ! -x .build/tools/hugo ]] || [[ "$(.build/tools/hugo version)" != "hugo v${version}"* ]]; then
  scripts/install-hugo.sh
fi
exec .build/tools/hugo server --source docs --bind 0.0.0.0 --port 1313 \
  --baseURL http://localhost/ --appendPort=false --disableLiveReload --disableFastRender \
  --destination "$PWD/.build/docs-preview" --panicOnWarning
