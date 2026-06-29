# Accepted vulnerabilities

This file lists vulnerabilities flagged by `govulncheck` that we have reviewed and deliberately accept for the time being. The CI security workflow allows these specific IDs through; any **other** vulnerability fails the build.

| ID | Module | Found in | Status | Notes |
|---|---|---|---|---|
| [GO-2026-4887](https://pkg.go.dev/vuln/GO-2026-4887) | `github.com/docker/docker` | `v28.5.2+incompatible` | No upstream fix | Docker backend removed in [ADR 0002, Phase 6](../adr/0002-port-docker-sdk-to-podman.md#phase-6--drop-the-docker-backend-rename-module-path); `docker/docker` is no longer imported by lazypodman and now appears only as a transitive dependency of the containers/podman tree. Retained until the CI govulncheck report confirms it is no longer reachable, then drop. |
| [GO-2026-4883](https://pkg.go.dev/vuln/GO-2026-4883) | `github.com/docker/docker` | `v28.5.2+incompatible` | No upstream fix | Same as above — Docker SDK now transitive-only; pending govulncheck confirmation of non-reachability. |
| [GO-2026-5617](https://pkg.go.dev/vuln/GO-2026-5617) | `github.com/docker/docker` | `v28.5.2+incompatible` | No upstream fix | Same as above — Docker SDK now transitive-only; pending govulncheck confirmation of non-reachability. |
| [GO-2026-5668](https://pkg.go.dev/vuln/GO-2026-5668) | `github.com/docker/docker` | `v28.5.2+incompatible` | No upstream fix | Same as above — Docker SDK now transitive-only; pending govulncheck confirmation of non-reachability. |
| [GO-2026-5746](https://pkg.go.dev/vuln/GO-2026-5746) | `github.com/docker/docker` | `v28.5.2+incompatible` | No upstream fix | Same as above — Docker SDK now transitive-only; pending govulncheck confirmation of non-reachability. |
| [GO-2025-3961](https://pkg.go.dev/vuln/GO-2025-3961) | `github.com/containers/podman/v5` | `v5.8.3` | No upstream fix | Reachable through the Podman bindings added in [ADR 0005](../adr/0005-podman-native-backend.md). No fix in the latest stable v5; tracked upstream. |
| [GO-2024-3042](https://pkg.go.dev/vuln/GO-2024-3042) | `github.com/containers/podman/v5` | `v5.8.3` | No upstream fix | Same as above — reachable via the Podman bindings tree; no fix in v5.8.3. |
| [GO-2026-5037](https://pkg.go.dev/vuln/GO-2026-5037) | stdlib (`crypto/x509`) | `go1.26.3` | Fixed in go1.26.4 | Toolchain vulnerability, not a dependency. Accepted only until the CI toolchain ships ≥ go1.26.4, then drop. |
| [GO-2026-5039](https://pkg.go.dev/vuln/GO-2026-5039) | stdlib (`net/textproto`) | `go1.26.3` | Fixed in go1.26.4 | Same as above — fixed by a toolchain bump to go1.26.4; remove once CI runs it. |

## How the allowlist works

`.github/workflows/security.yml` runs `govulncheck -format json ./...`, then a small filter compares the reported vulnerability IDs against this list (parsed from this file). The build fails if any **unknown** vulnerability is reported. If a vulnerability listed here is no longer reported, the entry should be removed from this file.

## When to remove an entry

- Upstream releases a fix and we bump the dependency → drop the entry.
- The migration removes the call path (ADR 0002 Phase 6 for the Docker SDK entries) → drop the entries.
- The vulnerability turns out not to apply to us after deeper analysis → drop the entry and document why in this file's history.

**Do not** add an entry just to make the build green. Any addition requires a written justification in this file.
