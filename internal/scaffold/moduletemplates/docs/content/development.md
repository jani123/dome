---
title: Development
---

Run `task dev:setup`, `task check`, `task test:race`, `task check:artifacts`.
Docker is required for `task test:integration` and `task test:e2e`; requested
suites fail if it is unavailable. Tests build tests/fixtures/debian/Dockerfile,
install normal packages, and run real probes/consumer agents in disposable storage.
Customize that fixture and tests to exercise your backend. Containers clean up
on test exit; use bounded contexts and collect logs on failures. Ordinary Debian
containers do not run systemd. Never replace failed real dependencies with stubs.

Optional devcontainer: `task devcontainer:up`, `task devcontainer:shell`,
`task devcontainer:check`. Root tasks work on the host without Devbox.
Hugo starts automatically with the devcontainer. Discover the assigned host port
with `task devcontainer:urls`; inspect `.build/docs-preview.log` for startup failures. Rebuild for initial port mapping; mappings do not
prove the server is running. Remote Docker hosts require access/tunnelling.

Run `task specs:setup` to intentionally refresh the seven pinned OpenSpec workflows;
they do not overwrite the separate dome-module skill. Knowledge is Google OKF v0.2,
plain source-linked Markdown with a local validator and no hosted memory engine.

When testing this devcontainer inside another restricted workspace, its runtime
proxy may need an explicitly configured parent peer. Follow the nested-host
section of `.devcontainer/README.md`; keep both allowlists enforced.
