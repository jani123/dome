# Architecture

- [Agent composition](library-composition.md) - Explicit Go composition keeps the core independent of concrete modules.
- [Reconciliation](reconciliation-and-ownership.md) - Ownership and fresh verification prevent accidental pruning and false applied state.
- [Durable state](state-and-locking.md) - SQLite records ownership and attempts while host observations remain authoritative.
- [Playbook execution](dag-execution-and-recovery.md) - Explicit DAG edges and recovery contracts make interruptions safe to revisit.
- [Resource transitions](transition-target-recovery.md) - A matching host target cannot bypass an unfinished linked workflow.
- [Resource modules](file-and-systemd-boundaries.md) - Built-in resources expose injectable host boundaries and explicit supported states.
