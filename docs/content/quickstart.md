---
title: Quickstart
weight: 10
---

## Prerequisites

- Linux with Go 1.26.11 or newer.
- Module-download access.
- Permission to access the target machine.
- Optional: Codex, the initially supported coding agent.

## Bootstrap

### 1. Create the project

```sh
go run github.com/jani123/dome/cmd/dome@{{< dome-version >}} init my-machine --module example.com/my-machine
cd my-machine
```

### 2. Set the file contents

In `config.yaml`:

```yaml
content: "Hello from Dome\n"
managed: true
```

### 3. Define the file

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

### 4. Build and plan

Run from `my-machine`:

```sh
CGO_ENABLED=0 go build -o agent ./cmd/agent
go test ./...
./agent init --config config.yaml --non-interactive
./agent plan --json
```

### 5. Create the file

After reviewing the plan:

```sh
./agent apply --json
cat sandbox/hello.txt
```

```text
Hello from Dome
```

### 6. Run it again

```sh
./agent apply --json
```

The second apply reports `no-op`.
