---
title: Releases
---

Create your Git repository and remote explicitly, push reviewed source, and use
immutable canonical v0.x.y or v1.x.y tags (including standard prereleases).
The generated workflow runs all required checks before creating a draft source
release using the job token. Enable repository Actions and grant release-job
contents write permission; no personal token is required. Inspect and publish
the draft explicitly. No agent binaries are distributed by this library workflow.

A pushed public tag can be resolved by Go before a GitHub Release is published.
Do not move published tags; fix with a new version and retract where appropriate.
Verify public exact-tag retrieval separately from local tests. Keep source,
licenses/notices and build instructions with distributions. Local go doc is the API
reference; no pkg.go.dev request, availability poll or indexing gate is used.
