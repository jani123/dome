---
name: dome-module
description: Author and verify this standalone Dome resource-module library using the pinned API, OpenSpec and its real container test fixtures.
---

This skill matches Dome __VERSION__. Compare go list -m output with dome.json
skill_version/template_version/dome_version before API-dependent edits. For drift,
regenerate a temporary matching repository and review updates, preserving custom
source. Use `go doc -all __MODULE__` and
`go doc github.com/jani123/dome/core/resource` locally, not latest-main API guesses.
Read AGENTS.md, knowledge and specs. Propose an OpenSpec change before new behavior.
Use injectable host boundaries, stable non-secret identities/snapshots, truthful
observation and safe repeated apply/deletion. Restore from identity without old
secrets. The generated default backend is unfinished and must fail explicitly.
Run task check/test:race/check:artifacts and tagged test:integration/test:e2e for
real behavior. Test-only adapters are not production support. Apply host changes
only at the user's request through a consumer agent. Do not publish or push without
explicit authorization. Local API docs do not depend on hosted indexing.
