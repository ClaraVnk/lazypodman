# Accepted vulnerabilities

This file lists vulnerabilities flagged by `govulncheck` that we have reviewed and deliberately accept for the time being. The CI security workflow allows these specific IDs through; any **other** vulnerability fails the build.

| ID | Module | Found in | Status | Notes |
|---|---|---|---|---|
| [GO-2026-4887](https://pkg.go.dev/vuln/GO-2026-4887) | `github.com/docker/docker` | `v28.5.2+incompatible` | No upstream fix | The Docker backend was removed in [ADR 0002, Phase 6](../adr/0002-port-docker-sdk-to-podman.md#phase-6--drop-the-docker-backend-rename-module-path), so lazypodman no longer imports `docker/docker` directly. govulncheck nonetheless still reports it reachable: the containers/podman bindings tree itself depends on the moby/docker types, so removing our backend cannot eliminate it. No upstream fix; not eliminable without an upstream change to the Podman bindings. |
| [GO-2026-4883](https://pkg.go.dev/vuln/GO-2026-4883) | `github.com/docker/docker` | `v28.5.2+incompatible` | No upstream fix | Same as above — reachable transitively via the containers/podman tree, not via lazypodman's own code. |
| [GO-2026-5668](https://pkg.go.dev/vuln/GO-2026-5668) | `github.com/docker/docker` | `v28.5.2+incompatible` | No upstream fix | Same as above — reachable transitively via the containers/podman tree, not via lazypodman's own code. |
| [GO-2025-3961](https://pkg.go.dev/vuln/GO-2025-3961) | `github.com/containers/podman/v5` | `v5.8.3` | No upstream fix | Reachable through the Podman bindings added in [ADR 0005](../adr/0005-podman-native-backend.md). No fix in the latest stable v5; tracked upstream. |
| [GO-2024-3042](https://pkg.go.dev/vuln/GO-2024-3042) | `github.com/containers/podman/v5` | `v5.8.3` | No upstream fix | Same as above — reachable via the Podman bindings tree; no fix in v5.8.3. |
| [GO-2026-5037](https://pkg.go.dev/vuln/GO-2026-5037) | stdlib (`crypto/x509`) | `go1.26.3` | Fixed in go1.26.4 | Toolchain vulnerability, not a dependency. Accepted only until the CI toolchain ships ≥ go1.26.4, then drop. |
| [GO-2026-5039](https://pkg.go.dev/vuln/GO-2026-5039) | stdlib (`net/textproto`) | `go1.26.3` | Fixed in go1.26.4 | Same as above — fixed by a toolchain bump to go1.26.4; remove once CI runs it. |
| [GO-2026-5856](https://pkg.go.dev/vuln/GO-2026-5856) | stdlib (`crypto/tls`) | `go1.26.3` | Fixed in go1.26.5 | Toolchain vulnerability (Encrypted Client Hello privacy leak), not a dependency. Accepted only until the CI toolchain ships ≥ go1.26.5, then drop. |
| [GO-2026-5853](https://pkg.go.dev/vuln/GO-2026-5853) | `github.com/sigstore/fulcio` | `v1.8.5` | Not exercised | SSRF / JWKS substitution in Fulcio's OIDC discovery. The only trace govulncheck reports is `podman.init → types.init → certificate.init`, i.e. package initialisation — lazypodman never performs sigstore signing or OIDC issuer discovery, so the vulnerable code path is unreachable at runtime. Fixed in v1.8.6. The transitive cascade that originally deferred this bump (grpc, otel) has since been taken for unrelated reasons, so the remaining cost is limited to fulcio itself plus google-api/goa — reassess the bump next time this file is touched. |
| [GO-2026-5116](https://pkg.go.dev/vuln/GO-2026-5116) | `github.com/containers/buildah` | `v1.43.2` | Not exercised — no fix | Build breakout via a malicious Containerfile or Git HTTP server. lazypodman never imports `containers/buildah` (zero direct imports) and never builds images: it is a TUI over the Podman socket, and image builds are performed by the Podman service, not in our process. The traces govulncheck reports are interface dispatch and package initialisation, not the vulnerable path — `utils.ColoredStringDirect → fmt.Sprint → define.Isolation.String` (and the same for `NetworkConfigurationPolicy.String` / `PullPolicy.String`) only reach `String()` methods that `fmt` can dispatch to on any value in the binary, plus `podman.init → types.init → define.init`. No fixed version published upstream. Revisit when buildah ships a fix. |
| [GO-2026-5932](https://pkg.go.dev/vuln/GO-2026-5932) | `golang.org/x/crypto` | `v0.52.0` | No fix — package unmaintained | `x/crypto/openpgp` is deprecated and unmaintained upstream; there is no fixed version. Pulled in transitively by the containers/image tree, and reinforced by the `containers_image_openpgp` build tag we set to avoid a cgo dependency on gpgme. Not eliminable without an upstream migration away from `x/crypto/openpgp`. |

## How the allowlist works

`.github/workflows/security.yml` runs `govulncheck ./...` and parses its text output, keeping only `Vulnerability #N: GO-YYYY-NNNN` lines — those are the vulnerabilities reachable from our code, as opposed to the ones merely present in modules we require. It then compares those IDs against this list (parsed from this file). The build fails if any **unknown** vulnerability is reported. If a vulnerability listed here is no longer reported, the entry should be removed from this file.

## When to remove an entry

- Upstream releases a fix and we bump the dependency → drop the entry.
- The migration removes the call path (ADR 0002 Phase 6 for the Docker SDK entries) → drop the entries.
- The vulnerability turns out not to apply to us after deeper analysis → drop the entry and document why in this file's history.

**Do not** add an entry just to make the build green. Any addition requires a written justification in this file.
