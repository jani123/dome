# Go Infrastructure Agent — Implementation Plan

## Objective

Build a modular, Go-native infrastructure automation framework combining declarative reconciliation, persistent Terraform-like state, and idempotent DAG-based playbooks.

The framework is distributed as a Go library. Users create their own `main.go`, explicitly import the modules they need, and compile a standalone agent binary.

The first implementation must deliver a working vertical slice with unit tests and container-based end-to-end tests.

Do not stop after implementing interfaces or scaffolding. Complete the functionality and verify it.

## 1. Architecture

Separate the framework into independently testable packages:

- **Core:** Resource contracts, identities, ownership, reconciliation, planning, and pruning.
- **Registry:** Explicit registration of independently developed modules.
- **State:** Persistent SQLite storage, migrations, and execution history.
- **DAG:** Dependency validation, scheduling, concurrency, checkpointing, and recovery.
- **Playbook:** Imperative workflows and idempotent steps.
- **Configuration:** Typed YAML loading, validation, and SOPS integration.
- **CLI:** Cobra-based commands and diagnostics.
- **Modules:** Concrete infrastructure capabilities.

Keep the core independent of concrete modules.

Modules must be usable from external Go repositories without modifying the framework.

Do not use Go's runtime plugin system, global registration through `init()`, Yaegi, or a custom DSL.

## 2. Public Go API

Users explicitly construct the agent and register modules.

Target API:

```go
package main

import (
    "context"
    "log"

    "github.com/example/infra/agent"
    "github.com/example/infra/modules/file"
    "github.com/example/infra/modules/systemd"
)

func main() {
    a, err := agent.New(
        agent.WithModules(
            file.Module{},
            systemd.Module{},
        ),
        agent.WithSQLite("/var/lib/infra/state.db"),
        agent.WithDesiredState(DesiredState),
        agent.WithPlaybooks(Playbooks),
    )
    if err != nil {
        log.Fatal(err)
    }

    if err := a.Execute(context.Background()); err != nil {
        log.Fatal(err)
    }
}
```

`Execute` starts the Cobra CLI.

The exact API may be refined during implementation, but preserve explicit module registration and composable, strongly typed Go resource definitions.

## 3. Declarative Resources

Each resource has a stable identity and a desired state.

Implement:

- Observation of actual state.
- Planning changes.
- Applying changes.
- Verification.
- Ownership tracking.
- Automatic pruning.

Supported operations:

- Create
- Update
- Delete
- No-op

Pruning is enabled by default for previously managed resources.

Never delete unrelated host resources.

If desired-state evaluation or validation fails, abort without pruning.

Initially implement two modules:

### File module

Manage file content, permissions, and deletion.

Support idempotent writes, atomic replacement, and verification.

### systemd module

Manage service enabled/running state.

Support start, stop, enable, disable, and status inspection.

Abstract all OS interactions behind mockable interfaces.

## 4. Resource Transitions

Resources may optionally define a playbook for complex state transitions.

Example:

```go
&postgres.Cluster{
    Name:    "main",
    Version: "18",
    Upgrade: playbook.Ref("postgres-upgrade"),
}
```

When the observed version differs from the desired version:

1. Calculate the transition.
2. Resolve the optional playbook.
3. Persist an execution instance.
4. Lock affected resources.
5. Execute the DAG.
6. Verify the resulting resource state.
7. Commit the applied state to SQLite.

If no playbook exists, use the resource's default safe transition implementation.

Never automatically perform an unsafe transition.

For the initial vertical slice, implement the generic transition mechanism and demonstrate it using a simple test resource. PostgreSQL support is not required yet.

## 5. DAG Playbook Engine

Playbooks are directed acyclic graphs of imperative steps.

Each step must have:

- Stable ID
- Dependencies
- Observe/check behavior
- Apply behavior
- Verification
- Recovery behavior
- Timeout and retry policy
- Optional resource locks

The scheduler must:

- Reject cyclic graphs.
- Run independent nodes concurrently.
- Respect dependencies.
- Persist execution progress.
- Prevent conflicting operations.
- Resume interrupted executions.
- Block dependent nodes after failures.
- Support explicit recovery branches.

Example:

```go
playbook.New("upgrade",
    backupStep,
    validationStep,
    upgradeStep,
    verificationStep,
)
```

Dependencies must be explicit; list order must not implicitly define execution order.

## 6. Idempotency and Crash Recovery

Use the lifecycle:

`Observe → Decide → Apply → Verify → Persist`

Before external side effects, persist the execution attempt.

After a crash, observe the actual state before deciding whether to retry.

Do not blindly repeat an interrupted operation.

Distinguish:

- Satisfied
- NeedsApply
- InProgress
- Failed
- Unknown

Steps must either be safely repeatable or provide explicit recovery logic.

Persist stable execution IDs and node IDs.

Completed steps may be skipped on recovery only when their completion remains valid.

Test interruption at the boundary between successful external execution and SQLite persistence.

## 7. SQLite Persistence

Use **`modernc.org/sqlite`**, without CGO.

Use `database/sql` and versioned migrations.

Persist:

- Managed resource ownership
- Last successfully applied state
- Git/configuration revision
- DAG execution instances
- Node states and attempts
- Operation history
- Locks and recovery metadata

Enable foreign keys and configure WAL/busy timeout appropriately.

Distinguish observed, desired, and managed state.

Do not treat persisted state as proof of actual host state.

Use SQLite transactions for state consistency, not as a substitute for external operation verification.

## 8. Configuration and Secrets

Support two configuration sources:

- Plain YAML configuration.
- SOPS-encrypted YAML secrets.

Load both into user-defined typed Go structures.

Do not create a second infrastructure DSL in YAML.

Configuration values may change without rebuilding the binary. Changes to Go-defined resource topology or behavior require a new build.

Use SOPS with host-specific keys.

Never log decrypted secrets or persist them in SQLite.

Validate configuration before reconciliation.

Support both local configuration files and Git-backed configuration. Implement local loading first, then add Git retrieval after the core vertical slice; Git support is required for the completed implementation.

Allow users to configure the repository URL, revision or branch, configuration paths, and SSH authentication. Support public HTTPS and explicitly opted-in unencrypted HTTP as well. Keep Git retrieval independent of any specific Git server implementation.

Fetch configuration at a resolved commit, validate it before reconciliation, and persist the successfully applied Git revision in SQLite only after successful reconciliation. Retrieval, decryption, or validation failures must abort without pruning or advancing the applied revision. Never log or persist authentication secrets.

## 9. CLI

Use **`github.com/spf13/cobra`**.

Implement:

```text
agent run
agent plan
agent apply
agent status
agent modules
agent playbook list
agent playbook run <name>
agent operations list
agent operations inspect <id>
agent version
```

Commands must have useful descriptions, flags, error messages, and help output.

Support machine-readable JSON output for plans and operation status.

Use appropriate exit codes for validation and execution failures.

## 10. Errors

Use **`github.com/samber/oops`**.

All errors crossing module boundaries must preserve their causes and carry useful context:

- Module
- Resource ID
- Operation ID
- DAG node ID
- Error category

Use stable machine-readable error codes.

Never include secrets in error metadata.

## 11. Unit Testing

Use **`github.com/stretchr/testify`**:

- `require`
- `assert`
- `mock`

Keep unit tests alongside production code, with a corresponding `_test.go` file for each file containing testable logic.

Mock all external interactions:

- Filesystem
- Processes
- Network
- Time
- Git
- SQLite repository interfaces
- External commands

Tests must not modify the developer's host.

Cover normal behavior, validation, errors, idempotency, pruning, concurrency, and recovery.

## 12. End-to-End Testing

Use:

- **Ginkgo v2**
- **Gomega**
- **Testcontainers for Go**

Build a real agent binary and run it inside a Linux container.

Tests must validate observable system effects, not only internal return values.

Required scenarios:

1. Create a managed file.
2. Update the file.
3. Reconcile again and verify no changes.
4. Remove the resource from desired state and verify pruning.
5. Execute a DAG with parallel branches.
6. Interrupt execution and resume from SQLite state.
7. Trigger a transition playbook from a desired-state change.
8. Verify that a failed transition does not advance the applied state.
9. Verify that unrelated resources remain untouched.

Use isolated test containers and temporary volumes.

Do not require privileged host access for ordinary tests.

### Git server integration

Use a small project-owned image containing Git's `git http-backend`, an HTTP CGI server, and OpenSSH for a required, separate integration suite, using Ginkgo v2, Gomega, and Testcontainers for Go.

Image source: `tests/integration/gitserver/`. Install standard Debian packages in the image; do not copy host binaries or substitute fake Git implementations. Build the image automatically with Testcontainers.

Run the flow over both smart HTTP and authenticated SSH. Unencrypted HTTP retrieval must require an explicit opt-in; SSH host and client key verification remain required.

Test the full flow:

1. Start an isolated Git test server container with temporary storage and a test SSH key.
2. Create a test repository over SSH.
3. Push configuration using real Git over the transport under test.
4. Start the agent with Git-backed configuration, using verified SSH keys or explicit HTTP opt-in.
5. Verify initial reconciliation through observable system effects and the persisted applied commit.
6. Push an updated configuration and run reconciliation again.
7. Verify the updated system effects and the newly persisted applied commit.
8. Reconcile the same revision again and verify idempotency.
9. Restart the agent and confirm state and revision persistence.
10. Verify that retrieval, decryption, validation, or reconciliation failures do not advance the applied revision; failures before reconciliation must not trigger pruning.

Clean up test repositories, containers, keys, and volumes after the suite.

Keep Git server integration tests separate from the faster core E2E suite.

If Git server image setup or compatibility blocks progress, report the issue and continue independent core work. Keep the Git integration requirement outstanding until the real suite passes; do not substitute fake Git integration tests or silently skip the suite.

## 13. Repository Structure

Suggested structure:

```text
.
├── agent/
├── core/
│   ├── resource/
│   ├── reconcile/
│   ├── registry/
│   ├── dag/
│   ├── playbook/
│   └── state/
├── config/
├── cli/
├── modules/
│   ├── file/
│   └── systemd/
├── examples/
│   └── basic-agent/
├── tests/
│   ├── e2e/
│   └── integration/
├── go.mod
├── README.md
└── Makefile
```

Use one Go module initially. Third-party infrastructure modules must be able to live in their own repositories and import the public framework interfaces.

Avoid leaking internal implementation packages into public module contracts.

## 14. Development Order

### Milestone 1 — Foundations

Implement module registry, resource interfaces, typed configuration, SQLite schema, and basic Cobra CLI.

Write unit tests immediately.

### Milestone 2 — Reconciliation

Implement the file module, ownership tracking, planning, applying, verification, and pruning.

Prove that repeated reconciliation is idempotent.

### Milestone 3 — DAG Engine

Implement graph validation, parallel scheduling, persisted checkpoints, locks, and recovery.

Prove crash recovery with deterministic tests.

### Milestone 4 — Transitions

Connect resource differences to optional playbook execution.

Verify applied state advances only after successful transition verification.

### Milestone 5 — E2E

Build the example agent, run it in Testcontainers, and verify actual filesystem changes and DAG execution.

Add systemd integration only where the container environment supports it reliably.

### Milestone 6 — Git and Secrets

Implement YAML/SOPS loading and Git configuration retrieval.

Complete the required Git HTTP/SSH integration suite, including authenticated retrieval, configuration updates, idempotency, persisted revisions, restart behavior, and failure handling.

## 15. Definition of Done

The implementation is complete when:

- The example agent compiles and runs.
- External modules can register without modifying core.
- SQLite state survives process restarts.
- Declarative resources reconcile and prune correctly.
- DAG execution supports concurrency and recovery.
- Resource changes can trigger optional playbooks.
- Unit tests pass.
- Container-based E2E tests pass.
- Git-backed configuration works with SSH authentication and tracks successfully applied revisions in SQLite.
- The separate Git HTTP/SSH integration suite passes against a real server.
- The CLI exposes the implemented functionality.
- The README documents installation, usage, configuration, and tests.

Do not claim features are implemented unless they work and are tested.

Do not replace incomplete functionality with no-op stubs that report success.

If time is limited, prioritize a fully working vertical slice over broad but incomplete feature coverage.

## 16. Final Report

At the end, provide:

- Implemented features.
- Architecture and key design decisions.
- Commands to build and run the agent.
- Unit and E2E test results.
- Any failing or skipped tests.
- Known limitations.
- Next recommended implementation milestone.

Keep the report concise and factual.
