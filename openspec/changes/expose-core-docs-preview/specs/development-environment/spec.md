## ADDED Requirements

### Requirement: Discoverable local documentation preview

The devcontainer SHALL publish Hugo container port 1313 on a dynamically assigned host-loopback port. `task docs:serve` SHALL serve guides and `task devcontainer:urls` SHALL report current workspace bindings without starting containers or listing unrelated containers. Discovery SHALL distinguish mappings from readiness and explain remote Docker access.

#### Scenario: Discover the docs preview

- **WHEN** a contributor starts the devcontainer Hugo server and requests workspace URLs on the host
- **THEN** discovery reports the current mapped docs URL and guides are reachable there

#### Scenario: Workspace is stopped

- **WHEN** no matching workspace is running
- **THEN** discovery reports that state with startup guidance and does not start containers


#### Scenario: Concurrent workspaces

- **WHEN** two workspaces serve guides concurrently
- **THEN** each reports its own distinct dynamically assigned localhost URL

#### Scenario: Recreated workspace

- **WHEN** a workspace is recreated and URL discovery is run again
- **THEN** discovery reports current bindings without cached stale addresses

#### Scenario: Unrelated containers

- **WHEN** unrelated containers publish ports on the same Docker daemon
- **THEN** discovery excludes their bindings from the workspace results
