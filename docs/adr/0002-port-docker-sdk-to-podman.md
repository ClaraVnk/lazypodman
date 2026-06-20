# 0002 — Port from Docker SDK to Podman bindings

- **Date** : 2026-06-20
- **Status** : Proposed
- **Relates to** : [ADR 0001 — Hard fork from lazydocker](0001-hard-fork-from-lazydocker.md)

## Context

Inherited from upstream lazydocker, the codebase talks directly to the Docker daemon via the official Docker Go SDK (`github.com/docker/docker`). Concretely, the surface used is:

| Package | Role |
| --- | --- |
| `docker/docker/client` | Daemon client (connect, list, inspect, exec, attach) |
| `docker/docker/api/types/container` | Container payloads |
| `docker/docker/api/types/image` | Image payloads |
| `docker/docker/api/types/network` | Network payloads |
| `docker/docker/api/types/volume` | Volume payloads |
| `docker/docker/api/types/events` | Event stream |
| `docker/docker/api/types/filters` | Server-side filtering |
| `docker/docker/pkg/stdcopy` | Demux of multiplexed stdout/stderr |

These imports are concentrated in **13 files under `pkg/commands/`**. The GUI layer (`pkg/gui/`) consumes the resulting domain objects via the `commands` package — so the GUI is *almost* runtime-agnostic by design.

Podman exposes a Docker-compatible REST socket, so `lazypodman` already works against Podman today by pointing `DOCKER_HOST` at `unix:///run/user/$UID/podman/podman.sock`. But the Docker-compat layer:

- does not expose Podman-native concepts (pods, quadlets, generate kube, healthchecks the Podman way),
- has historical gaps (notably exec/attach edge cases on Podman),
- couples us to the Docker types as a public contract forever.

The endgame is to talk to Podman via its native Go bindings: [`github.com/containers/podman/v5/pkg/bindings`](https://pkg.go.dev/github.com/containers/podman/v5/pkg/bindings).

## Decision

Adopt a **strangler-fig** migration in 6 phases. Each phase is independently shippable, reviewable, and reversible. No big-bang rewrite.

### Phase 1 — Runtime abstraction

Introduce an internal interface in `pkg/commands` (provisional name `ContainerRuntime`) that exposes everything the GUI currently needs (list/inspect/exec/attach/logs/events/stats/prune for containers, images, networks, volumes). The interface deals in **lazypodman-owned domain types**, not Docker SDK types.

Goal: the GUI imports `pkg/commands` only — never `docker/docker/*`.

Risk: the Docker SDK types are rich; we may need a careful mapping layer. Mitigation: start with the minimum surface the TUI actually renders, grow as needed.

### Phase 2 — Docker backend behind the interface

Implement the abstraction with the existing Docker SDK code. **Zero behavior change.** Tests pass identically. This phase is purely a refactor that proves the interface is sufficient and lets us swap implementations in Phase 3.

### Phase 3 — Podman backend

Add a second implementation using `github.com/containers/podman/v5/pkg/bindings`. Selectable via a config field (`runtime: docker | podman`) and/or env (`LAZYPODMAN_RUNTIME`). Default stays `docker` for this phase — Podman backend is opt-in until parity is reached.

Side effects:
- New dependency (`pkg/bindings` and its tree). Vendor and audit (`govulncheck`).
- Native auto-discovery of the user Podman socket (`$XDG_RUNTIME_DIR/podman/podman.sock`).

### Phase 4 — Default to Podman

When the Podman backend reaches feature parity (verified by a shared compliance test suite running against both backends), flip the default to `podman`. Docker backend stays available as a fallback (`runtime: docker`).

### Phase 5 — Podman-specific features

Pods view (list/inspect/start/stop pods), quadlet support (`systemctl --user` integration to enable/disable quadlets), generate-kube export, rootless ergonomics (socket bring-up if missing).

### Phase 6 — Drop the Docker backend, rename module path

Once the Podman backend has been the default for at least one release cycle and no significant Docker-only usage is reported:

- Remove the Docker backend code and the `docker/docker` dependency.
- Rename the Go module from `github.com/jesseduffield/lazydocker` to `github.com/ClaraVnk/lazypodman` (single mechanical commit, mass `gofmt`-able).
- Rename the binary, scripts, packaging assets (`Dockerfile` → `Containerfile`, `docker-compose.yml` → archived).

Until Phase 6 lands, the module path stays as upstream to keep `git cherry-pick upstream/master -- ...` cheap.

## Consequences

- **Pro**: incremental, low-risk, each phase ships value. Upstream lazydocker fixes remain easy to absorb until Phase 6.
- **Pro**: forced separation of concerns (runtime ↔ TUI), valuable even outside the Podman context.
- **Con**: temporary code complexity (two backends side-by-side during phases 3-5).
- **Con**: dependency footprint grows (Podman bindings are not small).

## Open questions

- Remote Podman (over SSH or TCP) — in scope or out for v1 ?
- Buildah / Skopeo integration — separate ADR ?
- Compose support — `podman compose` shim, native, or out of scope ?

These will be answered in follow-up ADRs as they come up.

## References

- Podman Go bindings tutorial : https://github.com/containers/podman/blob/main/pkg/bindings/README.md
- Strangler-fig pattern : https://martinfowler.com/bliki/StranglerFigApplication.html
- Upstream Docker SDK : https://pkg.go.dev/github.com/docker/docker
