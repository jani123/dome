# Delivery verification

## Scope and evidence

Use real target platforms and actual tool sessions; local cross-builds or static
skill parsing do not stand in for native execution or agent behavior. Reinspect
external state when this change is applied. Capture artifact URLs, exact versions,
commit IDs, checksums, runner/platform, and observed results in release/CI evidence.
No source-code change is needed merely to record the absence of evidence.

Public read-only inspection on 2026-11-13 at 07:05 UTC confirmed a public repository
with default branch main and Pages enabled. The Pages URL returned HTTP 404, there
were no public releases or remote tags, and the public API did not yet detect a
license. Local GPLv3 files do not mean those commits have been published. The
inspection environment was x86_64 and had no GitHub token or Git credential helper;
this is a dated observation, not a permanent limitation or claim about all users.

## Execution

Run the published exact-tag bootstrap on native amd64 and arm64, including the
normal skill and template versions. Check plan without resource mutation, explicit
apply, no-op repeat, content updates, unrelated-file preservation, invalid-config
failure, pruning, and recovery. Verify both initial RC and stable selection rules.

Fresh maintainer-session acceptance is owned by `adopt-project-knowledge-and-specs` task 5.4; reference its evidence rather than duplicating that check. Use a generated consumer project
for a separate live Codex discovery/authoring/build/plan/requested-apply exercise.
Do not confuse maintainer OpenSpec skills with the consumer Dome skill.

Verify repository Actions policy, exact required checks, native runner availability,
and github-pages environment rules with appropriate credentials. Confirm public
module resolution, selected-version local `go doc` output/examples, deployed
site/guide version paths, and source license/notices after the approved source/tag is public. Never move an immutable tag
or silently mark an unavailable environment successful.
