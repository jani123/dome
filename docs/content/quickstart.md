---
title: Quickstart
weight: 10
---

## Prerequisites

Use Linux with Go 1.26 or newer. Codex is the initially supported optional coding
agent. You need module-download access and permission to access the target machine.
The starter manages a local sandbox and needs no root permissions.

## Bootstrap

Use the published release matching these guides. Development guides show the
planned first stable version, which requires publication before it can resolve:

```sh
go run github.com/jani123/dome/cmd/dome@{{< dome-version >}} init my-machine --module example.com/my-machine
cd my-machine
CGO_ENABLED=0 go build -o agent ./cmd/agent
go test ./...
./agent plan --json
```

The command refuses existing destinations. To add Dome alongside an existing
project, generate a new subdirectory with its own module. Go dependencies and
matching Codex skills are installed together. Generation never applies resources.
Until the first release, see the repository README's local development procedure;
these release commands cannot resolve an unpublished version.

Inspect the plan, then when ready:

```sh
./agent apply --json
./agent apply --json
./agent status
```

The first apply creates `sandbox/hello.txt` with permissions `0640`. The next
apply reports `no-op`. `plan` observes resources without applying host changes,
but may create SQLite state. Keep the sandbox directory and run the starter from
its project root so relative state/config paths and resource identity stay stable.

Edit `config.yaml` to change the content. Set `managed: false` to prune the
previously owned file. Unrelated files remain untouched. Do not delete the state
database to reset an agent that owns real resources.
