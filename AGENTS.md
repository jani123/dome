# Repository rules

These rules apply throughout the repository. Use `README.md` for package layout,
public API examples, supported behavior, and development commands.

## Documentation and working notes

- Never commit ad hoc plans, handovers, task checklists, or agent progress/session
  notes, regardless of their filename or location. Keep temporary working
  documents in ignored local storage such as `.build/notes/`.
- Structured OpenSpec artifacts under `openspec/` are the explicit exception:
  capability specs and change proposals, designs, tasks, and archives belong in
  version control. This does not permit standalone plan or handover documents.
- Curated OKF project knowledge is maintained documentation: preserve lasting
  facts, rationale, and source references, not chat transcripts or session notes.
- Move lasting decisions, supported behavior, setup instructions, and operational
  guidance into the relevant README or documentation. Integrate that material for
  its readers; do not preserve a working plan by renaming it as documentation.
- Before changing behavior, read relevant entries from `knowledge/index.md` and
  `openspec/specs/`, then follow their code/test sources. Apply the OpenSpec scope rule below.
- Update affected specs, curated knowledge, and user documentation together.
  Preserve source/verification provenance; do not turn historical test results or
  local builds into current or public acceptance claims. Run `task check:artifacts`
  when these maintained artifacts change; setup is documented in `docs/development.md`.
- Before committing, inspect staged files for working documents. Ignore patterns
  are a convenience, not an exception to this rule.

## When to use OpenSpec

- Use an OpenSpec change for new features, architectural changes, public API/CLI
  or configuration contract changes, migrations or breaking upgrades, and changes
  to reconciliation, recovery, persisted-state or security guarantees. Use it when
  significant cross-cutting work needs design decisions or coordinated acceptance,
  or when the user explicitly requests it.
- Routine dependency/security patches, compatible toolchain updates, bug fixes
  restoring specified behavior, internal refactors, and documentation, test or CI
  maintenance do not require a new proposal unless they meet a criterion above.
  Judge the actual impact, not merely the change category or number of files.
- For routine work, implement directly, run relevant checks, and update affected
  documentation and existing specs only when their content changes. Do not create
  proposals, tasks or new requirements solely to record maintenance work.
- Bug fixes require regression coverage at the appropriate level. If a bug
  exposes a missing, ambiguous or incorrect feature requirement, update the
  affected OpenSpec requirement/scenario as well as executable tests. Restoring
  an already specified behavior does not require a new proposal; changing the
  intended contract follows the OpenSpec criteria above.
- For qualifying work, record requirements, design, tasks and verification in
  OpenSpec before implementation. If routine work reveals a qualifying change,
  plan that expanded scope before implementing it. Keep ordinary notes local.

## Architecture and configuration

- Dome is a Go library (`github.com/jani123/dome`) that users compile into their
  own agents. Register modules explicitly; keep core independent of concrete
  modules and public contracts usable by external Go modules.
- Do not introduce global registration through `init()`, runtime plugins,
  Yaegi, or another infrastructure DSL.
- Define resource topology and behavior in Go. Load configuration values into
  user-defined typed structures from YAML, and use SOPS for encrypted secrets.
- Keep changes focused, preserve unrelated work, and make reasonable API
  refinements when needed. Update the README when public APIs, CLI behavior,
  supported configuration, or development setup changes.

## Reconciliation, recovery, and state

- Follow `Observe → Decide → Apply → Verify → Persist`. Persist execution
  attempts before external side effects, and record applied state only after
  successful verification.
- Prune only previously managed resources. Configuration retrieval, decryption,
  evaluation, or validation failures must prevent reconciliation and pruning.
- Re-observe interrupted work before deciding whether to retry. Steps must be
  safely repeatable or provide explicit recovery logic. Re-observe and verify
  completed steps before skipping them on resume.
- Resume unfinished transition DAGs even when the resource already matches its
  target. Do not bypass pending recovery by changing or removing that target.
- Preserve dependency ordering, resource locking, and failure blocking. Failed
  operations must not advance their applied resource state, and failed
  reconciliation must not advance the overall applied Git revision.
- Use `database/sql`, versioned migrations, and `modernc.org/sqlite` without
  CGO. Keep observed, desired, and managed state distinct; persisted state is
  not proof of actual host state.
- Never put secrets in logs, errors, identifiers, snapshots, or persisted
  execution metadata. Persist non-secret representations such as content
  hashes instead of secret contents.
- Use `github.com/samber/oops` at error boundaries. Preserve causes and specific
  error categories, and attach relevant module, resource, execution, and DAG
  node identifiers without exposing secrets.

## Testing and infrastructure

- Keep external boundaries injectable. Unit tests mock filesystem, processes,
  network, Git, time, persistence interfaces, and external commands. Use
  Testify assertions and mocks, with tests alongside testable production code.
- Tag tests that exercise real filesystem, SQLite, processes, host keys, or
  network boundaries with `integration`. Use isolated temporary storage and
  clean up resources; do not mutate unrelated host resources.
- Use Ginkgo, Gomega, and Testcontainers for container E2E and Git integration
  tests. Exercise the compiled agent and verify observable effects, including
  idempotency, pruning, concurrency, failure handling, and crash recovery.
- Use the project-owned Git HTTP/SSH image in `tests/integration/gitserver/`.
  Install standard packages rather than injecting host binaries or copying
  libraries. Keep SSH authentication and host verification coverage, and
  require explicit opt-in for unencrypted HTTP retrieval.
- Diagnose network failures using their actual endpoints and errors. Fix proxy
  configuration rather than introducing alternate downloads, replacement
  runtimes, or automatic fallback paths to bypass blocked access.
- Do not substitute stubs for real integration dependencies, silently skip
  failing tests, or report an unverified feature as complete.

## Verification

Run the checks relevant to the change. Changes affecting shared runtime behavior
or completion of a substantial feature require the full suite:

- `task check`: formatting, vet, and unit tests.
- `task test:race`: unit tests with the race detector.
- `task test:e2e`: the compiled agent in Linux containers.
- `task test:integration`: real boundary tests and Git HTTP/SSH integration.

Report what changed, what was verified, and any failing or skipped checks or
remaining limitations. Keep implementation details and commands in the README
rather than duplicating it here.
