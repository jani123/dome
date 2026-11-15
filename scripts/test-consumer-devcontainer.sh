#!/usr/bin/env bash
# Explicit real-CLI acceptance; needs Docker, Node, and the Dev Container CLI.
set -euo pipefail
repo_root=$PWD
cli=${DOME_DEVCONTAINER_CLI:-devcontainer}
command -v "$cli" >/dev/null || { echo 'Set DOME_DEVCONTAINER_CLI to the real Dev Container CLI executable' >&2; exit 1; }
mkdir -p .build
fixture=$(mktemp -d "$repo_root/.build/consumer-quickstart.XXXXXX")
project="$fixture/my-machine"
container_id=''
cleanup() {
  if [[ -n "$container_id" ]]; then docker rm --force "$container_id" >/dev/null; fi
}
trap cleanup EXIT
# Explicitly route this restricted test environment's configured proxy hostname.
# Neither credentials nor environment-specific addresses enter generated assets.
proxy_args=()
proxy_hosts=()
for variable in HTTP_PROXY HTTPS_PROXY; do
  if [[ -n "${!variable:-}" ]]; then
    proxy_host=$(PROXY_ENV_NAME="$variable" node -e 'process.stdout.write(new URL(process.env[process.env.PROXY_ENV_NAME]).hostname)')
    proxy_ip=$(getent ahostsv4 "$proxy_host" | awk 'NR == 1 { value=$1 } END { print value }')
    [[ -n "$proxy_ip" ]] || { echo 'Configured proxy hostname cannot be resolved' >&2; exit 1; }
    proxy_args+=(--add-host "$proxy_host:$proxy_ip")
    proxy_hosts+=("--add-host=$proxy_host:$proxy_ip")
  fi
done
# Local-source equivalent of the documented @main bootstrap. The published
# library pin lets this pre-publication check use real dependencies without stubs.
docker run --rm --user "$(id -u):$(id -g)" --env HOME=/tmp --env CGO_ENABLED=0 \
  --env HTTP_PROXY --env HTTPS_PROXY --env NO_PROXY \
  "${proxy_args[@]}" --volume "$repo_root:/source:ro" --volume "$fixture:/bootstrap" \
  --workdir /source golang:1.26.11-bookworm \
  go run -ldflags '-X main.version=v0.1.0-rc.1' ./cmd/dome \
  init /bootstrap/my-machine --project-module example.com/my-machine \
  > "$fixture/bootstrap.log" 2>&1

# Exercise the exact YAML and Go example in the guide, not a separate fixture.
node - "$repo_root" "$project" "${proxy_hosts[@]}" <<'JS'
const fs = require('node:fs');
const [root, project, ...hosts] = process.argv.slice(2);
const guide = fs.readFileSync(`${root}/docs/content/quickstart.md`, 'utf8');
const fences = [...guide.matchAll(/^```(yaml|go)\n([\s\S]*?)^```/gm)];
fs.writeFileSync(`${project}/config.yaml`, fences.find(fence => fence[1] === 'yaml')[2]);
const example = fences.find(fence => fence[1] === 'go')[2];
const path = `${project}/cmd/agent/main.go`;
const source = fs.readFileSync(path, 'utf8');
const start = source.indexOf('func desired(');
const end = source.indexOf('\nfunc main(', start);
if (start < 0 || end < 0) throw new Error('Generated desired-state function not found');
fs.writeFileSync(path, source.slice(0, start) + example.trimEnd() + '\n' + source.slice(end));
if (hosts.length) {
  const config = JSON.parse(fs.readFileSync(`${project}/.devcontainer/devcontainer.json`, 'utf8'));
  config.runArgs = [...new Set(hosts)];
  fs.writeFileSync(`${project}/.devcontainer/acceptance.json`, JSON.stringify(config, null, 2));
}
JS
args=(--workspace-folder "$project")
if [[ -f "$project/.devcontainer/acceptance.json" ]]; then args+=(--override-config "$project/.devcontainer/acceptance.json"); fi
# Limit cleanup to the exact unique fixture, including a failed partial startup.
cleanup() {
  if [[ -n "$container_id" ]]; then
    docker rm --force "$container_id" >/dev/null
  else
    while IFS= read -r owned; do
      [[ -z "$owned" ]] || docker rm --force "$owned" >/dev/null
    done < <(docker ps -aq --filter "label=devcontainer.local_folder=$project")
  fi
}
env -u LD_LIBRARY_PATH "$cli" up "${args[@]}" > "$fixture/up.json" 2> "$fixture/up.log"
container_id=$(node -e 'process.stdout.write(JSON.parse(require("fs").readFileSync(process.argv[1],"utf8")).containerId)' "$fixture/up.json")
execute() { env -u LD_LIBRARY_PATH "$cli" exec "${args[@]}" "$@"; }
execute bash -euc 'test "$(id -un)" = dev; test "$(pwd)" = /workspace; go version; CGO_ENABLED=0 go build -o agent ./cmd/agent; go test ./...; ./agent init --config config.yaml --non-interactive; test ! -e sandbox/hello.txt; ./agent plan --json; test ! -e sandbox/hello.txt; ./agent apply --json; cat sandbox/hello.txt; ./agent apply --json' > "$fixture/walkthrough.log" 2>&1
# Stop/start the generated workspace and verify stable resource and state paths.
docker stop "$container_id" > /dev/null
env -u LD_LIBRARY_PATH "$cli" up "${args[@]}" > "$fixture/reopen.json" 2> "$fixture/reopen.log"
execute ./agent apply --json > "$fixture/repeated-apply.json" 2> "$fixture/repeated-apply.log"
node - "$project" "$fixture" "$(id -u)" "$(id -g)" <<'JS'
const fs = require('node:fs');
const assert = require('node:assert/strict');
const [project, fixture, uid, gid] = process.argv.slice(2);
assert.equal(fs.readFileSync(`${project}/sandbox/hello.txt`, 'utf8'), 'Hello from Dome\n');
for (const relative of ['go.mod', 'cmd/agent/main.go', 'agent', 'sandbox/hello.txt', '.dome/state.db']) {
  const stat = fs.statSync(`${project}/${relative}`);
  assert.equal(stat.uid, Number(uid), relative);
  assert.equal(stat.gid, Number(gid), relative);
}
assert.match(fs.readFileSync(`${fixture}/walkthrough.log`, 'utf8'), /no-op/);
assert.match(fs.readFileSync(`${fixture}/repeated-apply.json`, 'utf8'), /no-op/);
JS
echo "Devcontainer quickstart passed; local verification artifacts: $fixture"
