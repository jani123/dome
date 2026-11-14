#!/usr/bin/env bash
set -euo pipefail
scripts/openspec.sh config set profile custom
scripts/openspec.sh config set workflows '["explore","propose","apply","update","sync","archive","verify"]'
scripts/openspec.sh init --tools codex --profile custom --no-animation
