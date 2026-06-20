# Accepted vulnerabilities

This file lists vulnerabilities flagged by `govulncheck` that we have reviewed and deliberately accept for the time being. The CI security workflow allows these specific IDs through; any **other** vulnerability fails the build.

| ID | Module | Found in | Status | Notes |
|---|---|---|---|---|
| [GO-2026-4887](https://pkg.go.dev/vuln/GO-2026-4887) | `github.com/docker/docker` | `v28.5.2+incompatible` | No upstream fix | Reachable from `pkg/commands` (Docker SDK). Will be eliminated when the Docker SDK is removed in [ADR 0002, Phase 6](../adr/0002-port-docker-sdk-to-podman.md#phase-6--drop-the-docker-backend-rename-module-path). |
| [GO-2026-4883](https://pkg.go.dev/vuln/GO-2026-4883) | `github.com/docker/docker` | `v28.5.2+incompatible` | No upstream fix | Same as above — reachable via the inherited Docker SDK code path. |

## How the allowlist works

`.github/workflows/security.yml` runs `govulncheck -format json ./...`, then a small filter compares the reported vulnerability IDs against this list (parsed from this file). The build fails if any **unknown** vulnerability is reported. If a vulnerability listed here is no longer reported, the entry should be removed from this file.

## When to remove an entry

- Upstream releases a fix and we bump the dependency → drop the entry.
- The migration removes the call path (ADR 0002 Phase 6 for the Docker SDK entries) → drop the entries.
- The vulnerability turns out not to apply to us after deeper analysis → drop the entry and document why in this file's history.

**Do not** add an entry just to make the build green. Any addition requires a written justification in this file.
