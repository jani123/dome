# Tasks

## 1. Consumer CLI

- [x] 1.1 Add required --project-module without an identity alias and with independent command flag storage; verify unit tests for the new flag, missing/empty/invalid identities, rejection of the old identity spelling, help, and unchanged resource-module-required cases.
- [x] 1.2 Update README, generated guidance, current specs and curated initialization knowledge; verify artifact validation and existing explicit-module tests.

## 2. Container quickstart

- [x] 2.1 Add the pinned official Docker Go shell, caller ownership, and writable cache environment; use --project-module in next quickstart and --module for older release guides and verify rendered guide checks and shell syntax.
- [x] 2.2 Exercise the actual Docker image and bind mount with local implementation, initializing with --project-module, building/testing, planning, creating the sandbox file, and repeating apply; verify host-visible file content and UID/GID without claiming public main acceptance.

## 2b. External resource modules

- [x] 2.3 Generate imports, Module registrations, and pinned dependencies for repeatable --module selections; verify validation, unique aliases, required project identity, and editable parameter guidance with unit tests.
- [x] 2.4 Preserve requested dependency pins through staged resolution; verify pure pin checks and real publication rollback on version conflicts or resolution failure.
- [x] 2.5 Compile and run a consumer with a real separately generated module via local replacement; verify explicit registration and plan without claiming public package availability, and update README, generated docs, knowledge, and artifact checks.

## 3. Verification

- [x] 3.1 Run task check, task test:race, task test:e2e, task test:integration, docs builds, and task check:artifacts; report actual results and remaining limitations.

## Verification evidence

The Docker smoke used the current compiled initializer with published
`v0.1.0-rc.1` as the generated library dependency in
`golang:1.26.11-bookworm`. It verified creation with `--project-module`, build,
Go tests, plan, file contents, repeated no-op apply, and host UID/GID ownership.
The test environment needed its explicit outbound proxy passed to Docker after
an initial DNS timeout to proxy.golang.org. This is local command verification,
not evidence that the new flag is available on public main.

External-module verification compiled a consumer with a separately generated
resource module using explicit local source replacements. The compiled consumer
listed both file and widget registrations, planned/applied its real sandbox file,
and reported no-op on repeat. Requested-pin conflict and failed-resolution tests
verified no destination or staging output survived. The full check, race, E2E,
and integration suites, docs version/build checks, and artifact validation passed
for this implementation; no public external-package acceptance is implied.
