# Dome development container

This follows Stoca's Devbox and Docker Compose setup, with Go tooling only.
Devbox supplies Go 1.25 as a bootstrap, plus gopls, Delve, gofumpt, Task, Git
and ripgrep. Automatic Go toolchain selection uses the Go 1.26.11+ minimum in
go.mod. Keep automatic selection enabled, or provide a compiler satisfying that
minimum explicitly.
The image installs Codex 0.162.0 with Node/npm, matching Stoca, and
[Hunk](https://github.com/modem-dev/hunk) 0.25.0 using its checksum-verified
standalone installer. Run `hunk diff` in the checkout to review changes.
The image also installs Bat, available as `bat` (and Debian's `batcat`).
It installs [broot](https://github.com/Canop/broot/releases/tag/v1.62.0) 1.62.0
from the SHA-256-verified upstream archive, using precompiled Linux amd64/arm64
musl binaries. Run `broot` to browse files; run `broot --install` as the workspace
user to set up the optional `br` shell function for directory navigation.
Python and application-specific servers are not installed.

The Git integration suite builds its own small HTTP/SSH server image from
`tests/integration/gitserver/` using standard Debian packages. No separate Git
forge binary or external server image is needed. Run `task test:integration` to
build and exercise the fixture automatically.

With Docker, Task and the Dev Container CLI installed on the host:

```sh
task devcontainer:up
task devcontainer:shell
task devcontainer:check
task devcontainer:exec -- task test
task devcontainer:codex -- login --device-auth
task devcontainer:codex
task devcontainer:codex -- resume --last
```

VS Code/Cursor **Reopen in Container** can use the same configuration. If the
host selects an older Compose Bake backend, launch the editor with
`COMPOSE_BAKE=false` or use the lifecycle tasks, which select direct BuildKit. The checkout is mounted at
`/workspaces/dome`; Devbox metadata, Go modules and caches persist in named
volumes. On the host, `devbox shell` provides the same tools.

Codex login and session state persist in the ignored `.codex-local` directory
(`CODEX_HOME`). Its user-owned npm installation persists in a named volume.
The launcher uses Stoca's `--no-daemon
--dangerously-bypass-approvals-and-sandbox` options inside the isolated container.

The Go module is `github.com/jani123/dome`. Run `task build`, `task test`,
`task fmt`, or `task check` inside the development shell. Go automatically
downloads the toolchain required by `go.mod` when necessary. Container creation
also works before dependencies are downloaded. See the root README for the
separate `task test:e2e` and `task test:integration` suites.

The workspace uses an isolated internal Docker bridge network. A separate
nginx docs relay joins the private and ingress networks and publishes only the
loopback docs port; the workspace keeps proxy-only egress. HTTP(S) egress goes
through the Squid allowlist in `proxy/allowed-domains.txt`; direct external
traffic and public DNS are blocked. Use HTTPS Git URLs, such as
`https://github.com/jani123/dome.git`; SSH remotes do not use this HTTP proxy.
The allowlist includes GitHub repository/API/release hosts, `docs.github.com`,
and `jani123.github.io` for Dome's documentation site. No host home directory
or Docker socket is mounted, and the docs port is published only on host loopback with a Docker-assigned port.

## Docker-in-Docker

The official Docker-in-Docker Dev Container feature installs Docker Engine,
CLI, Buildx and Compose v2. The feature sets `moby: false` to install Docker CE
because the Devbox base image uses Debian Trixie, where the feature's Moby
packages are unavailable. Its startup hook starts a dedicated daemon inside
the workspace; named volumes preserve Docker and containerd storage across
workspace recreation. The host Docker socket is not used. The `devbox` user
can run Docker without sudo.

The feature runs the workspace in privileged mode. The proxy-only network
topology is retained, but it is not an enforced isolation boundary against
privileged workspace processes. Codex's launcher options are unchanged.

Rebuild to apply the feature and Docker Hub proxy policy:

```sh
task devcontainer:rebuild
task devcontainer:check
task devcontainer:shell
docker pull hello-world
docker run --rm hello-world
```

The image installs `docker-daemon.json` as `/etc/docker/daemon.json`, explicitly
configuring Docker image pulls through the Compose proxy. This survives workspace
recreation and does not depend on the daemon inheriting shell variables. Docker
Hub and GitHub Container Registry hosts are included in the proxy allowlist.
Other registries require adding their authentication, registry and layer hosts
to the allowlist and rebuilding. Build steps and nested containers need explicit
proxy settings; host networking lets them resolve and reach the outer Compose
proxy. For example, from the development shell:

```sh
docker build --network host \
  --build-arg HTTP_PROXY="$HTTP_PROXY" \
  --build-arg HTTPS_PROXY="$HTTPS_PROXY" \
  --build-arg NO_PROXY="$NO_PROXY" .
docker run --rm --network host \
  -e HTTP_PROXY -e HTTPS_PROXY -e NO_PROXY \
  -e http_proxy -e https_proxy -e no_proxy alpine:latest \
  sh -c 'wget -qO- https://proxy.golang.org/'
```

With the default nested bridge, the Compose hostname `proxy` is not available.
The same allowlist applies to proxied build and container traffic. Nested test ports remain separate from the workspace docs binding.

After changing the image or allowlist, run `task devcontainer:rebuild`.
This recreates the workspace container and interrupts active sessions.
Host image builds and host Devbox shells use the host network.

Devbox package resolution creates `devbox.lock`; commit it after the first
successful setup to share the resolved tool versions.

## Maintainer knowledge and specifications

`task dev:setup` also installs the pinned native OpenSpec CLI with checksum
verification. The Devbox shell exposes `.build/tools` and disables OpenSpec
telemetry/update checks. Versioned maintainer skills are already in the checkout;
refresh intentionally with `task specs:setup` after a tool upgrade. Google OKF
knowledge is plain repository Markdown/YAML and needs no cloud credentials or
third-party memory service. See [the maintainer workflow](../docs/development.md).

## Local documentation preview

The Dev Container post-start hook launches Hugo automatically, including after
container restarts, and waits for an HTTP response. Rebuild/reopen the container
after pulling this lifecycle change. Run `task devcontainer:urls` on the host and
open the reported Docs URL. Docker assigns an available localhost port, allowing
multiple workspaces to run concurrently. Hugo listens on container port 1313;
live reload is disabled. Repeated hooks reuse the running server through a
container-local process lock. Startup failures fail the hook; inspect
`.build/docs-preview.log` inside the workspace. To retry without reopening, run
`bash .devcontainer/start-docs.sh` inside the container.

`task docs:serve` remains available for foreground serving outside the automatic
lifecycle (stop the background server first if using the same port).
`task docs:build` builds and checks static guides without serving them.

URL discovery reads current workspace mappings without starting containers.
A mapping is not a readiness check: Hugo must be running. Recreated workspaces
may receive a different port, so rerun the command. Run discovery from the host
using the same checkout path as the Dev Container CLI. With remote Docker, the
loopback binding belongs to the Docker host and requires access or a tunnel.

Image builds forward proxy variables as standard build arguments and use the
daemon host network to reach the proxy. Compose build entitlements make this
explicit. Lifecycle tasks use Compose's direct BuildKit path (`COMPOSE_BAKE=false`)
because older Compose Bake versions do not forward host-network entitlements.
Runtime workspace networking stays private.

## Running inside another restricted devcontainer

A generated devcontainer can itself be tested inside a proxy-restricted parent.
Its image builds use the parent's explicit proxy variables, but its runtime
Squid proxy also needs an explicit upstream when the nested Docker bridge cannot
resolve or directly reach external hosts. In that case append a parent peer to
the nested fixture's `proxy/squid.conf`, replacing the example IP with the
reachable trusted outer proxy address:

```text
cache_peer 192.0.2.10 parent 3128 0 no-query default name=outer-egress
cache_peer_access outer-egress allow approved
never_direct allow all
```

Give the nested proxy a distinct `visible_hostname` so chained HTTP requests
are not treated as a loop; generated modules use their module name by default.
Keep the existing allowlist and destination checks. The outer proxy must allow
the same actual endpoints; this chain does not bypass either policy. Rebuild the
nested proxy after changing its configuration. This is explicit network
customization for nested verification, not an automatic download fallback.
