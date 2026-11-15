# Tasks

## 1. Runtime profile and workspace storage

- [x] 1.1 Implement versioned strict runtime-profile YAML and workspace-relative path validation, with source kinds, state path and non-secret credential references; verify mocked unit cases for unknown fields, unsupported versions, invalid combinations, secret rejection, absolute/local paths and repository-relative blob paths.
- [x] 1.2 Add injectable workspace loading and safe atomic no-replace profile publication with restricted permissions and equivalent-repeat handling; verify unit fault injection and integration-tagged real permission, symlink, cancellation, concurrent-writer and preserved-state cases.
- [x] 1.3 Document profile schema, .dome/agent.yaml discovery, --workspace and --agent-config semantics in configuration/deployment guides and source-linked configuration knowledge; validate links and schema examples with task check:artifacts and profile parsing tests.

## 2. Host initialization command

- [x] 2.1 Add agent init using flags/profile input and injectable terminal/prompt interfaces without requiring new Backend methods; verify missing-only SSH/local prompts, complete unattended input, defaults, explicit HTTP opt-in, optional encrypted-secret references and non-interactive/JSON/EOF/cancellation failures using unit tests.
- [x] 2.2 Wire initialization into agent.Command while preserving cli.New compatibility and existing Go-configured consumers; verify no backend reconciliation/state calls during init and no prompts outside init with captured mock calls and CLI tests.
- [x] 2.3 Keep Git-backed setup settings-only and local setup pointed at existing resource configuration; verify no resource example, repository, key or SQLite database is created and no existing configuration changes in real filesystem integration tests. Update CLI help and README with managed-host interactive and unattended setup examples.

- [x] 2.4 Move advertised source/credential/secrets/state flags to init and retain hidden legacy runtime registrations, preserving accepted flag positions; verify exact visible/absent flag names for root/init/plan/apply/run/status/playbook/operations help, inherited/global sections, repeated command-tree use and retained legacy overrides. Update normal README/generated consumer/guide examples to show init once followed by workspace-based runtime commands, with overrides confined to compatibility guidance.

## 3. Runtime defaults and precedence

- [x] 3.1 Load saved settings only for runtime source/state commands and resolve explicit flags over profiles over Go defaults, preserving explicit false/empty flag presence and coherent source switches; verify Backend Options for plan/apply/run/status/playbook run/resume and operation diagnostics, plus help/version/list availability with invalid profiles.
- [x] 3.2 Resolve workspace paths without changing working directory or consumer-defined resource identities; verify identical state/config resolution across invocation directories, legacy no-profile behavior, overrides leaving profiles unchanged, and invalid profiles blocking before backend execution.
- [x] 3.3 Update generated consumer docs/source examples and core example guidance for init then apply, preserving existing sandbox identity and local configuration; verify render/unit and generated-consumer integration smoke tests. Update consumer-operations knowledge and README/configuration/deployment guides with source precedence, supervisor use and existing-state migration guidance.

## 4. Integrated acceptance

- [x] 4.1 Extend compiled-agent container E2E coverage for init then repeated apply, idempotency, diagnostics, workspace selection and failed setup preserving existing execution history; run task test:e2e and record actual results.
- [x] 4.2 Extend project-owned Git SSH/HTTP integration scenarios to initialize once and retrieve committed configuration/secrets without repeated source flags, including known-host rejection, HTTP opt-in and retrieval failure preserving revision/state; run task test:integration without substitute dependencies or silent skips.
- [x] 4.3 Run task check, task test:race, task check:artifacts and relevant static/workflow checks; compare implementation against all new scenarios and record actual checks and remaining limitations in verification.md before marking completion.
