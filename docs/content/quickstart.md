---
title: Quickstart
weight: 10
---

## Prerequisites

- Linux with Go 1.26.11 or newer, or Docker.
- Optional: Dev Container CLI, or VS Code with the Dev Containers extension.
- Module-download access.
- Permission to access the target machine.
- Optional: Codex, the initially supported coding agent.

## Bootstrap

### 1. Create the project

{{< bootstrap >}}

`example.com/my-machine` names your new Go project; it does not need to exist
online.

### 2. Open the development container (optional)

If you want to develop in a container, run from the host in `my-machine`:

{{< open-devcontainer >}}

Otherwise, continue with local Go in `my-machine`. Run the remaining commands
from that directory, or from `/workspace` in the devcontainer terminal.

### 3. Set the file contents

{{< file-contents >}}

### 4. Define the file

{{< define-file >}}

### 5. Build and plan

Run from the project directory, locally or inside the devcontainer:

```sh
CGO_ENABLED=0 go build -o agent ./cmd/agent
go test ./...
./agent init --config config.yaml --non-interactive
./agent plan --json
```

### 6. Create the file

After reviewing the plan:

```sh
./agent apply --json
cat sandbox/hello.txt
```

```text
Hello from Dome
```

### 7. Run it again

```sh
./agent apply --json
```

The second apply reports `no-op`.

### 8. Keep the agent running

From the same project directory and environment:

```sh
# Reconcile immediately, then wait 30 seconds between successful runs.
./agent run --interval 30s
# Press Ctrl+C when finished.
```

The agent stays in the foreground. Edit `config.yaml` in another terminal to
change the file contents; the next reconciliation updates `sandbox/hello.txt`.
A reconciliation failure exits the process. See [deployment](../deployment/) for
running under a supervisor.
