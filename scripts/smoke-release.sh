#!/usr/bin/env bash
set -euo pipefail
version=${1:?release tag required}
arch=${2:?amd64 or arm64 required}
[[ "$version" =~ ^v0\.[0-9]+\.[0-9]+(-rc\.[0-9]+)?$ ]] || exit 1
case "$arch" in amd64) test "$(uname -m)" = x86_64;; arm64) test "$(uname -m)" = aarch64;; *) exit 1;; esac
stage=$(mktemp -d)
trap 'rm -rf "$stage"' EXIT
(cd .build/release; sha256sum --check SHA256SUMS)
tar -xzf ".build/release/dome_${version}_linux_${arch}.tar.gz" -C "$stage"
test "$("$stage/dome" version)" = "$version"
"$stage/dome" init "$stage/consumer" --module example.com/release-smoke
cd "$stage/consumer"
# Resolve the public tagged module: no workspace/replacement/local source shortcuts.
test "$(go list -m -f '{{if .Replace}}replacement{{else}}{{.Version}}{{end}}' github.com/jani123/dome)" = "$version"
CGO_ENABLED=0 go build -o agent ./cmd/agent
go test ./...
./agent plan --json > "$stage/plan.json"
test ! -e sandbox/hello.txt
./agent apply --json > "$stage/apply.json"
./agent apply --json > "$stage/noop.json"
grep -q '"no-op"' "$stage/noop.json"
test -f sandbox/hello.txt
printf 'managed: false\n' > config.yaml
./agent apply --json > "$stage/prune.json"
test ! -e sandbox/hello.txt
