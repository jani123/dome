#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p .build
# The lock lives inside this container, not the shared checkout. Kernel-owned
# locks disappear when the server exits, including after a container restart.
nohup setsid flock --nonblock /tmp/dome-docs-preview.lock \
  devbox run -- task docs:serve >> .build/docs-preview.log 2>&1 < /dev/null &
for ((attempt = 0; attempt < 60; attempt++)); do
  if curl --noproxy '*' --fail --silent --output /dev/null http://127.0.0.1:1313/; then
    echo "Docs preview ready; run task devcontainer:urls on the host."
    exit 0
  fi
  sleep 1
done
echo "Docs preview did not become ready; see .build/docs-preview.log" >&2
tail -n 30 .build/docs-preview.log >&2
exit 1
