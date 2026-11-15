#!/usr/bin/env bash
set -euo pipefail
# Input is a list of published release tags supplied by the trusted workflow.
# Rebuild each snapshot from its own tagged source and pinned Hugo version.
tags_file=${1:?path to published release tag list required}
repo_root=$PWD
site="$repo_root/.build/site"
rm -rf "$site"
mkdir -p "$site"
# Supply the picker from the same published tags used to assemble snapshots.
# Validate before passing values into Hugo templates.
doc_versions=next
while IFS= read -r tag; do
  [[ -z "$tag" ]] && continue
  [[ "$tag" =~ ^v0\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$ ]] || { echo 'Invalid release tag' >&2; exit 1; }
  doc_versions+=",$tag"
done < <(sort -Vr "$tags_file")
export HUGO_PARAMS_DOCVERSIONS="$doc_versions"
scripts/build-docs.sh next
cp -a .build/docs/next "$site/next"
stable=''
while IFS= read -r tag; do
  [[ -z "$tag" ]] && continue
  [[ "$tag" =~ ^v0\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$ ]] || { echo 'Invalid release tag' >&2; exit 1; }
  git rev-parse --verify "refs/tags/$tag" >/dev/null
  source_dir="$repo_root/.build/docs-src/$tag"
  mkdir -p "$(dirname "$source_dir")"
  git worktree add --detach "$source_dir" "refs/tags/$tag"
  # Failure leaves the worktree available for diagnosis; no snapshot is deployed.
  (cd "$source_dir"; scripts/install-hugo.sh; scripts/build-docs.sh "$tag")
  cp -a "$source_dir/.build/docs/$tag" "$site/$tag"
  git worktree remove --force "$source_dir"
  if [[ "$tag" != *-* ]]; then stable="$tag"; fi
done < <(sort -V "$tags_file")
if [[ -n "$stable" ]]; then
  # Keep the version's base URL in the alias so links resolve to immutable guides.
  cp -a "$site/$stable" "$site/latest"
fi
{
  echo '<!doctype html><html lang="en"><meta charset="utf-8"><meta name="viewport" content="width=device-width"><title>Dome documentation</title><h1>Dome documentation</h1><ul>'
  echo '<li><a href="next/">Development guides (next)</a></li>'
  if [[ -n "$stable" ]]; then printf '<li><a href="latest/">Latest stable (%s)</a></li>\n' "$stable"; fi
  while IFS= read -r tag; do
    [[ -z "$tag" ]] && continue
    printf '<li><a href="%s/">%s</a></li>\n' "$tag" "$tag"
  done < <(sort -Vr "$tags_file")
  echo '</ul></html>'
} > "$site/index.html"
go run scripts/check-docs.go "$site" /dome/
