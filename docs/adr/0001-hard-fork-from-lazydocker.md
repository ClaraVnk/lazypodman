# 0001 — Hard fork from lazydocker for Podman support

- **Date** : 2026-06-20
- **Status** : Accepted

## Context

We want a TUI to manage Podman containers, pods, images, volumes and networks. The reference experience in the Docker ecosystem is [lazydocker](https://github.com/jesseduffield/lazydocker) (MIT, written in Go). Podman exposes a Docker-compatible REST API, so lazydocker already *works* against Podman when `DOCKER_HOST` is pointed at the Podman socket — but the experience misses Podman-native concepts (pods, quadlets, rootless ergonomics) and requires manual socket setup.

Three options were considered:

1. **Thin wrapper** around lazydocker that auto-configures `DOCKER_HOST` and starts the user Podman socket.
2. **Hard fork** of lazydocker, replace the Docker SDK with the native Podman Go bindings, add Podman-specific features.
3. **From-scratch rewrite** with a modern TUI framework (Bubble Tea) and native Podman bindings.

## Decision

**Option 2 — hard fork of lazydocker.**

- License-compatible (lazydocker is MIT, we keep the original copyright notice and add our own).
- Inherits a mature, battle-tested TUI architecture and feature set.
- Allows us to gradually swap the runtime client (Docker SDK → `github.com/containers/podman/v5/pkg/bindings`) while keeping the rest of the codebase stable.
- Opens the door to Podman-only features (pod view, quadlet management, rootless socket discovery) that the upstream project will not adopt.

Upstream is tracked as the `upstream` git remote (`git@github.com:jesseduffield/lazydocker.git`) so we can cherry-pick TUI improvements during the porting phase.

## Consequences

- Initial commits carry lazydocker's git history — attribution is preserved by design.
- The codebase is in a transitional state until the Docker SDK → Podman bindings swap is complete. Build flags and dependencies are unchanged from upstream until then (`GOFLAGS=-mod=vendor`).
- The Go module path will be renamed from `github.com/jesseduffield/lazydocker` to `github.com/ClaraVnk/lazypodman` in a dedicated commit, once we are ready to diverge structurally from upstream (cherry-picking is easier while paths match).
- A new ADR will be written when the Docker SDK is replaced and again when Podman-specific features are added.

## References

- Upstream lazydocker: https://github.com/jesseduffield/lazydocker
- Podman Go bindings: https://pkg.go.dev/github.com/containers/podman/v5/pkg/bindings
- Podman REST API compatibility: https://docs.podman.io/en/latest/markdown/podman-system-service.1.html
