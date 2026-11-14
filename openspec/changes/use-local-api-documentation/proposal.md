# Proposal

## Why

Core API guidance should follow the selected Go dependency through `go doc`, independently of a hosted documentation index.

## What Changes

- Make local `go doc` the core API reference in README, Hugo guides/layout and authoring assets.
- Remove explicit indexing or hosted-documentation acceptance requirements while preserving public module resolution and source/license verification.
- Explain that public proxy downloads may still trigger third-party indexing.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `documentation-publishing`: Core API guidance should follow the selected Go dependency through `go doc`, independently of a hosted documentation index.

## Impact

README, docs layouts/content, knowledge, current release acceptance and active verify-public-delivery artifacts; no runtime API changes.
