---
title: Failure and recovery
weight: 60
---

Start with `agent status`, `agent operations list`, and
`agent operations inspect <execution-id>`. To resume a standalone playbook use
`agent playbook run <name> --resume <execution-id>`. A subsequent reconciliation
resumes linked unfinished resource transitions, including interrupted pruning.

Persisted state is not evidence of host state. Completed steps are re-observed
and verified before skipping; interrupted non-repeatable work needs explicit
recovery. Unknown, failed, or in-progress observations must not lead to blindly
repeating external effects. Preserve graph fingerprints and callback versions.

A failed dependency blocks descendants; independent branches may finish.
A successful recovery branch does not hide the original error. Failed resources
must not advance their applied state, and failed reconciliation retains the last
successful overall Git revision. Earlier successful host changes can remain;
Dome does not roll back the whole host.

Do not delete SQLite state, change an unfinished transition target, or remove it
to escape recovery. Keep ownership and execution records and inspect the actual
host. Before upgrading, stop the agent and back up its state directory, including
SQLite sidecar files if present. Never copy a live SQLite database as an ordinary
single-file backup. Older binaries may not understand newer migrations.
