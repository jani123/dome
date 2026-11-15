#!/usr/bin/env bash
set -euo pipefail
version=${1:?exact tag required}
[[ "$version" =~ ^v0\.[0-9]+\.[0-9]+(-rc\.[0-9]+)?$ ]] || { echo 'Expected v0.x.y or v0.x.y-rc.n' >&2; exit 1; }
[[ "$(git rev-parse HEAD)" == "$(git rev-parse "refs/tags/$version^{commit}")" ]] || { echo 'Tag does not match checkout' >&2; exit 1; }
test -z "$(git status --porcelain --untracked-files=no)"
mkdir -p .build/release
for arch in amd64 arm64; do
  dir=$(mktemp -d .build/release/package.XXXXXX)
  CGO_ENABLED=0 GOOS=linux GOARCH="$arch" go build -trimpath \
    -ldflags "-s -w -X main.version=$version" -o "$dir/dome" ./cmd/dome
  cp LICENSE CHANGELOG.md "$dir/"
  DOME_NOTICE_GOARCH="$arch" go run scripts/dependency-notices.go ./cmd/dome "$dir/THIRD_PARTY_NOTICES.txt"
  printf 'Dome %s\nCommit %s\nLinux %s\n' "$version" "$(git rev-parse HEAD)" "$arch" > "$dir/BUILD.txt"
  tar -czf ".build/release/dome_${version}_linux_${arch}.tar.gz" -C "$dir" dome LICENSE CHANGELOG.md THIRD_PARTY_NOTICES.txt BUILD.txt
  rm -rf "$dir"
done
# Ship corresponding source for the initializer and its linked Go dependencies.
source_dir=$(mktemp -d .build/release/source.XXXXXX)
git archive HEAD | tar -x -C "$source_dir"
go mod vendor -o "$source_dir/vendor"
cat > "$source_dir/BUILD-SOURCE.md" <<BUILD
# Build this release from corresponding source

Requires Go 1.26.11+ on Linux. All Go module sources used by this checkout are
included under vendor with their upstream license and attribution material.

    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -trimpath -ldflags "-s -w -X main.version=$version" -o dome ./cmd/dome

Use GOARCH=arm64 for that target. The release commit is $(git rev-parse HEAD).
See README.md and scripts/release.sh for packaging and verification commands.
BUILD
tar -czf ".build/release/dome_${version}_source.tar.gz" -C "$source_dir" .
rm -rf "$source_dir"
(cd .build/release; sha256sum "dome_${version}_"*.tar.gz > SHA256SUMS)
