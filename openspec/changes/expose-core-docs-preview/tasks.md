# Tasks

## 1. Preview and discovery

- [x] 1.1 Add direct serve and discovery tasks/scripts and Compose mapping; validate shell syntax and task discovery.
- [x] 1.2 Update specs, knowledge and usage docs; validate static guides and maintained artifacts.
- [x] 1.3 Verify real random-port HTTP serving, concurrent workspaces, and stopped discovery with disposable containers.

- [x] 1.4 Verify recreated-workspace bindings and explicit exclusion of unrelated labelled/unlabelled containers with disposable fixtures; do not infer passing results from inspection alone.

## Verification evidence

- `bash -n scripts/serve-docs.sh .devcontainer/urls.sh` passed.
- `docker compose -f .devcontainer/compose.yaml config --quiet` passed.
- Task discovery listed the new preview and URL tasks.
- `task docs:build` built the guides and passed local path checks.
- Two disposable Debian containers running the pinned Hugo served the home and quickstart pages over real HTTP on distinct Docker-assigned loopback ports (32768 and 32769 during that run). URL discovery reported the matching workspace addresses.
- After stopping one fixture, discovery reported no running workspace without starting it. All fixture containers were removed.
- Strict OpenSpec validation, `task check:artifacts`, and `git diff --check` passed.

These checks establish local preview behavior, not public Pages deployment or a rebuild of the active workspace. The active devcontainer was deliberately left running unchanged; contributors apply its new mapping through the documented rebuild command. Agent runtime code did not change, so the runtime unit/race/E2E/Git integration suites were not rerun for this tooling change.

## Additional topology verification

The isolated workspace network suppresses direct host port publishing. A
read-only unprivileged nginx relay now joins private/ingress networks and forwards
only to workspace:1313. Three disposable project fixtures verified real home and
quickstart HTTP responses on distinct assigned loopback ports, recreation, stopped
workspace detection, and exclusion of another running Compose project. All
containers/networks were removed. Compose configuration validation passed.
