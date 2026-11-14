#!/usr/bin/env bash
set -euo pipefail
root=${1:?workspace root required}
workspaces=$(docker ps --filter "label=devcontainer.local_folder=$root" \
  --filter label=com.docker.compose.service=workspace --format '{{.ID}}')
if [[ -z "$workspaces" ]]; then
  echo 'No running workspace found. Run task devcontainer:up on the Docker host.'
  exit 0
fi
found=false
declare -A seen=()
while IFS= read -r workspace; do
  project=$(docker inspect --format '{{index .Config.Labels "com.docker.compose.project"}}' "$workspace")
  if [[ -z "$project" ]]; then
    ids=$workspace
  else
    ids=$(docker ps --filter "label=com.docker.compose.project=$project" --format '{{.ID}}')
  fi
  while IFS= read -r id; do
    [[ -n "$id" && -z "${seen[$id]:-}" ]] || continue
    seen[$id]=1
    service=$(docker inspect --format '{{index .Config.Labels "com.docker.compose.service"}}' "$id")
    [[ "$service" != proxy ]] || continue
    bindings=$(docker port "$id")
    while IFS= read -r binding; do
      [[ -n "$binding" ]] || continue
      found=true
      port=${binding%% *}
      address=${binding##* -> }
      if [[ "$port" == 1313/tcp && ( "$service" == docs || "$service" == workspace ) ]]; then
        printf 'Docs: http://%s/ (%s %s)\n' "$address" "$service" "$id"
      else
        printf '%s %s: %s\n' "$service" "$port" "$address"
      fi
    done <<< "$bindings"
  done <<< "$ids"
done <<< "$workspaces"
if [[ "$found" == false ]]; then
  echo 'Workspace is running with no published ports. Run task devcontainer:rebuild to apply port configuration.'
else
  echo 'Mappings do not prove server readiness. Start Hugo with task devcontainer:exec -- task docs:serve.'
  echo 'For a remote Docker daemon, localhost refers to that host; use an SSH tunnel.'
fi
