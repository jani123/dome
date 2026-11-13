#!/usr/bin/env bash
set -euo pipefail
version=$(cat .openspec-version)
[[ "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || { echo 'Invalid OpenSpec version pin' >&2; exit 1; }
[[ $(uname -s) == Linux ]] || { echo 'OpenSpec installer supports Linux' >&2; exit 1; }
case $(uname -m) in x86_64) arch=x64;; aarch64|arm64) arch=arm64;; *) echo 'Unsupported OpenSpec architecture' >&2; exit 1;; esac
mkdir -p .build/tools
stage=$(mktemp -d .build/tools/openspec.XXXXXX)
trap 'rm -rf "$stage"' EXIT
archive="openspec-linux-${arch}.tar.gz"
base="https://github.com/Fission-AI/OpenSpec/releases/download/v${version}"
curl --fail --silent --show-error --location "$base/$archive" --output "$stage/$archive"
curl --fail --silent --show-error --location "$base/SHA256SUMS" --output "$stage/SHA256SUMS"
(cd "$stage"; awk -v archive="$archive" '$2 == archive { print }' SHA256SUMS | sha256sum --check --status)
tar -xzf "$stage/$archive" -C "$stage" openspec
test "$("$stage/openspec" --version)" = "$version"
install -m 0755 "$stage/openspec" .build/tools/openspec
.build/tools/openspec --version
