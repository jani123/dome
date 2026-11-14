# Module initializer verification

## Implemented local scope

The CLI renders a standalone public resource-module library and stages dependency
resolution plus pinned OpenSpec refresh before no-replace publication into an
existing empty directory. Names/paths/versions, file modes, hidden entries,
symlinked ancestors, cancellation, failed preparation and concurrent collisions
have unit or integration coverage. Publication never rolls back exposed paths;
partial output is retained explicitly because deleting a concurrently replaced
inode cannot be made conditional with ordinary Linux unlink.

The generated package separates resource/module/backend source and provides an
explicit unfinished production adapter, mocked tests, a public consumer example,
Task-only tooling, OpenSpec/OKF source provenance, Hugo, source-release CI and
project-adapted Devbox/devcontainer assets. Shared asset drift is checked by
`task check:artifacts` and the maintained sync command.

## Observed verification

An isolated `example.com/module-fixture` repository used an explicit test-only
replacement to this checkout. The real PrepareModule path ran Go resolution,
checksum-installed OpenSpec, refreshed its workflows and published prepared
assets. The resulting repository passed build, unit/race, tagged-suite compilation,
local go doc, knowledge/workflow validation, task setup/check/artifact validation,
Hugo build and actionlint/shell syntax checks. Both Ginkgo/Gomega/Testcontainers
suites built the owned Debian fixture and ran a real probe or compiled consumer,
verifying plan without mutation, contents after apply, no-op repeat, pruning and
unrelated-file preservation. Requested container checks were not skipped.

Core `task check`, `task test:race`, `task test:e2e`, `task test:integration`,
`task check:workflow`, `task check:static`, `task docs:build` and
`task check:artifacts` passed. Default and explicit initializer version metadata
were exercised, and invalid release input was rejected without remote writes.
Native CI is configured to exercise both generated repository families; this is
configuration evidence, not a claim those hosted jobs already ran.

Disposable private/ingress-network fixtures also served the generated module's
Hugo pages through the shared docs relay. URL discovery verified real responses,
distinct ports, stopped/recreated workspaces and unrelated-project exclusion.

## Remaining acceptance

The initial 2026-11-13 build attempt was blocked by the outer proxy at the
Docker CE signing-key endpoint. After the user rebuilt that environment on
2026-11-14, both that endpoint and the Go vulnerability database returned HTTP
200. The generated environment then built and started with the real pinned
Docker-in-Docker feature, completed post-create `task dev:setup`, and passed
`task doctor`, Codex/Hunk version checks, Docker Buildx and Compose checks.

Because this acceptance fixture runs inside another restricted devcontainer, its
proxy explicitly chained to the reachable outer proxy while retaining both
allowlists and destination checks. An HTTP failure exposed the inherited shared
Squid visible hostname; generated proxies now have module-specific identities.
The same Debian InRelease endpoint then returned HTTP 200. This is an explicit
proxy-configuration fix, not alternate downloads, disabled verification or a
substitute runtime. Image build output is now captured by the generated test
helper for failure diagnostics.

Both generated `task test:integration` and `task test:e2e` passed inside that
workspace using its separate nested Docker daemon (29.9.3). The suites built the
real Debian image and executed the probe and consumer agent. The test-only local
Dome source replacement and read-only source mount remain explicit; this does
not establish public module availability or native arm64 support. The live `task docs:serve` preview and `task devcontainer:urls` output also
served the home and development guide over the assigned loopback URL. Task 4.5 is
verified. Disposable containers, networks and test-owned volumes are cleaned up
after verification; ignored local logs retain detailed command evidence.

Public exact-tag/native-platform bootstrap without replacements, live generated
module Codex discovery/authoring, and actual GitHub draft-release gates remain
unverified and unchecked. Local tests do not establish those claims. No remote
repository, tag, push, release, commit or archive was created for this work.
