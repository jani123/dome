# Spec Delta

## ADDED Requirements

### Requirement: Initialize a runtime workspace

The compiled agent SHALL expose init to create a runtime workspace in the current directory or an explicit workspace directory. It SHALL persist source and state settings without reconciling resources or opening applied state. Runtime initialization SHALL remain distinct from generating Go source projects.

#### Scenario: Initialize once and apply later

- **WHEN** an operator initializes a workspace with Git URL, ref, configuration path and required SSH file references and later invokes apply there
- **THEN** apply uses the saved source and state settings without requiring the source flags again

#### Scenario: Select a workspace explicitly

- **WHEN** an operator invokes a runtime command with an explicit workspace directory
- **THEN** settings and workspace-relative paths resolve against that directory without changing the identity of existing managed resources or silently selecting a different state database

### Requirement: Prompt only for missing initialization input

Init SHALL accept a bootstrap configuration file and explicit flags and ask interactively for remaining required settings when terminal input is available. Complete supplied input SHALL require no prompts. Non-interactive or JSON invocations with missing input SHALL return validation errors without blocking. Runtime commands other than init SHALL NOT prompt.

#### Scenario: Interactive SSH setup

- **WHEN** an interactive init selects SSH Git and has no supplied URL, key path or known-hosts path
- **THEN** it requests those missing values and validates them before saving, without asking for private-key contents or trusting unknown hosts

#### Scenario: Automated setup

- **WHEN** init receives complete input with non-interactive mode
- **THEN** initialization succeeds without reading terminal input

#### Scenario: Incomplete automated setup

- **WHEN** non-interactive init, piped input reaching EOF, cancelled input or JSON setup lacks required values
- **THEN** it fails without publishing an initialized profile, mutating existing files or applying resources

### Requirement: Host runtime setup remains separate from repository authoring

Init SHALL create the local runtime settings scaffold and reference an existing local resource configuration or committed Git configuration. Git-backed initialization SHALL NOT create resource configuration examples, clone or initialize a source repository, or require a commit step on the managed host. Source-project generation SHALL remain in the existing developer initializer commands.

#### Scenario: Configure a Git-backed managed host

- **WHEN** an operator initializes a host with an existing Git configuration repository
- **THEN** only the local runtime profile and supporting workspace directories are created, and apply later retrieves resource configuration from the selected commit

#### Scenario: Configure an existing local source

- **WHEN** an operator selects local resource configuration
- **THEN** init validates and saves the existing file reference without generating or overwriting consumer-specific configuration

### Requirement: Safe repeatable initialization

Init SHALL preserve existing workspace profiles, configuration, keys and state. Equivalent completed setup SHALL be a no-op; conflicting setup SHALL fail without overwriting. Publication SHALL reject symlinks and concurrent replacement, and partial failure SHALL leave no valid initialized profile pointing at unfinished setup.

#### Scenario: Reinitialize a managed workspace

- **WHEN** init runs in a workspace containing a profile and existing execution history
- **THEN** equivalent settings are retained, conflicting changes are rejected, and execution history remains unchanged

#### Scenario: Interrupted publication

- **WHEN** initialization fails or a concurrent writer claims an output path
- **THEN** existing and concurrently written files remain intact and no incomplete workspace is reported as initialized

### Requirement: Workspace-oriented runtime help

Setup flags SHALL be shown on init help and hidden from root and other runtime command help, including inherited/global sections. Runtime help and normal usage examples SHALL present saved workspace configuration and workspace/profile selectors. Legacy source, credential, secrets and state override flags SHALL remain accepted without being advertised on runtime commands.

#### Scenario: Inspect plan and apply help

- **WHEN** an operator requests plan or apply help
- **THEN** help shows workspace/profile selection and command-specific options such as JSON without advertising Git URL/ref/path/HTTP, SSH key/known-hosts, resource configuration, secrets or state setup flags

#### Scenario: Inspect setup help

- **WHEN** an operator requests init help
- **THEN** help lists supported local/Git source, credential-file, secrets and state settings and unattended initialization options

#### Scenario: Hidden inherited setup flags

- **WHEN** an operator requests root, run, status, playbook or operation help
- **THEN** inherited/global help sections do not reintroduce setup flags

#### Scenario: Preserve existing scripted overrides

- **WHEN** an existing script supplies a legacy runtime setup flag explicitly
- **THEN** it retains the specified override behavior despite that flag being hidden from runtime help
