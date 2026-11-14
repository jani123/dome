#!/usr/bin/env bash
set -euo pipefail

# Fix named-volume roots without changing ownership of the shared checkout.
sudo chown "$(id -u):$(id -g)" .devbox "$HOME/.cache" "$HOME/go"
sudo chown -R "$(id -u):$(id -g)" "$NPM_CONFIG_PREFIX"
devbox run -- task dev:setup
