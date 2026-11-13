---
type: Architecture
title: "Agent composition"
description: "Explicit Go composition keeps the core independent of concrete modules."
generated: {by: codex/2026-11-13, at: "2026-11-13T07:03:33Z"}
sources:
  - id: source-1
    resource: ../../agent/agent.go
  - id: source-2
    resource: ../../core/resource/resource.go
  - id: source-3
    resource: ../../core/registry/registry.go
  - id: source-4
    resource: ../../tests/integration/consumer_test.go
  - id: source-5
    resource: ../../agent/example_test.go
---

# Agent composition

Agents are compiled Go programs owned by their consumers. Explicit registration makes dependencies visible at construction and permits external modules to implement public contracts without importing internal packages. Runtime plugins, global module registration, and an infrastructure DSL are outside the architecture.

## Boundary

The demonstration agent contains intentional crash/failure hooks and is not the clean starter.

Behavioral contract: [Agent composition](../../openspec/specs/agent-composition/spec.md).
