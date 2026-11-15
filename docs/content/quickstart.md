---
title: Quickstart
weight: 10
---

## Prerequisites

- Linux with Docker.
- Dev Container CLI, or VS Code with the Dev Containers extension.
- Optional: Go 1.26.11 or newer for direct project creation.
- Module-download access.
- Permission to access the target machine.
- Optional: Codex, the initially supported coding agent.

## Bootstrap

### 1. Create the project

{{< bootstrap >}}

`example.com/my-machine` names your new Go project; it does not need to exist
online.

### 2. Open the development container

From the host, in `my-machine`:

{{< open-devcontainer >}}

Run all remaining commands in the container terminal at `/workspace`.

### 3. Set the file contents

In `config.yaml`:

```yaml
content: "Hello from Dome\n"
managed: true
```

### 4. Define the file

In `cmd/agent/main.go`, use this `desired` function with the generated imports
and `Settings` and `Secrets` types:

```go
func desired(ctx context.Context, doc config.Document) ([]resource.Resource, error) {
    settings, _, err := config.Parse[Settings, Secrets](ctx, doc, nil)
    if err != nil {
        return nil, err
    }
    if !settings.Managed {
        return nil, nil
    }
    root, err := filepath.Abs("sandbox")
    if err != nil {
        return nil, fault.Wrap(fault.Validation, err)
    }
    return []resource.Resource{
        &file.File{
            Path:    filepath.Join(root, "hello.txt"),
            Content: []byte(settings.Content),
            Mode:    0o640,
        },
    }, nil
}
```

The generated `main` registers the file module:

```go
agent.WithModules(file.Module{}),
```

### 5. Build and plan

Run inside the devcontainer:

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

To reopen the CLI terminal later, run `devcontainer exec --workspace-folder . bash`
from the host project directory. In VS Code, use Reopen in Container again.
