# Tasks

## 1. Consumer workspace

- [x] 1.1 Scaffold the pinned named-user devcontainer, correct embedded paths, and stable workspace; verify JSON rendering, user/toolchain settings, and generated consumer tests.
- [x] 1.2 Update generated README, main README, current specs and curated knowledge with startup/re-entry/ownership and proxy guidance; verify artifact checks.

## 2. Quickstart walkthrough

- [x] 2.1 Make Docker bootstrap return to the host and add CLI/VS Code tabs before the remaining container commands; verify rendered commands and independent tab click/keyboard/default-selection tests.
- [x] 2.2 Add and run a real Dev Container CLI acceptance command covering bootstrap, configuration/example code, container startup, build/test, plan, file creation, no-op repeat, re-entry, host ownership, and scoped cleanup; record actual source/version and environment limitations.

## 3. Verification

- [x] 3.1 Run task check, task test:race, task test:e2e, task test:integration, docs checks and artifact validation; keep public-main/GUI claims separate from actual local acceptance.

## Verification evidence

Local acceptance used Dev Container CLI 0.82.0 and real Docker. The bootstrap
executed the current source initializer inside `golang:1.26.11-bookworm`, with
`v0.1.0-rc.1` as its published library dependency. This intentional local-source
variation tests the unpushed scaffold; it does not certify public main.

The generated named-user workspace built and tested the exact YAML/Go example
from quickstart, initialized and planned without creating the file, created
`sandbox/hello.txt` with the expected contents, and reported no-op on repeat.
After an actual stop and CLI restart, repeated apply remained no-op. Host files
including the binary, managed file, and SQLite state had caller UID/GID. The
acceptance command removed its own container and retained ignored diagnostics.
This proxy-only environment required an explicit test-local hostname mapping;
the generated source configuration retained environment-variable references.

Task check, race, container E2E, and integration suites passed. Hugo builds,
versioned-guide command checks, both tab-group interaction/keyboard/default tests,
shell syntax, and artifact validation passed. VS Code GUI interaction was not
exercised; its shared Dev Container configuration and rendered opening instructions
were checked without claiming GUI acceptance.
