#!/usr/bin/env bash
set -euo pipefail
# Install the pinned official Hugo release after checking its published checksum.
version=$(cat .hugo-version)
case $(uname -m) in x86_64) arch=amd64;; aarch64|arm64) arch=arm64;; *) echo 'Unsupported Hugo build architecture' >&2; exit 1;; esac
mkdir -p .build/tools
stage=$(mktemp -d .build/tools/hugo.XXXXXX)
trap 'rm -rf "$stage"' EXIT
archive="hugo_${version}_linux-${arch}.tar.gz"
base="https://github.com/gohugoio/hugo/releases/download/v${version}"
curl --fail --silent --show-error --location "$base/$archive" --output "$stage/$archive"
curl --fail --silent --show-error --location "$base/hugo_${version}_checksums.txt" --output "$stage/checksums.txt"
(cd "$stage"; awk -v archive="$archive" '$2 == archive { print }' checksums.txt | sha256sum --check --status)
tar -xzf "$stage/$archive" -C "$stage" hugo
install -m 0755 "$stage/hugo" .build/tools/hugo
.build/tools/hugo version
