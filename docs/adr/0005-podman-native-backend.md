# 0005 — Podman native backend (Phase 3)

- **Date** : 2026-06-21
- **Status** : Accepted — Phase 3 implemented (all method groups except SDK-native attach) and the default backend flipped to Podman (ADR 0002 Phase 4), gated by the compliance suite. Phase 6 since removed the Docker backend entirely (Podman-only), so the suite now runs Podman alone.
- **Relates to** : [ADR 0002 — Port from Docker SDK to Podman bindings](0002-port-docker-sdk-to-podman.md) (implements its Phase 3), [ADR 0003 — Runtime interface and domain types](0003-runtime-interface-and-domain-types.md)

## Context

Phase 1 is done: `pkg/gui` and `pkg/commands` talk to the engine exclusively through `runtime.ContainerRuntime`, and the Docker SDK is isolated inside `pkg/runtime/docker`. The interface deals in lazypodman-owned `domain` types, so a second backend is now a self-contained package that implements the same interface — no GUI or command churn.

Today lazypodman already runs against Podman through Podman's **Docker-compatible** REST socket (point `DOCKER_HOST` at `unix://$XDG_RUNTIME_DIR/podman/podman.sock`). That works but:

- it does not expose Podman-native concepts (pods, quadlets, `generate kube`, healthchecks the Podman way) — these are the reason the fork exists (Phase 5);
- the compat layer has historical gaps (exec/attach edge cases, stats shape differences);
- it keeps us semantically coupled to Docker's API as the contract.

Phase 3 adds a second `runtime.ContainerRuntime` implementation backed by Podman's native Go bindings: [`github.com/containers/podman/v5/pkg/bindings`](https://github.com/containers/podman/blob/main/pkg/bindings/README.md). It is **opt-in** this phase; the default stays `docker` until parity is verified (Phase 4).

## Decision

### Package layout

Add `pkg/runtime/podman`, mirroring `pkg/runtime/docker`: one `Runtime` struct satisfying `runtime.ContainerRuntime`, a `mapper.go` translating Podman bindings types ↔ `domain` types, an `errors.go` mapping Podman errors onto the existing `runtime` sentinels (`ErrNotFound`, `ErrConflict`, `ErrUnauthorized`, `ErrUnsupported`, `ErrUnavailable`), and a `connect.go` for connection/host resolution.

The compile-time check `var _ runtime.ContainerRuntime = (*Runtime)(nil)` is the contract that keeps both backends honest.

### Connection model

Podman bindings differ from the Docker client object: `bindings.NewConnection(ctx, uri)` returns a **`context.Context` that carries the connection**, and every call threads that context (`containers.List(conn, opts)`, `images.List(conn, opts)`, …). The adapter stores the connection context built once at construction and derives per-call contexts from it (cancellation/timeout compose normally).

Socket/URI resolution, in decreasing precedence:

1. `CONTAINER_HOST` environment variable (Podman's native equivalent of `DOCKER_HOST`);
2. rootless default `unix://$XDG_RUNTIME_DIR/podman/podman.sock`;
3. rootful default `unix:///run/podman/podman.sock`.

Remote Podman over SSH (`ssh://user@host/run/...`) is supported by the bindings natively via the URI; wiring it through config is deferred (see open questions).

### Backend selection

- New config field `runtime: docker | podman` in `pkg/config` (default `docker` this phase).
- Env override `LAZYPODMAN_RUNTIME` (takes precedence over config).
- A small factory in `pkg/commands` (e.g. `newRuntime(cfg) (runtime.ContainerRuntime, error)`) picks the implementation. `NewDockerCommand` already constructs the runtime in one place (Phase 1d.5), so the factory slots in there. The struct/type name `DockerCommand` is left as-is until Phase 6's mechanical rename, to keep upstream cherry-picks cheap.

Fail-fast: an unknown `runtime:` value is a config error at startup, not a silent fallback.

### Dependency strategy — the main risk

The Podman bindings pull a **large** transitive tree (`containers/common`, `containers/image`, `containers/storage`). Parts of that tree historically require CGO and C libraries (`gpgme`, `libdevmapper`, `btrfs`) behind build tags. We do **not** want CGO or those system libs in lazypodman's build.

`pkg/bindings` is an HTTP client to the Podman socket and should not *need* the storage/image graph drivers at runtime, but Go's package graph can still drag them in at compile time. Mitigation, to be validated by a spike (Phase 3a) **before** committing the vendor bump:

- Build with the standard Podman client tags: `containers_image_openpgp` (pure-Go OpenPGP, drops `gpgme`/CGO), `exclude_graphdriver_btrfs`, `exclude_graphdriver_devicemapper`, and `remote` where applicable.
- Measure the vendored size delta and confirm `CGO_ENABLED=0` still builds on linux + windows.
- Run `govulncheck` (reachable-only, as the existing CI already does for the Docker SDK) and `osv-scanner` on the new tree; extend the allowlist only for unreachable upstream advisories.
- If the tree proves unacceptable (CGO unavoidable, size explosion, cross-compile breakage), **fall back** to a thin hand-written HTTP client against the documented Podman REST API rather than adopting `pkg/bindings`. This keeps Phase 3 reversible.

Go version: `go.mod` currently declares `go 1.22`. Podman v5 forces `go 1.25` (measured); bump `go.mod` and the CI matrix in the same PR that adds the dependency (Phase 3b — see refinement below).

### Spike results (Phase 3a)

The dependency strategy was validated on a throwaway branch against `github.com/containers/podman/v5@v5.8.3`:

- **`CGO_ENABLED=0` builds clean** with tags `containers_image_openpgp exclude_graphdriver_btrfs exclude_graphdriver_devicemapper remote` — no CGO, no system libs. The thin-HTTP-client fallback is therefore **not** needed.
- Footprint: **+73 modules**, ~535 non-stdlib packages, ~54 MB of new sources (`containers` 32M + `go.podman.io` 14M + `opencontainers` 7.6M). Forces `go 1.22 → 1.25`.
- `govulncheck -scan module` (upper bound, no reachability): 7 advisories — 3 stdlib (toolchain bump), 1 windows-only `x/sys`, 3 in `docker/docker`+`docker/cli` already present in the tree and already allowlisted in CI. **No net-new vulnerability from the Podman bindings.**

**Version choice**: target **v5.8.3** (latest stable v5, aligned with the 5.x daemon ecosystem). The module is deprecated in favour of `go.podman.io/podman/v6`, but **v6 has no stable release yet** (only `v6.0.0-rc1`), so the migration to the `go.podman.io` path is deferred until v6 is GA.

**Refinement vs the sub-PR sketch below**: the dependency (and the `go 1.25` bump) is deferred from 3a to **3b**, where the first real method group needs a live connection. 3a ships dependency-free scaffolding, keeping its branch tip green with no dead dependency.

### Staged sub-PRs

Mirroring the Phase 1d split — each independently buildable, with the Docker backend untouched and default:

- **3a — Spike + scaffolding (done).** Validated the dependency strategy on a throwaway branch (see "Spike results"), then landed `pkg/runtime/podman` with the `Runtime` satisfying the interface via `runtime.ErrUnsupported` stubs, the config field, env override and factory. Default stays docker; podman backend selectable but non-functional. **Dependency-free** — the bindings land in 3b.
- **3b — Containers + dependency.** Add the `go.podman.io`/`containers/podman/v5` dependency (vendored, `go 1.25` bump, CI build tags), `connect.go` (bindings connection + `CONTAINER_HOST`/rootless/rootful resolution), and the first real method group: List/Inspect/Start/Stop/Restart/Pause/Unpause/Remove/Top/Prune + mapper.
- **3c — Images.** List/Remove/History/Prune + mapper.
- **3d — Networks + Volumes.** List/Remove/Prune + mappers.
- **3e — Events, Stats, Logs.** Event stream → `domain.Event`; stats stream → `domain.Stats`; log stream demux (Podman multiplexes differently from Docker — verify the framing).
- **3f — Dual-backend compliance suite.** A table-driven test exercising the full `ContainerRuntime` surface against a live socket, parameterized over both backends, gated on socket availability (skipped in unit CI, run in an integration job that brings up rootless Podman — and Docker-compat for the docker backend).

### Verification gate at every sub-PR

Same as ADR 0004: `go build`/`go vet`/`gofmt`/`go test` green, plus — for 3b onward — a manual smoke with `LAZYPODMAN_RUNTIME=podman` against a real rootless Podman socket, navigating the relevant panels. The compliance suite (3f) becomes the durable parity gate that Phase 4 (flip the default) depends on.

## Consequences

- **Pro**: the fork finally talks to Podman natively, unlocking Phase 5 (pods, quadlets, generate-kube). Both backends coexist behind one interface, so regressions are isolated per-backend.
- **Pro**: the compliance suite gives an objective, reusable definition of "parity" for the Phase 4 default flip.
- **Con**: dependency footprint grows significantly; the build-tag/CGO story is the real risk and is gated behind a spike with a documented fallback.
- **Con**: two backends to maintain through Phases 3–5 (mitigated by the shared interface + compliance suite; the Docker backend is dropped in Phase 6).

## Open questions

- **Remote Podman** (`CONTAINER_HOST=ssh://…`): expose via config now, or defer until after parity? Leaning defer — local rootless first.
- **Pods**: out of scope here (Phase 5), but the mapper should not actively preclude a future `domain.Pod`.
- **Compose**: Podman's `podman compose` shim vs native — unchanged by this ADR; the `CommandTemplates.DockerCompose` path stays CLI-based for now.
- **Buildah / Skopeo**: separate ADR if/when image build/copy features are added.

## References

- Podman Go bindings tutorial : https://github.com/containers/podman/blob/main/pkg/bindings/README.md
- Podman build tags : https://github.com/containers/podman/blob/main/install.md#build-tags
- Podman REST API reference : https://docs.podman.io/en/latest/_static/api.html
- ADR 0002 (phased port plan) : [0002-port-docker-sdk-to-podman.md](0002-port-docker-sdk-to-podman.md)
