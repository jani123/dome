---
title: Deploy your compiled agent
weight: 50
---

Dome releases provide a library and initializer. Your agent binary is built and
distributed by your project. Coding agents are authoring tools, not a runtime
requirement. Initial targets are Linux amd64 and arm64; native release smoke tests
must pass for a platform before it is advertised as released.

Build on the development machine, explicitly targeting the deployment machine:

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o agent ./cmd/agent
```

Use `arm64` for that architecture. Cross-compilation does not replace native tests.
Deliver the binary with its exact source revision, dependency files, build scripts,
license/notices, and the customer's configuration. See the licensing guide.

Provision a stable working directory, resource parent directories, SQLite state
storage, configuration, SOPS keys, and verified Git credentials on the target.
Restrict key/state permissions. Run with only the permissions the resources need.
The sandbox starter uses relative paths, so set its working directory explicitly.
For real deployments, prefer explicit stable absolute resource/state paths.

Run `agent init` once in the stable working directory, or provision a complete
runtime profile with `agent init --agent-config settings.yaml --non-interactive`.
The managed host saves only runtime settings and credential-file references;
Git-backed resource configuration stays in its existing repository. Select the
existing SQLite path when adopting this workflow on an already managed host.
Initialization never resets or migrates execution history.

Supervise `agent --workspace /path/to/workspace run` with systemd or another host supervisor.
The process stays in the foreground, reconciles immediately, and waits 30 seconds
after each successful reconciliation. Use a positive `--interval 1m` to override
the wait. Stop it with Ctrl+C or the supervisor’s termination signal. For one-shot
execution, use `apply`; migrate old one-shot `run` and `run --interval 0`
invocations to `apply`. Zero and negative intervals are rejected.
A reconciliation failure exits; configure restart/reporting behavior deliberately.
There is no embedded webhook service or automatic host rollback. Never let two
state databases manage the same host resources. Preserve SQLite across restarts.
