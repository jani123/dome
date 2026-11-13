#!/usr/bin/env bash
set -euo pipefail
if [[ ! -x .build/tools/openspec ]]; then
  echo 'Run make tools-openspec first' >&2; exit 1
fi
if [[ $(.build/tools/openspec --version) != "$(cat .openspec-version)" ]]; then
  echo 'OpenSpec version differs from .openspec-version; run make tools-openspec' >&2; exit 1
fi
# Avoid changing personal profile/configuration while refreshing repo-owned skills.
export XDG_CONFIG_HOME="$PWD/.build/openspec-config"
export OPENSPEC_TELEMETRY=0 DO_NOT_TRACK=1 OPENSPEC_NO_UPDATE_CHECK=1
exec .build/tools/openspec "$@"
