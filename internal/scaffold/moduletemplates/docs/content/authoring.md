---
title: Authoring
---

Read AGENTS.md, knowledge and OpenSpec specs before changes. Propose first, apply,
verify, and archive only completed scope. Implement Backend with truthful host
observation and safely repeatable apply/deletion. Do not store secrets in identity,
snapshots, logs or errors. Use local `go doc` for the selected dependency.

The test-only file adapter under tests/fixture is a real demonstration of the
contract, not your production host integration. Add domain-specific integration
and failure/recovery cases before claiming supported behavior.
