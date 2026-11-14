# Design

## Context

Core already pins Hugo and uses Compose and host-side Dev Container CLI tasks. Existing workspace networking is private with proxy egress.

## Decisions

Serve on container port 1313 with loopback-only dynamic host publication. Use relative guide links and disable live reload to avoid a second port. Discover running workspace containers by the Dev Container CLI local-folder label plus Compose service label, then inspect all published ports. Keep discovery read-only and explain mapping versus readiness and remote Docker hosts.

## Risks / Trade-offs

Existing workspaces require recreation for the new port binding; no live workspace is recreated automatically. Tests use disposable containers instead of interrupting the active development environment.

## Network verification refinement

A disposable real-Docker check found that internal-only bridge workspaces do not expose published ports. Preserve the isolated private workspace network and proxy-only egress. A minimal nginx docs relay joins the private and ingress networks, forwarding only to workspace:1313 and publishing a random loopback host port. URL discovery resolves the workspace's Compose project and lists that project's published service bindings, excluding the internal egress proxy. Verify the actual relay topology with disposable containers rather than only a default bridge fixture.
