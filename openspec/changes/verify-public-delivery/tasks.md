# Delivery acceptance

- [x] 1. Reinspect public metadata/tags/releases/Pages and record dated observations without publishing changes.
- [x] 2. Verify authenticated Actions permissions, branch checks, Pages environment, and native runner availability.
- [x] 3. Exercise native amd64 and arm64 release bootstrap against the real public exact tag without local replacements.
- [ ] 5. Exercise real Codex discovery/authoring and requested apply in a generated consumer project; record versions/results.
- [ ] 6. Verify public tagged-module resolution and selected-version `go doc` output/examples, source license/notices, and RC/stable release artifacts with matching checksums/source; no pkg.go.dev indexing or availability gate.
- [ ] 7. Verify deployed Pages root/assets, immutable versions, development guides, and stable alias after authorized publication.
- [ ] 8. Verify consumer update and recovery behavior not covered by initial native bootstrap (task 3), retain exact-version evidence, and reference rather than repeat that bootstrap evidence.

RC portions of tasks 6 and 7 have evidence in
[verification](verification.md). Those tasks remain open for stable artifacts,
external site/alias acceptance and remaining scoped checks; RC evidence does not
stand in for a stable release or the owner's live hosting-project exercise.

## Acceptance ownership and dependencies

- Fresh maintainer-session acceptance is owned exclusively by [adopt-project-knowledge-and-specs, task 5.4](../adopt-project-knowledge-and-specs/tasks.md). Reference its evidence; do not run or track a duplicate here.
- Task 3 owns existing consumer native/public bootstrap; task 5 owns the existing consumer skill exercise. Module-starter platform/devcontainer/skill acceptance belongs to [initialize-module-repository](../initialize-module-repository/tasks.md), whose generated assets differ. Reuse shared public Dome release evidence where applicable.
- Task 6 owns actual public module/API/source-release acceptance. [use-local-api-documentation](../use-local-api-documentation/tasks.md) owns changing current documentation and local guidance checks, not repeating public release acceptance.
- Archive after this change's scoped acceptance is verified; other changes retain responsibility for their own completion.
