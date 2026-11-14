#!/usr/bin/env bash
set -euo pipefail
version=${1:?exact release tag required}
[[ "$version" =~ ^v[01]\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z.-]+)?$ ]] || { echo 'Expected canonical v0/v1 release tag' >&2; exit 1; }
# Go module validation rejects malformed prereleases and major-path mismatches.
go run scripts/check-release.go "$version"
[[ "$(go list -m -f '{{.Path}}')" == '__MODULE__' ]] || exit 1
[[ "$(git rev-parse HEAD)" == "$(git rev-parse "refs/tags/$version^{commit}")" ]] || exit 1
[[ -z "$(git status --porcelain --untracked-files=no)" ]] || exit 1
mkdir -p .build/release
stage=$(mktemp -d .build/release/source.XXXXXX)
trap 'rm -rf "$stage"' EXIT
git archive HEAD | tar -x -C "$stage"
go mod vendor -o "$stage/vendor"
tar -czf ".build/release/module_${version}_source.tar.gz" -C "$stage" .
(cd .build/release; sha256sum "module_${version}_source.tar.gz" > SHA256SUMS)
flags=()
if [[ "$version" == *-* ]]; then flags+=(--prerelease); fi
gh release create "$version" ".build/release/module_${version}_source.tar.gz" .build/release/SHA256SUMS \
  --verify-tag --draft "${flags[@]}" --title "__NAME__ $version" --notes-file CHANGELOG.md
