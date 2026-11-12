---
title: Work with Codex
weight: 20
---

Open Codex in the generated project and ask it to use the `dome` skill. The project
includes `.agents/skills/dome/SKILL.md` and consumer `AGENTS.md`. Restart the coding
agent if needed for discovery. Installing the Go dependency alone does not install
these files. See [Codex skill discovery](https://learn.chatgpt.com/docs/customization/overview).

A useful first request is: “Use the Dome skill to add another sandbox file. Build
and test the agent and show me its JSON plan.” Request apply when you want those
changes made. No model or coding-agent service is required at runtime.

The skill checks the dependency and `dome.json` versions before authoring. Use
`go doc` against installed source; APIs from main may differ from a pinned release.
A mismatch requires reviewing and updating the dependency and generated assets
together. Keep custom resources and instructions during upgrades.
