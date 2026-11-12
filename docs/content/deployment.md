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

Supervise `agent run --interval 30s` with systemd or another host supervisor.
A reconciliation failure exits; configure restart/reporting behavior deliberately.
There is no embedded webhook service or automatic host rollback. Never let two
state databases manage the same host resources. Preserve SQLite across restarts.
