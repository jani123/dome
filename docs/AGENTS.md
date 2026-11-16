# Documentation authoring rules

Follow the repository rules in [../AGENTS.md](../AGENTS.md). These rules describe
how to write and present Dome's user guides. Keep generated project instructions
consistent with the same conventions.

## Structure and wording

- Write for someone following the procedure for the first time. Lead with the
  action and use short, numbered steps for walkthroughs.
- Keep prerequisites in bullet points. Mark optional tools and steps explicitly,
  and explain where readers continue if they skip an optional step.
- Keep the main path concise: commands, example files, and observable results.
  Put background and operational detail in the relevant reference guide.
- State which file to edit and provide usable example code or configuration.
  Name expected outputs and managed paths so readers can verify the result.
- Do not repeat instructions in several places on the same page. Keep opening,
  closing, and re-entry guidance with the alternative to which it belongs.

## Commands and examples

- Show shell commands in `sh` code blocks. Use short `#` comments for procedural
  instructions inside command panels instead of paragraphs interleaving inline
  commands. Describe editor actions as comments when presenting an editor-based
  alternative in the same format.
- Put each command on its own line. Separate distinct actions with blank lines;
  split long commands with shell line continuations. Avoid dense chained commands.
- Do not include shell prompts such as `$` in copyable commands. Put expected
  output in a separate `text` block, not among commands.
- Make the execution context explicit: host, temporary Docker container, or
  generated development container. Show when to leave a shell before the next
  host command. Keep later agent runs in the same environment and workspace path.
- Keep instructions such as “when finished” in comments so stop/exit commands are
  clearly separate from startup and from the remaining walkthrough.
- Avoid unnecessary host installations. Offer the supported container alternative
  where useful, while preserving the local Go path and optional devcontainer use.
- Explain placeholders where they occur. Distinguish a project's Go module
  identity from an external dependency; a local project identity need not exist
  online. Label illustrative external dependencies rather than implying they exist.
- Derive initializer versions and flags from the selected guide version. Use the
  existing version shortcodes/template logic; do not hard-code a planned release
  into development guides or give older releases unsupported flags.

## Tabbed alternatives

- Use tabs for alternative ways to complete the same step, such as Go versus
  Docker bootstrap, or Dev Container CLI versus VS Code. Do not put sequential
  steps into separate tabs or add tabs for information all readers need.
- Default to the simplest direct path: Go for creation and CLI for opening the
  development container. Keep the Docker/editor alternatives available without
  displaying all of their content at once.
- Keep each alternative self-contained. Include its commands, brief comments,
  context switches, and closing instructions inside its own panel.
- Present tabs as one bordered panel with a clearly highlighted selected tab.
  Clicking a tab must switch its content and selected state. Preserve keyboard
  navigation, focus visibility, ARIA relationships, and independent tab groups.
- Reuse the `command-tabs` markup, `data-tabs` behavior, and shared tab script.
  Give every tab and panel a unique ID on its page. Ensure the controls fit mobile
  widths and wide commands scroll inside the panel without widening the page.

## Rendering and verification

- Guide content lives in `docs/content/`; Hugo templates and shortcodes live in
  `docs/layouts/`; styles and scripts live in `docs/assets/`. Keep the fingerprinted
  asset pipeline so changes receive fresh URLs. Do not enable arbitrary raw HTML
  in Markdown to work around rendering problems.
- Inspect the rendered page, not only Markdown source. Shortcode tags must render
  as panels, in the right place, in both clean builds and the running preview.
- Use the relevant checks: `scripts/build-docs.sh next`,
  `bash scripts/test-docs-versions.sh`, `node --test scripts/test-doc-tabs.mjs`, and
  `node --test scripts/test-docs-preview.mjs`. Run `task check:artifacts` when
  maintained documentation, knowledge, or specs change.
- Exercise changed walkthroughs with their actual tools and dependencies. For the
  consumer container workflow, use `scripts/test-consumer-devcontainer.sh` with
  the real Dev Container CLI. Test first apply, repeated no-op, and re-entry when
  relevant; do not substitute a stub for the documented tool.
- Report what was actually verified. Local-source checks, public version
  resolution, browser interaction, and VS Code GUI acceptance are distinct; do
  not claim an unexercised path passed.
