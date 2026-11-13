# Proposal

## Why

Local implementation and tests do not establish native arm64 behavior, live
coding-agent usability, public exact-tag bootstrap, or deployed documentation.
Keep these delivery checks visible until supported by actual evidence.

## What Changes

- Verify the existing onboarding and release workflow against real supported
  platforms, coding-agent sessions, public module resolution, and GitHub settings.
- Validate a fresh session can use the new maintainer memory/spec workflow, and
  separately validate the generated consumer skill through a real Codex exercise.
- Publish reviewed artifacts/guides only after the release gates and publication
  authorization/access are satisfied; retain any unperformed check as active.

## Capabilities

### New Capabilities

None. This is acceptance work for existing implementations.

### Modified Capabilities

None. No supported behavior changes are proposed; `.openspec.yaml` explicitly
skips deltas. A defect requiring changed behavior gets its own scoped delta.

## Impact

Affects release evidence, GitHub configuration, native CI runners, guide/package
availability, and live maintainer/consumer workflows. No new runtime feature or
fictional reconstruction of past release approvals is included.
