# 0006 — Podman-native features as optional capability interfaces (Phase 5)

- **Date** : 2026-06-21
- **Status** : Accepted
- **Relates to** : [ADR 0002 — Port from Docker SDK to Podman bindings](0002-port-docker-sdk-to-podman.md) (implements its Phase 5), [ADR 0005 — Podman native backend](0005-podman-native-backend.md)

## Context

Phases 1–4 delivered a runtime abstraction with two backends and made Podman the default. Phase 5 adds the features that justify the fork: **pods**, **quadlets**, and **generate kube**. These are Podman-native concepts with no Docker equivalent.

The question is where they live. Putting `ListPods`, `GenerateKube`, etc. on the core `runtime.ContainerRuntime` interface would force the Docker backend to implement a pile of methods that can only ever return `ErrUnsupported`, and would let GUI code call pod methods on a Docker runtime that can never satisfy them.

## Decision

Model Podman-native features as **optional capability interfaces**, separate from `ContainerRuntime`:

```go
// pkg/runtime
type PodRuntime interface {
    ListPods(ctx context.Context) ([]domain.PodInfo, error)
    StartPod(ctx context.Context, id string) error
    StopPod(ctx context.Context, id string, timeout *time.Duration) error
    RestartPod(ctx context.Context, id string) error
    RemovePod(ctx context.Context, id string, force bool) error
    PrunePods(ctx context.Context) (domain.PruneReport, error)
}

type KubeGenerator interface {
    GenerateKube(ctx context.Context, ids []string) ([]byte, error)
}
```

- The **Podman** backend implements these; the **Docker** backend does not (it simply lacks the methods).
- Callers **type-assert** for the capability: `if pr, ok := rt.(runtime.PodRuntime); ok { … }`. The GUI shows the Pods panel only when the active backend advertises `PodRuntime`, so on Docker the panel is absent rather than broken.
- New domain types (`domain.PodInfo`, `domain.PodContainer`, `domain.PodStatus`) follow the same owned-types rule as the rest of the port — no Podman SDK types cross the package boundary.

This keeps `ContainerRuntime` honest (every method works on every backend) and makes capabilities discoverable and additive.

## Staged sub-PRs

- **5a — Pods (runtime)** *(this increment)*: `domain.Pod*` types, the `PodRuntime` interface, and the Podman implementation (`ListPods` + lifecycle + prune). No GUI yet, so no user-visible change; the capability is dormant until 5b.
- **5b — Pods panel (GUI)**: a Pods side panel (list/inspect/start/stop/remove), shown only when the backend implements `PodRuntime`. Containers gain a "pod" column/grouping.
- **5c — generate kube**: `KubeGenerator` + a GUI action to export a pod/container to a Kubernetes YAML.
- **5d — Quadlets**: enable/disable quadlet units via `systemctl --user` (a CLI/exec concern, not the bindings) behind a `QuadletManager` capability. Separate design notes when picked up.

## Consequences

- **Pro**: Docker stays free of dead pod stubs; the interface contract that "every method works on every backend" holds.
- **Pro**: capabilities are discoverable at runtime and the GUI adapts (panel present iff supported).
- **Con**: GUI code carries type-assertions at the capability boundary — acceptable and localized.
- **Con**: a future second pod-capable backend would re-implement `PodRuntime`; fine, that is the point of the interface.

## References

- Podman pods bindings: `github.com/containers/podman/v5/pkg/bindings/pods`
- `podman generate kube`: https://docs.podman.io/en/latest/markdown/podman-kube-generate.1.html
- Quadlets: https://docs.podman.io/en/latest/markdown/podman-systemd.unit.5.html
