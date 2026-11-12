# Dome onboarding and release plan

Status: implementation in progress, 2026-11-12. Based on `HANDOVER.md` and
current repository inspection. This is a scoped delivery plan, not a replacement
for `AGENTS.md` or a restoration of the deleted framework implementation plan.
Local implementation and verification are underway; public deployment/release acceptance is outstanding.

## Outcome and scope

A consumer runs one bootstrap command, opens Codex in the generated project, and
describes the machine they want to manage. Codex authors Go resources and
playbooks, builds and tests the project, and inspects the compiled agent's JSON
plan. The consumer requests application explicitly. The compiled agent then runs
without AI or access to a coding-agent service.

Deliver the initializer, clean starter, consumer skill, API examples, public
guides, GitHub Actions, and a repeatable release process. Preserve the current
Go-library architecture, typed YAML/SOPS configuration, explicit registration,
resource ownership, and reconciliation/recovery contracts. Limit runtime changes
to specific API gaps found while building a real external consumer.

Defer MCP, other coding-agent adapters, Windows/macOS support, package-manager
repositories, automatic upgrades, and a general infrastructure DSL. End users
build and distribute their own agent binaries; Dome releases do not distribute
a universal host-management agent.

## Verified baseline and remaining discovery

Initial inspection before implementation confirmed:

- Module `github.com/jani123/dome` declares Go `1.26.0`.
- `make build` compiles the demonstration agent in `examples/basic-agent`.
- Existing checks are `make check`, `make test-race`, `make test-e2e`, and
  `make test-integration`.
- No local release tags, `.github` workflows, license, or changelog were found.
- Origin points to `git@github.com:jani123/dome.git`.

Remote inspection over HTTPS on 2026-11-12 confirms the repository is public,
its default branch is `main`, and no remote tags or public releases exist.
The owner enabled Pages, and the repository API now confirms `has_pages: true`.
The API reports no detected license; `https://jani123.github.io/dome/` still
returns HTTP 404 pending the documentation build and deployment. GitHub documentation is now
reachable through the devcontainer proxy. Actions settings, branch protection,
and runner availability still need inspection in phase 0.

Use `https://github.com/jani123/dome.git` for Git access in this devcontainer.
The configured origin still uses SSH, which bypasses the HTTP proxy and fails
under the container's blocked external DNS/direct-network policy. HTTPS Git and
the GitHub API were verified successfully; the remote configuration was preserved.
Previous test results in the handover are historical.

## Proposed delivery decisions

Owner confirmed the first release sequence (`v0.1.0-rc.1` → `v0.1.0`), Linux
amd64/arm64 and Codex support, Hugo with GitHub Actions, and one shared version
for the library, initializer, templates, skill, and docs on 2026-11-12.
The owner agreed to GPLv3 for customer source access; GPL-3.0-only without a
linking exception is now adopted. Other choices below remain implementation defaults.

Owner's licensing requirement: paid implementation/services are welcome, but a
customer receiving a compiled Dome-based agent must also be able to obtain its
complete corresponding source and build material, so they can modify and rebuild
their infrastructure agent without depending on its original supplier. GPLv3
without a linking exception implements this goal; dependency review and
redistribution notice generation are recorded in `docs/DEPENDENCY_REVIEW.md`. Document customer-owned
configuration, repository access, credentials, and SOPS keys separately: source
licensing alone does not ensure operational independence. Standard GPLv3 permits
multiple source-delivery mechanisms, so distinguish its legal requirements from
the recommended handover of source alongside every binary.
Make this goal explicit in the public guides and generated consumer README.
Explain that GPLv3 generally covers a distributed combined agent linked with
Dome, while merely managed applications are not covered just by being managed.
Do not describe this as a library-only source obligation.

| Area | Proposed choice |
| --- | --- |
| First release | Rehearse with `v0.1.0-rc.1`, then publish `v0.1.0` after acceptance. Confirm these versions are unused. |
| Compatibility | Patch releases preserve exported APIs, CLI/JSON contracts, and supported state. Breaking v0 changes require a minor version, migration notes, and explicit consumer action. Document that v0 remains experimental. |
| Cadence | Release when a tested milestone is ready; patches as needed. No calendar promise. Only the latest minor receives fixes initially. |
| Version scope | One repository tag versions the library, initializer, embedded templates, skill, examples, and documentation source. |
| Initial platforms | Linux amd64 and arm64 initializer binaries and consumer agents. Run native smoke tests on both before claiming support; cross-compilation alone is insufficient. |
| Initial coding agent | Codex, with `.agents/skills/dome/SKILL.md` and consumer `AGENTS.md`. Other agents can use the Go library manually. |
| CLI layout | `cmd/dome` is the initializer; the consumer compiles its own `agent`. Keep initializer commands separate from the existing consumer CLI. |
| Documentation | Markdown under `docs/`, built with pinned Hugo and a minimal repository-owned layout; GitHub Pages hosts guides and links to pkg.go.dev. No duplicated API signatures. |
| Release tooling | Go build/archive tooling plus GitHub Actions and GitHub CLI; avoid adding a release framework initially. Maintainers curate release notes from merged PRs. |
| Release trigger | Reviewed release-preparation PR, then a maintainer-created `v*` tag triggers verification and draft artifacts. Publish only after external-consumer checks pass. |
| Licensing | GPL-3.0-only, without a linking exception. Include complete corresponding source/build material and upstream notices in releases. |

Confirmed choices are recorded above; remaining choices are implementation defaults.
Any platform unsupported by actual runners/tests stays explicitly unreleased.

## Phase 0 — Establish release prerequisites

Deliverables:

- Verify GitHub visibility, default branch, tags/releases, Pages availability,
  Actions permissions, and available Linux arm64 runners.
- Obtain the owner's license choice, add the approved license, and check required
  dependency notices for distributed binaries.
- Add a concise compatibility/support policy and changelog. Document minimum Go,
  supported OS/architectures, state migration expectations, and pre-release use.
- Record the first release version and identify the maintainer responsible for
  publishing and repository settings.

Acceptance:

- Remote facts are recorded from actual inspection, and required settings are
  configured or listed as concrete blockers.
- License is owner-approved. No public release is marked ready while it is absent.
- Candidate version is unused; no existing tag or published version is replaced.

## Phase 1 — Make the public API usable from another module

Depends on phase 0's version policy; documentation work can begin before licensing
is resolved.

Deliverables:

- Audit exported package/type/function comments across composition, CLI,
  configuration, resource contracts, registry, reconciliation, state, DAG,
  playbook, errors, and built-in modules.
- Add testable `Example...` functions for composition, typed configuration,
  explicit module registration, file resources, playbooks, and transition recovery.
  Use mocked boundaries for ordinary examples; tag real-boundary tests integration.
- Add a separate-module consumer fixture that imports public packages only.
  Keep `examples/basic-agent` as the advanced demonstration.
- Fix only demonstrated public API obstacles and document any refinements.

Acceptance:

- The external fixture builds/tests without importing `internal` packages.
- Examples compile and deterministic examples run under `go test`.
- Docs explain safe repeatability/recovery, hashes instead of secrets in state,
  managed-only pruning, and the effect of partial reconciliation failures.
- Run all four repository checks if shared runtime behavior changes.

## Phase 2 — Build the starter and initializer

Depends on phase 1. Proposed bootstrap, once a release exists:

```sh
go run github.com/jani123/dome/cmd/dome@v0.1.0 init my-machine --module example.com/my-machine
```

The release version here is illustrative until published. Go, Codex, network
access for module resolution, and access to the target Linux machine remain
prerequisites; bootstrap does not install them.

Deliverables:

- Implement `cmd/dome` with `init <directory> --module <module-path>` and `version`.
  Keep generation logic and embedded assets under `internal/scaffold`.
- Generate `go.mod`, `go.sum`, `cmd/agent/main.go`, typed settings, configuration,
  consumer README, `.gitignore`, consumer `AGENTS.md`, the Dome skill, and a
  non-secret scaffold manifest identifying release/template/skill versions.
- Default to a file resource inside a project-local sandbox, local SQLite state,
  and explicit permissions. No privileged paths, demo failure hooks, secrets,
  or automatic apply. Generated instructions include plan/apply/status commands.
- Use the initializer's exact release version as the dependency/asset version.
  Reject development builds without an explicit test-only version injection;
  do not silently fall back to `main`, `latest`, or an arbitrary remote template.
- Validate module names and destinations before writing. Support fresh directories;
  for an existing project, generate a separate consumer subdirectory, preserving
  its module and instructions. Defer merging into an existing module.
- Refuse collisions and symlink/path escapes. Use staged generation; on failure,
  clean up only files created by that invocation. Report dependency-resolution
  failures accurately. Bootstrap may run `go mod tidy`, but never run the agent.

Acceptance:

- Inject filesystem/process boundaries in unit tests. Cover invalid input,
  collisions, failure cleanup, version metadata, and preservation of existing files.
- Integration tests generate a real external module, resolve dependencies, build,
  and inspect `agent plan --json` without a local source replacement in the final
  release test. Before tagging, use a temporary local replacement explicitly.
- Container tests verify initial apply, no-op apply, content change, removal/pruning,
  unrelated-file preservation, and invalid-config failure without mutations.
- Native Linux amd64/arm64 bootstrap/build smoke tests pass before both binaries ship.

## Phase 3 — Package the consumer skill and update workflow

Depends on phases 1–2; ship the skill inside the same embedded asset set.

Deliverables:

- Author a consumer skill with valid name/description front matter and focused
  references. Install at `.agents/skills/dome/SKILL.md`, the project-local location
  documented by [official Codex customization guidance](https://learn.chatgpt.com/docs/customization/overview).
- Teach it to inspect `go.mod`, the scaffold manifest, installed source (`go doc`),
  and version-matched examples before generating Go. Do not use current-main APIs
  for a consumer pinned to an older release.
- Include workflows for resource and playbook authoring, validation, JSON plan
  inspection, user-requested apply, operation inspection, and safe crash recovery.
  Preserve secrets, locking, dependency/failure behavior, and pending transitions.
- Keep consumer instructions distinct from this repository's maintainer rules.
- Document an explicit update workflow: install/run the target initializer version,
  generate a temporary project, review dependency and skill/reference diffs, update
  the dependency plus metadata together, then test and plan before deployment.
  Preserve user-authored Go and instructions; no unattended file replacement.

Acceptance:

- Validate packaging/front matter and referenced files automatically.
- Test two fixture versions: a mismatch must produce a clear corrective workflow,
  rather than silently mixing dependencies and skill assets.
- Conduct a real Codex smoke session in a generated project: discover the skill,
  add a file resource and playbook, build/test, inspect JSON plan, and apply only
  when requested. Record Codex version and observed results; do not claim that
  static packaging checks establish coding-agent discovery or behavior.

## Phase 4 — Publish guides with tested examples

Depends on phases 1–3 for final content; site scaffolding can proceed earlier.

Deliverables:

- Guides: prerequisites/bootstrap, conversational onboarding, Go resource and
  playbook authoring, typed YAML, SOPS, Git/SSH verification, Linux deployment,
  state ownership, failure inspection/resume, upgrades, and release/support policy.
- Deployment guide distinguishes the development machine from the target host,
  shows consumer compilation/distribution and supervision, and explains paths,
  permissions, state backups, and secret/key provisioning.
- Derive runnable snippets from tested source fixtures where possible; validate
  remaining commands with generated-project smoke tests. Link exported APIs to
  version-qualified pkg.go.dev references.
- Publish immutable guide snapshots at `/v0.1.0/` and subsequent release paths,
  a `/latest/` stable alias, and clearly labeled `/next/` development guides.
  Prereleases never replace the stable alias. Keep old snapshots when deploying.
- Configure the GitHub project-site base path; build/check links on PRs. Deploy
  only trusted commits through the `github-pages` environment using the official
  [Pages Actions workflow](https://docs.github.com/en/pages/getting-started-with-github-pages/using-custom-workflows-with-github-pages).

Acceptance:

- Site builds with pinned tooling; internal links, asset paths, and version
  navigation pass checks. Verify the deployed URL after enabling Pages.
- A clean Linux consumer completes the published quickstart without demo hooks.
- Older guide snapshots remain intact; stable navigation never points to an RC.
- README points users to quickstart, API reference, and authoritative development
  commands. There is one maintained source for each guide/example.

### Owner instructions: enable GitHub Pages

Status: the owner reports Pages enabled, confirmed by `has_pages: true` in the
repository API on 2026-11-12. Publishing-source selection is owner-reported;
deployment permissions and environment rules still need verification when the
workflow is implemented. The public site currently returns HTTP 404.

The following steps are retained for setup reference. The site will become available after
phases 4–5 add the documentation build and deployment workflow; selecting the
source alone does not publish the current repository.

1. Sign in with repository admin or maintainer permissions and open
   [Dome Pages settings](https://github.com/jani123/dome/settings/pages).
2. Under **Build and deployment → Source**, select **GitHub Actions**.
   Use the planned custom workflow; there is no need to create a workflow from
   GitHub's suggested templates or select a publishing branch/folder.
3. In **Settings → Actions → General**, ensure repository policy permits Actions
   and the official actions used by the documentation workflow. Keep default
   workflow permissions read-only; the deployment job will request its specific
   permissions in YAML.
4. Once `.github/workflows/docs.yml` is implemented and merged, check
   **Settings → Environments → github-pages**. Ensure its deployment rules allow
   `main` for development guides and the release tags used for versioned guides.
   If required reviewers are configured, approve each intended deployment there.
5. Open [Dome Actions](https://github.com/jani123/dome/actions), select the
   documentation workflow, and use **Run workflow** on `main` if a manual run is
   needed. The implementation must include `workflow_dispatch` for this option.
6. After a successful build/deploy, open `https://jani123.github.io/dome/` and
   follow its guide links. Confirm CSS, assets, and version navigation work under
   `/dome/`. The Hugo base URL must be `https://jani123.github.io/dome/`.

Implementation requirements: build the static site, upload the generated output
with `actions/upload-pages-artifact`, then deploy with `actions/deploy-pages` in
the `github-pages` environment. The deploy job must depend on the successful
build and have `pages: write` and `id-token: write`. Pin action revisions and
keep PR builds separate from trusted deployment. No personal access token or
custom domain is needed for this setup.

If deployment fails, inspect the workflow's build/deploy logs and environment
rules. If the site returns 404 before the first successful deployment, complete
the workflow implementation and deployment first. If assets return 404 after
deployment, check the `/dome/` base path and uploaded output directory.

References: [configure the publishing source](https://docs.github.com/en/pages/getting-started-with-github-pages/configuring-a-publishing-source-for-your-github-pages-site)
and [custom Pages workflows](https://docs.github.com/en/pages/getting-started-with-github-pages/using-custom-workflows-with-github-pages).

## Phase 5 — Enforce CI and rehearse releases

Start baseline CI after phase 0; complete release automation after phases 1–4.

Deliverables:

- `.github/workflows/ci.yml`: separate required jobs for `make check`,
  `make test-race`, `make test-e2e`, and `make test-integration`, plus scaffold,
  external-consumer, skill, and documentation validation. Do not use path filters
  that leave required checks permanently pending.
- Use Go from `go.mod`, pinned formatter/site tooling, explicit timeouts, and
  dependency caches. Run container suites on Linux with real Docker and the
  project-owned Git HTTP/SSH image, including authentication/host verification.
- `.github/workflows/docs.yml`: build PR documentation without deploying it;
  deploy trusted development/release snapshots with serialized updates.
- `.github/workflows/release.yml`: validate tag/version consistency, run the full
  suite against the tagged commit, build `CGO_ENABLED=0` initializer archives,
  record version/commit, and create a draft release with checksums and notes.
- Artifacts: Linux amd64/arm64 initializer archives, SHA-256 checksums, license/
  notices, and release notes. The library is consumed through its Go module tag;
  embedded skill/templates come from that same commit. No demo agent binary.
- Use read-only default token permissions, job-specific publication permissions,
  full commit SHA action pins, and no secrets or write tokens for untrusted PR
  execution, following [GitHub's secure-use guidance](https://docs.github.com/en/actions/reference/security/secure-use).

Acceptance:

- PR failures block merging, and no test suite is silently skipped or substituted.
- Rehearse `v0.1.0-rc.1`: download/checksum/extract/run each supported initializer,
  generate a project, resolve the exact tagged module without a local replacement,
  and execute the acceptance flow in an isolated environment.
- Public module resolution and pkg.go.dev documentation are checked after tagging;
  verify license recognition and published examples. A tag itself makes module
  source available independently of GitHub Release publication.
- Only publish the draft release after these checks pass. Then deploy its docs;
  promote `/latest/` only for a stable release. Report partial publication failures
  and retry the failed publication step against the same immutable artifacts.

## Phase 6 — Ship and validate the stable consumer experience

Depends on all earlier acceptance checks and resolution of licensing/access.

- Prepare the stable release PR with curated changes, compatibility/migration notes,
  supported platforms, and full verification results for its final commit.
- Tag `v0.1.0`, run release verification, and repeat public resolution/bootstrap
  checks for the stable version before publishing the release and stable guides.
- Exercise the complete path: bootstrap → Codex discovery → Go authoring →
  build/test → JSON plan → requested apply → idempotency → update → recovery.
- Record artifact URLs/checksums, Pages URL, pkg.go.dev availability, and observed
  results. Leave no platform or onboarding claim based solely on cross-compilation.
- Never move a published tag. Fix a bad version with a new release and documented
  migration; retract unusable module versions when appropriate. Back up state
  before migration; do not promise an older binary can read a newer database.

Release mechanics follow the [Go module release workflow](https://go.dev/doc/modules/release-workflow).
Support policy must distinguish API compatibility, CLI/JSON compatibility, and
forward state migrations; a binary rollback is not a host-state rollback.

## Completion checklist

- [x] Adopt GPL-3.0-only and review linked dependency licensing.
- [ ] Verify remaining GitHub Actions/environment/required-check settings.
- [x] Documented/tested public APIs and external consumer fixture.
- [x] Safe version-pinned initializer and clean starter (local replacement test).
- [ ] Discoverable, version-matched Codex skill and documented manual updates.
- [ ] Versioned Pages guides and pkg.go.dev reference verified publicly.
- [ ] Required CI runs every repository suite plus onboarding checks.
- [ ] RC rehearsal and stable release pass on each advertised platform.
- [ ] Bootstrap-to-recovery consumer flow observed and recorded.

## Implementation progress

Implemented locally: GPLv3 license and dependency review, public API comments and
executable examples, atomic Linux initializer, external consumer/starter tests,
embedded version-matched Codex skill, Hugo guides/layout, link checking, required
CI workflows, native-platform jobs, draft release packaging with vendored source,
and versioned Pages assembly from published immutable release tags.

Verified locally: all four repository suites passed, including container E2E and
real Git HTTP/SSH integration. The generated-consumer tests use an explicit local
replacement until a public tag exists. Hugo development/site assembly and link
checks passed; actionlint and shell syntax checks passed. Final package rehearsal
and any later verification are recorded below as work progresses.

Outstanding acceptance: native arm64 execution, real Codex discovery/authoring,
public tagged-module bootstrap, pkg.go.dev recognition, GitHub required-check and
environment settings, and first Pages deployment. No remote tag/release was
created, no public deployment is claimed, and these items remain unchecked.
The skill creator's Python validator initially could not run because PyYAML is
absent; Go tests validate the rendered front matter and packaged source instead.

Keep command details and operational instructions in README and the guides;
use this checklist to track delivery and acceptance.
