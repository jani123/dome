# Tasks

## 1. Implementation and verification

- [x] 1.1 Replace all remaining core Make wrappers and preserve initializer VERSION injection/default behavior; verify development and explicit candidate version outputs and task discovery.
- [x] 1.2 Migrate CI matrices/tool installation and maintained README/docs/rules/knowledge to Task; verify actionlint, shell syntax and no executable Make references remain in current workflows/tasks.
- [x] 1.3 Remove Makefile after checking equivalent build/docs/tool/artifact tasks; run `task check`, `task test:race`, `task test:e2e`, `task test:integration`, `task docs:build`, and `task check:artifacts`, reporting any failures or unavailable checks.

## Workflow follow-up

- Verify implementation, sync capability deltas, and archive only completed scope.
