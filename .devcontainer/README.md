# Dome development container

This follows Stoca's Devbox and Docker Compose setup, with Go tooling only.
Devbox supplies Go 1.25, gopls, Delve, gofumpt, Task, Git and ripgrep.
The image installs Codex 0.162.0 with Node/npm, matching Stoca.
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

VS Code/Cursor **Reopen in Container** also works. The checkout is mounted at
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

The workspace uses an internal Docker network. HTTP(S) egress goes
through the Squid allowlist in `proxy/allowed-domains.txt`; direct external
traffic and public DNS are blocked. Use HTTPS Git URLs, such as
`https://github.com/jani123/dome.git`; SSH remotes do not use this HTTP proxy.
The allowlist includes GitHub repository/API/release hosts, `docs.github.com`,
and `jani123.github.io` for Dome's documentation site. No host home directory
or Docker socket is mounted, and no application ports are published.

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
The same allowlist applies to proxied build and container traffic. No additional
ports are published by this setup.

After changing the image or allowlist, run `task devcontainer:rebuild`.
This recreates the workspace container and interrupts active sessions.
Host image builds and host Devbox shells use the host network.

Devbox package resolution creates `devbox.lock`; commit it after the first
successful setup to share the resolved tool versions.

## Maintainer knowledge and specifications

`task dev:setup` also installs the pinned native OpenSpec CLI with checksum
verification. The Devbox shell exposes `.build/tools` and disables OpenSpec
telemetry/update checks. Versioned maintainer skills are already in the checkout;
refresh intentionally with `make setup-specs` after a tool upgrade. Google OKF
knowledge is plain repository Markdown/YAML and needs no cloud credentials or
third-party memory service. See [the maintainer workflow](../docs/development.md).
