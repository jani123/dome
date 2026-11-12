# Dome

Dome is a Go library for compiling your own infrastructure agent. Resources and playbooks are defined in Go; YAML supplies typed configuration values, and SOPS supplies encrypted secrets. Modules register explicitly, without `init()` registration or runtime plugins.

Module: `github.com/jani123/dome`.

## Build and try the example

Use Linux and Go 1.26 or newer. `go.mod` pins dependency versions; Go 1.25 with automatic toolchain downloads enabled can bootstrap the required toolchain. Docker is needed only for container tests.

```sh
make build
mkdir -p /tmp/dome-example
printf 'content: hello\nversion: "1"\n' > .build/config.yaml
.build/agent --state .build/example.db --config .build/config.yaml plan --json
.build/agent --state .build/example.db --config .build/config.yaml apply --json
.build/agent --state .build/example.db status
```

The example manages `/tmp/dome-example/managed.txt` and optionally `version.txt`. Set `root` in configuration to an existing absolute directory, or set `DOME_ROOT`. Parent directories must already exist. File permissions are explicit: `0640` for the managed file and `0600` for the version file. `managed: false` removes the managed file from desired state, exercising ownership-based pruning. `fail_transition: true` demonstrates a failed version transition.

```yaml
root: /tmp/dome-example
content: hello
managed: true
version: "1"
fail_transition: false
```

The example's topology and workflow callbacks live in `examples/basic-agent/main.go`; these fields are that application's configuration, not a framework infrastructure DSL. Unknown YAML fields, malformed input, and invalid directories fail before reconciliation.

## Compile your own agent

In a separate Go module, import Dome and implement a typed desired-state function:

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/jani123/dome/agent"
    "github.com/jani123/dome/cli"
    "github.com/jani123/dome/config"
    "github.com/jani123/dome/core/resource"
    "github.com/jani123/dome/modules/file"
)

type Settings struct { Content string `yaml:"content"` }
type Secrets struct { Token string `yaml:"token"` }

func main() {
    a, err := agent.New(
        agent.WithModules(file.Module{}),
        agent.WithSQLite("/var/lib/my-agent/state.db"),
        agent.WithDesiredState(func(ctx context.Context, doc config.Document) ([]resource.Resource, error) {
            settings, _, err := config.Parse[Settings, Secrets](ctx, doc, nil)
            if err != nil { return nil, err }
            return []resource.Resource{
                &file.File{Path: "/etc/my-app.conf", Content: []byte(settings.Content), Mode: 0640},
            }, nil
        }),
    )
    if err == nil { err = a.Execute(context.Background()) }
    if err != nil {
        fmt.Fprintln(os.Stderr, agent.ErrorText(err))
        os.Exit(cli.ExitCode(err))
    }
}
```

Use a signal-aware context for a long-running agent, as the example does. `WithSource` sets a programmatic local or Git source; CLI source flags override it. Configuration is evaluated once for each plan, reconciliation, or playbook invocation. Library callers can use `Plan`, `Apply`, and `RunPlaybook` directly with `cli.Options{}`.

Third-party modules implement `resource.Module` and `resource.Resource` using public packages. `Module.Restore` rebuilds a validated deletion handle from a previously managed identity; it must not need old secret values. Resource identities are stable `(module, name)` pairs. Modules must validate unsafe inputs, report truthful observations, and make their default `Apply` safe to repeat. Filesystem and process interactions in the built-in modules have injectable interfaces.

## Reconciliation and state

Dome validates all resources and observes all planned changes before mutating the host. It applies desired changes before pruning resources removed from desired state. Only previously managed identities can be deleted; unrelated paths are untouched. A missing registered module blocks reconciliation rather than abandoning ownership silently.

Each resource follows `Observe → Decide → Apply → Verify → Persist`. Verification compares a fresh observation with the desired snapshot. SQLite records ownership, non-secret applied snapshots, configuration revision, resource operations, DAG executions, individual node attempts, errors by category, and locks. Resource state advances only after verification; the overall Git revision advances only after the entire reconciliation succeeds. Earlier successful resources can remain changed when a later resource fails; there is no automatic host rollback.

Snapshots are public diagnostics and durable data. Custom modules must use non-secret values, such as hashes, and must keep secrets out of identifiers, errors, and metadata. The file module persists content hashes and permissions, never file contents. Configuration blobs, decrypted secrets, and authentication material are not stored in SQLite.

SQLite uses `modernc.org/sqlite`, versioned transactional migrations, foreign keys, WAL, and a busy timeout. An exclusive process lock serializes invocations sharing one state database and is released by the kernel after a crash. Within a DAG, independent nodes execute concurrently and named resource locks prevent overlapping steps. Agents with separate state databases must not manage the same host resources. Direct users of `state.Store` and `dag.Engine` must hold `Store.Guard` while running operations.

## Playbooks and transitions

```go
book := playbook.New("deploy", &playbook.Step{
    ID: "write-config",
    Dependencies: []string{}, // explicit graph edges, never list order
    Locks: []string{"app-config"},
    Repeatable: true,
    Timeout: 30 * time.Second,
    Retry: playbook.RetryPolicy{Attempts: 3, Delay: time.Second},
    Observe: observeConfig, // returns Satisfied, NeedsApply, InProgress, Failed, or Unknown
    Apply: writeConfig,
    Verify: verifyConfig,
})
```

Register with `agent.WithPlaybooks(book)`. Every step needs `Observe`, `Apply`, and `Verify`, and must either declare `Repeatable: true` or supply `Recover`; invalid recovery contracts are rejected before execution. Retry attempts are bounded per invocation. Attempts are checkpointed before external effects. On resume, even completed steps are re-observed and verified before skipping. Safely repeatable steps can retry; interrupted non-repeatable steps and `Unknown`/`Failed` observations require an explicit `Recover` callback to establish a safe state. `InProgress` is never blindly applied again. Callbacks must honor context cancellation for timeouts to take effect.

`OnFailure` can point to an explicit recovery step. The branch's progress is persisted independently, and successful recovery does not hide the original failure or unblock its dependent nodes. Failed dependencies block their descendants; independent branches may finish. Bump the playbook's `Version` whenever callback behavior changes. Resume rejects a changed graph fingerprint.

A resource can implement `resource.Transition` to select a registered playbook instead of its default apply operation. Without a selected playbook, it must explicitly declare its default transition safe. This also applies to deletion handles during pruning. The reconciler durably links the resource identity, target snapshot, playbook, and execution ID before side effects, holds the database guard through the workflow, and verifies the final resource state before persisting it. An unfinished DAG is resumed even if the resource already matches its target. Changing or removing that target is blocked until its original transition completes. Interrupted pruning workflows also resume after the file has already disappeared. The example's `demo-version` resource demonstrates this mechanism; PostgreSQL is not implemented.

```sh
DOME_ROOT=/tmp/dome-example .build/agent --state .build/example.db playbook list
DOME_ROOT=/tmp/dome-example .build/agent --state .build/example.db playbook run parallel-demo
.build/agent --state .build/example.db operations list
.build/agent --state .build/example.db operations inspect EXECUTION_ID
.build/agent --state .build/example.db playbook run parallel-demo --resume EXECUTION_ID
```

## Configuration from Git and SOPS

Git sources support SSH with a private key and a verified `known_hosts` file, or public HTTPS. Unencrypted HTTP requires `config.GitOptions.AllowHTTP` or the CLI flag `--git-allow-http`. Set the repository-relative configuration path and branch, tag, or reachable commit. Fetching an explicit branch works even if the server's default branch is unborn. Retrieval reads committed blobs into memory without a checkout.

```sh
.build/agent --state .build/git.db apply --json \
  --git-url ssh://git@git.example.com:23231/configuration \
  --git-ref main --git-path config.yaml \
  --ssh-key /etc/dome/git.key --known-hosts /etc/dome/known_hosts
```

Obtain the server's host public key through a trusted channel before creating `known_hosts`. SSH host verification is mandatory. Private HTTPS authentication, encrypted SSH private keys, submodules, and Git LFS configuration blobs are not supported in this slice. Retrieval currently fetches branch/tag history afresh in memory, so configuration repositories should be small.

SOPS uses its Go decryption API and the normal host key environment, such as `SOPS_AGE_KEY_FILE=/etc/dome/age.key`. Secrets are decoded into a separate user-defined Go struct. The example accepts a `token` field to exercise loading but does not write it to its managed file.

```sh
SOPS_AGE_KEY_FILE=/etc/dome/age.key .build/agent apply \
  --state .build/example.db --config .build/config.yaml --secrets secrets.enc.yaml
```

With `--git-url`, `--secrets secrets.enc.yaml` is a path inside the selected commit. Retrieval, decryption, and validation failures abort before reconciliation or pruning. Reconciliation failures retain the last successfully applied overall revision.

`agent run` reconciles once by default. `agent run --interval 30s` polls and reconciles repeatedly; a failure exits so a supervisor can report and restart it. This is polling rather than a webhook service.

## CLI

Use `--help` on any command. Plans and operation status support JSON; status and listings already print JSON. Exit codes are `0` for success, `2` for validation/configuration failure, and `1` for execution failure.

```text
agent run [--interval 30s]
agent plan [--json]
agent apply [--json]
agent status
agent modules
agent playbook list
agent playbook run <name> [--resume <execution-id>]
agent operations list
agent operations inspect <id>
agent version
```

## Tests

```sh
make check           # formatting, vet, mocked unit tests
make test-race       # race detector
make test-e2e        # Ginkgo/Gomega + Testcontainers, real agent binary
make test-integration # real boundaries plus the separate required Git HTTP/SSH suite
```

The core E2E suite covers file creation/update/permissions, no-op reconciliation, ownership pruning, unrelated files, parallel branches, process interruption before SQLite completion, resume, resource transitions, failed transitions, invalid configuration, restart persistence, and SOPS with correct/incorrect host keys. Default unit tests use Testify assertions and mocks; retry waits, deadlines, lock contention, and CLI polling are tested using injected time. Tests involving real SQLite, atomic filesystem operations, SOPS host keys, or standalone agent files are tagged `integration` and run by `make test-integration` against isolated temporary storage. `make test-e2e` runs the independent container suite, including both standalone DAG and resource transition crashes before SQLite completion.

The integration suite creates a repository over SSH, pushes real Git commits over SSH or smart HTTP, starts the agent in another container, reconciles updated configuration, verifies revision persistence and idempotency, restarts the container, and tests retrieval, validation, decryption, and transition failures. Containers, repositories, keys, networks, and temporary state volumes are cleaned up. There are no fake-server substitutions or automatic skips.

The Git fixture image is built automatically by Testcontainers from `tests/integration/gitserver/`. It uses `debian:bookworm-slim` with standard Git, Apache CGI, OpenSSH, and Tini packages. Apache invokes Git's own `git-http-backend` for smart HTTP, and OpenSSH serves Git over SSH. The fixture needs no forge service, host-binary injection, copied libraries, external Git account, or GHCR server image.

The suite runs the same push/update/reconcile/persistence/failure flow over both transports. Repository provisioning uses SSH with a generated test key. HTTP access is anonymous, with push explicitly enabled on the isolated test repository; the test image is not a production hosting configuration. The agent rejects unencrypted HTTP unless `--git-allow-http` is supplied. SSH tests separately verify that incorrect client keys and unknown host keys are rejected.

Image builds forward the configured proxy environment as standard Docker build arguments and use the host build network so the devcontainer's proxy hostname is reachable. Runtime test containers use a separate isolated bridge network. The harness adds the Docker endpoint to its own `NO_PROXY` settings before starting HTTP clients so published test ports are reached directly. The proxy must allow Docker Hub image pulls and Debian package repositories; no external network is needed for Git traffic after the image is built. The local image and Docker build cache are retained for subsequent runs, while containers, repositories, keys, and temporary volumes are cleaned up.

The systemd module is tested with an injected command runner. Live systemd tests are not part of the ordinary container suite because its containers do not run systemd. It supports loaded `.service` units with enabled/disabled and active/inactive/failed states; other unit states fail explicitly. Pruning stops and disables a service rather than uninstalling it.

## Package layout and next work

- `agent`, `cli`, `config`: public composition, Cobra commands, typed local/Git/SOPS loading.
- `core/resource`, `core/registry`, `core/reconcile`: resource contracts, registration, planning, verification, and ownership pruning.
- `core/state`, `core/playbook`, `core/dag`, `core/fault`: SQLite, typed graphs, execution/recovery, contextual errors.
- `modules/file`, `modules/systemd`: initial infrastructure modules.
- `examples/basic-agent`: compiled consumer and demonstration transition resource.
- `tests/e2e`, `tests/integration`: independent container suites.

## Plan verification

This table maps the agreed scope to implementation and executable evidence. The required checks are `make check`, `make test-race`, `make test-e2e`, and `make test-integration`; the integration target includes the real boundary tests throughout the module as well as the separate Git server suite.

| Plan requirement | Implementation and evidence |
| --- | --- |
| Architecture, explicit public API, external modules (§1–2, §13) | `agent`, `core/resource`, `core/registry`; `tests/integration/consumer_test.go` compiles a separate Go module with a custom resource. |
| File/systemd resources, observe/plan/apply/verify, ownership and pruning (§3) | `modules/file`, `modules/systemd`, `core/reconcile`; mocked module/reconciliation tests, atomic filesystem integration tests, container create/update/no-op/prune/unrelated-file scenarios. |
| Safe transitions and applied-state verification (§4) | Durable resource/target/playbook execution intents; unit tests for unsafe transitions, failed verification and interrupted finalization; container process-crash recovery after the version file changed. |
| DAG validation, concurrency, dependencies, locks, retries and recovery branches (§5) | `core/playbook` and `core/dag` tests cover cycles, missing dependencies, rolling scheduling, conflicting locks, retry policies, deadlines, failure blocking and explicit recovery. |
| Observe-led crash recovery and stable execution/node IDs (§6) | Injected checkpoint failure tests and real container process exits after external effects, before completion persistence; completed work is verified rather than blindly repeated. |
| SQLite migrations, ownership, revisions, operations, attempts and locks (§7) | `core/state/sqlite_integration_test.go` covers persistence, schema upgrades, foreign keys, atomic lock acquisition, process guards, canonical paths and attempt history. |
| Typed YAML, SOPS host secrets, Git retrieval, failure safety (§8) | Mocked loaders, real SOPS boundary tests, container host-key tests, and real HTTP and authenticated SSH repository integration. |
| CLI commands, JSON, validation/execution exit codes (§9) | `cli` tests cover every requested command, polling, invalid commands/flags, source conflicts, listings and JSON plans/status. |
| Contextual cause-preserving errors and stable categories (§10) | `core/fault`, module boundary wrappers and tests; DAG errors carry execution/node context without masking observation or verification categories. |
| Testify/mock unit suite alongside testable packages (§11) | Default `go test ./...`; real I/O boundary tests are explicitly tagged `integration` rather than mixed into the unit suite. |
| Ginkgo/Gomega/Testcontainers E2E and required real Git server suite (§12) | `tests/e2e` exercises the compiled agent inside Linux containers; `tests/integration/git_test.go` creates/pushes real HTTP/SSH repositories and verifies updates, revisions, restart, secrets and failure handling. |
| Milestones and definition of done (§14–15) | The commands above verify the runnable example, external module registration, persistence, reconciliation/pruning, DAG recovery, transitions, Git and CLI. |
| Build, usage, tests and limitations report (§16) | Build/run examples and supported/unsupported behavior are documented in this README. |

Live systemd integration remains conditional under milestone 5: ordinary test containers do not run systemd, and no privileged host access is required. PostgreSQL was explicitly excluded from the initial slice. Incremental Git caching, private HTTPS credentials, submodules, LFS, and webhook delivery are not requirements of this plan. Custom modules remain responsible for safe default operations and secret-free snapshots and diagnostics.
