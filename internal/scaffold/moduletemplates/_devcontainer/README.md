# Development container

Devbox supplies Go 1.25 as a bootstrap; automatic toolchain selection uses the
Go 1.26.11+ minimum in go.mod. Keep automatic selection enabled or provide a
sufficient compiler explicitly. Devbox tools, non-root workspace,
Docker-in-Docker, bat/Hunk and optional Codex
follow the source Dome developer environment. broot 1.62.0 is installed from
SHA-256-verified precompiled Linux amd64/arm64 binaries. Run `broot` or `broot --install` to set up the `br` shell
function. Host Docker and Dev Container CLI
are required. Docker-in-Docker is privileged; this is not a hardened isolation
boundary. Credentials and sessions belong in ignored local storage.

Run task devcontainer:up, task devcontainer:shell, task devcontainer:check.
Post-create runs task dev:setup. To preview guides use
`task devcontainer:exec -- task docs:serve`, then `task devcontainer:urls` from
another host terminal. Rebuild after port/proxy changes. Mappings are not readiness
checks; a remote Docker daemon's localhost binding needs a tunnel.

The workspace's private network uses proxy/allowed-domains.txt for explicit HTTP(S)
egress. Add module-specific endpoints there and rebuild; do not bypass the proxy
or replace blocked downloads. Docker builds receive proxy arguments and use host
networking to reach the outer proxy. Nested test containers stay isolated; exempt
published test endpoints from proxying. Cache volumes are devcontainer-scoped.

Host development and CI can run root Task commands without Devbox. The tests'
Debian fixture is customizable and does not run systemd by default.

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
