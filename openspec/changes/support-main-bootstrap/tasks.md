# Tasks

## 1. Consumer development initialization

- [x] 1.1 Accept canonical resolved v0 pseudo-versions in initializer metadata detection and consumer scaffold validation; verify unit tests cover pseudo-version acceptance, exact-tag compatibility, dev/branch/invalid-version rejection, and unchanged module-scaffold release validation.
- [x] 1.2 Preserve identical immutable dependency, skill, template, and manifest versions for development consumers; verify rendered asset tests and generated-consumer integration build, plan, apply, and repeated no-op behavior.
- [x] 1.3 Update README, consumer authoring guidance, and curated version-pinning knowledge for development snapshots; verify source references and artifact checks without claiming public acceptance.

## 2. Selected-version guide commands

- [x] 2.1 Render main for next and the selected tag for release bootstrap commands; verify generated HTML tests for next, stable, and RC versions, including absence of a hard-coded future tag in next.
- [x] 2.2 Update version-guide documentation to explain development source resolution and immutable consumer pins; verify docs builds, internal links, and artifact checks.

## 3. Integration and acceptance

- [x] 3.1 Run task check, task test:race, task test:e2e, and task test:integration; record actual results and any environment limitations.
- [ ] 3.2 Once implementation is available on public main, execute the documented go run @main bootstrap without local replacements and verify dependency/asset version agreement, compiled consumer file creation, and repeated no-op apply. Keep this task open until actually exercised.
