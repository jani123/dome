# Public release acceptance

## Release and source identity

Observed on 2026-11-15. The owner explicitly authorized committing, pushing,
checking hosted pipelines and publishing an evaluation release before their
hosting-project acceptance.

[v0.1.0-rc.1](https://github.com/jani123/dome/releases/tag/v0.1.0-rc.1) was published
as a prerelease at 08:46:57 UTC, pointing at immutable commit
ff3620a4018e60b999ea02e834d92b9db794b6cb. No stable release was published and no
published tag was moved. Main contains the implementation, toolchain/dependency
security patches, devcontainer changes and scaffold-test maintenance.

## Hosted CI and native bootstrap

- [Initial main CI](https://github.com/jani123/dome/actions/runs/43693316131)
  passed all configured jobs, including native amd64 and arm64 starter checks.
- [Optimized main CI](https://github.com/jani123/dome/actions/runs/43694558851)
  passed all configured jobs after duplicate module checks were removed and the
  native cache namespace was separated from lighter jobs.
- [Release pipeline](https://github.com/jani123/dome/actions/runs/43695313195)
  passed its full reusable CI, packaging, both native archive/public-tag smoke
  jobs, and draft creation. The authorized draft was then published.

The native release smoke verifies SHA256SUMS, initializer version, public exact-tag
resolution without replacements, external consumer build/tests, agent init before
SQLite creation, plan, requested sandbox apply, no-op repeat and ownership pruning.
Recovery/update and live Codex acceptance remain separate open tasks.

The native Go-test steps on the optimized main cold-cache run took 488 seconds
(amd64) and 316 seconds (arm64); the identical release-source warm-cache steps
passed in 82 and 75 seconds. These are whole native test-step times, not isolated
unit-test timings. The smaller local module smoke passed in 4.80 seconds; the
local full acceptance comparison was 124.26 before and 91.01 seconds after the
maintenance, with cache/download/image variability. Real race and container
coverage is retained in full acceptance.

## Authenticated repository and deployment configuration

The authenticated account has repository administration/push access. Actions are
enabled and permit the pinned actions. Main CI checks passed; main has no branch
protection configured. Native ubuntu-24.04 and ubuntu-24.04-arm runners executed
successfully. Pages uses workflow deployment at https://jani123.github.io/dome/.

The automatic release-triggered
[Pages run](https://github.com/jani123/dome/actions/runs/43696100110) built the
versioned site successfully but its deployment was rejected because github-pages
allowed only main branch events. The annotation explicitly identified the release
tag as disallowed. The environment now allows branch main and tag pattern v0.*;
source checkout remains trusted main. GitHub would not retry the rejected run, so
[a replacement from the same tag](https://github.com/jani123/dome/actions/runs/43696272265)
was dispatched and both build and deployment passed. Its uploaded artifact grew
from 9,132 bytes for development-only docs to 16,329 bytes with the RC snapshot.
Local assembly of the published-tag site also passed link validation. Actions
success and artifact size alone are not counted as external rendered-site or
stable-alias acceptance; task 7 remains open. The initial rejection is retained as
historical evidence rather than hidden or called passing.

## Public module, API and corresponding source

A fresh external module with GOWORK disabled resolved
`github.com/jani123/dome@v0.1.0-rc.1`, reported that exact version with no replacement,
and successfully ran selected-version go doc for agent.New and workspace.Profile.
No pkg.go.dev availability or indexing check was used.

All three published archive SHA-256 values matched SHA256SUMS and GitHub's uploaded
asset digests. The source archive was additionally downloaded and hashed:

| Archive | SHA-256 |
| --- | --- |
| Linux amd64 | 8c3a85319253a5c1bdb46f65975f1a17c83feb7a0d1ef802b00c9f26a07de01a |
| Linux arm64 | 8941324a3164a7ff48fd63b6e3cacd5225afa58e35d7e5cf79f8feae4357f17a |
| Corresponding source | 8da686808bb9dd2a23abb5498dc7494424604bad4eb6ab3ab7d4d3e809727281 |

The source bundle contains LICENSE, BUILD-SOURCE.md, vendor/modules.txt and upstream
license material. It rebuilt the initializer on Linux amd64 with Go 1.26.11,
CGO_ENABLED=0, GOWORK=off, GOPROXY=off and -mod=vendor, and the binary reported the
correct release version. This establishes a build from the shipped source, not
byte-for-byte reproducibility. GitHub's license metadata now identifies GPL-3.0.
Native binary archives include the generated third-party notices as packaged by
the successful release workflow. Stable artifact selection remains unexercised.

## Remaining acceptance

Live Codex consumer/module authoring, the owner's hosting-project exercise,
module-specific public/native bootstrap and hosted fixture release gates, stable
artifacts/alias behavior, and consumer update/recovery acceptance remain open in
their owning changes. Core public release evidence does not substitute for a
separate generated-module forge fixture or a stable release.

Supporting API results, public-module/go doc output, downloaded checksum/source,
source rebuild and local site-build logs remain ignored under
`.build/notes/release-rc1-public/` and `.build/notes/public-rc1-module.log`.
