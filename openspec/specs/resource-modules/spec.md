# Resource modules

## Purpose

Retrospective specification of behavior inspected at `3e262920bbc20e41f843250f78a90c4935b9a036`.
This baseline documents existing implementation; it does not assert that these
requirements were authored before the code or that external acceptance passed.

## Requirements

### Requirement: Regular file safety

File resources SHALL use clean absolute identities, explicit safe modes, atomic replacement, and content-hash snapshots; non-regular or symlink paths SHALL be rejected.

Evidence: [modules/file/filesystem_integration_test.go](../../../modules/file/filesystem_integration_test.go).

#### Scenario: Regular file safety

- **WHEN** an owned path is replaced by a symlink
- **THEN** file observation or mutation fails rather than following the symlink

### Requirement: Injectable external operations

Built-in filesystem and systemctl operations SHALL be injectable for mocked unit tests; tests exercising actual host boundaries SHALL be tagged integration.

Evidence: [modules/systemd/systemd_test.go](../../../modules/systemd/systemd_test.go).

#### Scenario: Injectable external operations

- **WHEN** a unit test exercises service enable/start behavior
- **THEN** it uses the injected runner rather than host systemctl

### Requirement: Service support and pruning

Systemd resources SHALL accept loaded .service units with supported enabled/disabled and active/inactive/failed states; pruning SHALL stop and disable instead of uninstalling.

Evidence: [modules/systemd/systemd_test.go](../../../modules/systemd/systemd_test.go).

#### Scenario: Service support and pruning

- **WHEN** an unsupported/transitional service state is observed
- **THEN** the module fails explicitly instead of claiming convergence

## Evidence and limitations

- [modules/file/file.go](../../../modules/file/file.go)
- [modules/file/filesystem_integration_test.go](../../../modules/file/filesystem_integration_test.go)
- [modules/systemd/systemd.go](../../../modules/systemd/systemd.go)
- [modules/systemd/systemd_test.go](../../../modules/systemd/systemd_test.go)

Live systemd is not exercised by the normal containers because they do not run systemd. Service-unit installation and unsupported unit states are outside this module.

Rationale: [Built-in resources expose injectable host boundaries and explicit supported states.](../../../knowledge/architecture/file-and-systemd-boundaries.md).
