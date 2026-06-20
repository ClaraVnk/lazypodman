# 0003 — Runtime interface and domain types

- **Date** : 2026-06-20
- **Status** : Proposed
- **Relates to** : [ADR 0002 — Port from Docker SDK to Podman bindings](0002-port-docker-sdk-to-podman.md) (this is the detailed design for Phase 1)

## Context

Inventory of the upstream codebase as of `2493d37`:

### Public surface of `pkg/commands`

The package exposes:

- A **command** entry point: `DockerCommand` (connection, refresh* operations, prune*, log streaming, compose helpers).
- Five **domain structs**: `Container`, `Image`, `Network`, `Volume`, `Service`.
- Statistics types: `ContainerStats`, `DerivedStats`, `RecordedStats`.
- An `OSCommand` helper (OS-level wrapping — orthogonal to the runtime, **not affected by this ADR**).
- Errors: `ComplexError`, `HasErrorCode`, `WrapError`.
- A `Platform` struct (build metadata — orthogonal).
- An already-existing `LimitedDockerCommand` interface (small, used to break a cycle internally).

### Where the Docker SDK leaks

**Method signatures** that expose Docker SDK types in the public API:

| Method | Leaked type |
|---|---|
| `(*Container).Inspect()` | `container.InspectResponse` |
| `(*Container).Remove(opts)` | `container.RemoveOptions` |
| `(*Container).Top(ctx)` | `container.TopResponse` |
| `(*Image).Remove(opts)` | `image.RemoveOptions` |
| `(*Service).Remove(opts)` | `container.RemoveOptions` |

**Struct fields** that expose Docker SDK types (worse — read by the GUI directly):

| Struct | Field | Type |
|---|---|---|
| `Container` | `Container` | `container.Summary` |
| `Container` | `Details` | `container.InspectResponse` |
| `Container` | `Client` | `*client.Client` |
| `Image` | `Image` | `image.Summary` |
| `Image` | `Client` | `*client.Client` |
| `Network` | `Network` | `network.Inspect` |
| `Volume` | `Volume` | `*volume.Volume` |

**`pkg/gui` directly imports** `docker/docker/api/types/{container,events,image}` and `docker/docker/pkg/stdcopy`. ~30 distinct Docker SDK fields are read across `pkg/gui` (rendering panels for containers, images, networks, volumes).

## Decision

Introduce two new internal packages and keep `pkg/commands` thin (composition only). The split is deliberate: domain types must be importable **without** the runtime, and the runtime interface depends on the domain — never the reverse.

### Package layout

```
pkg/
  domain/         # NEW — lazypodman-owned value types, zero external deps
    container.go  # ContainerInfo, ContainerDetails, ContainerState, Port, Mount...
    image.go      # ImageInfo
    network.go    # NetworkInfo
    volume.go     # VolumeInfo
    event.go      # Event (replaces docker/docker/api/types/events)
    stats.go      # Stats (replaces the inherited stats types in pkg/commands)
  runtime/        # NEW — the abstraction
    runtime.go    # ContainerRuntime interface
    errors.go     # Sentinel errors (ErrNotFound, ErrConflict...)
    docker/       # FUTURE (Phase 2) — Docker backend implementing ContainerRuntime
    podman/       # FUTURE (Phase 3) — Podman backend implementing ContainerRuntime
  commands/       # EXISTING — slimmed: no more SDK types in its public API
```

### Domain types (sketch)

```go
// pkg/domain/container.go

type ContainerInfo struct {
    ID          string
    Name        string
    Image       string
    ImageID     string
    Command     string
    Created     time.Time
    State       ContainerState
    Status      string
    Ports       []Port
    Labels      map[string]string
    SizeRw      int64
    SizeRootFs  int64
}

type ContainerState string

const (
    ContainerStateCreated    ContainerState = "created"
    ContainerStateRunning    ContainerState = "running"
    ContainerStatePaused     ContainerState = "paused"
    ContainerStateRestarting ContainerState = "restarting"
    ContainerStateRemoving   ContainerState = "removing"
    ContainerStateExited     ContainerState = "exited"
    ContainerStateDead       ContainerState = "dead"
)

type Port struct {
    Host          string
    PublicPort    uint16
    ContainerPort uint16
    Protocol      string // "tcp" | "udp" | "sctp"
}

type ContainerDetails struct {
    ContainerInfo
    Args            []string
    Path            string
    Mounts          []Mount
    NetworkSettings NetworkSettings
    Health          *Health
    Config          Config
}
```

The exact fields are derived from the call-graph inventory — only what the GUI actually reads (or what an obvious near-future feature will read) is included. We do **not** mirror Docker's full schema.

### Runtime interface (sketch)

```go
// pkg/runtime/runtime.go

type ContainerRuntime interface {
    // Connection lifecycle
    Close() error

    // Containers
    ListContainers(ctx context.Context) ([]domain.ContainerInfo, error)
    InspectContainer(ctx context.Context, id string) (domain.ContainerDetails, error)
    StartContainer(ctx context.Context, id string) error
    StopContainer(ctx context.Context, id string, timeout *time.Duration) error
    RestartContainer(ctx context.Context, id string, timeout *time.Duration) error
    PauseContainer(ctx context.Context, id string) error
    UnpauseContainer(ctx context.Context, id string) error
    RemoveContainer(ctx context.Context, id string, opts RemoveContainerOptions) error
    ContainerLogs(ctx context.Context, id string, opts LogOptions) (io.ReadCloser, error)
    ContainerTop(ctx context.Context, id string) (domain.TopOutput, error)
    ContainerStats(ctx context.Context, id string) (<-chan domain.Stats, error)
    AttachContainer(ctx context.Context, id string, opts AttachOptions) (domain.AttachStream, error)
    PruneContainers(ctx context.Context) (domain.PruneReport, error)

    // Images
    ListImages(ctx context.Context) ([]domain.ImageInfo, error)
    RemoveImage(ctx context.Context, id string, opts RemoveImageOptions) error
    ImageHistory(ctx context.Context, id string) ([]domain.ImageHistoryItem, error)
    PruneImages(ctx context.Context) (domain.PruneReport, error)

    // Networks
    ListNetworks(ctx context.Context) ([]domain.NetworkInfo, error)
    RemoveNetwork(ctx context.Context, id string) error
    PruneNetworks(ctx context.Context) (domain.PruneReport, error)

    // Volumes
    ListVolumes(ctx context.Context) ([]domain.VolumeInfo, error)
    RemoveVolume(ctx context.Context, name string, force bool) error
    PruneVolumes(ctx context.Context) (domain.PruneReport, error)

    // Events
    Events(ctx context.Context, since time.Time) (<-chan domain.Event, <-chan error)
}

type RemoveContainerOptions struct {
    Force         bool
    RemoveVolumes bool
    Links         bool
}

type LogOptions struct {
    Follow     bool
    Tail       string // "all" | "<N>"
    Since      time.Time
    Until      time.Time
    Timestamps bool
}
```

### Migration sub-phases

Phase 1 is itself broken down into reviewable PRs:

1. **1a** — Create `pkg/domain` with the value types and unit tests. No callers yet. *(this PR)*
2. **1b** — Create `pkg/runtime` with the interface + sentinel errors. No implementation yet. *(this PR or next)*
3. **1c** — Implement the runtime against the existing Docker SDK in `pkg/runtime/docker`. Pure adapter — `pkg/commands` keeps working in parallel during the transition.
4. **1d** — In `pkg/commands`, replace the Docker SDK calls with calls through `runtime.ContainerRuntime`. Domain types replace the SDK types in struct fields. The `*client.Client` field disappears. Public method signatures stop leaking SDK types.
5. **1e** — In `pkg/gui`, replace direct reads of `container.Summary`/`image.Summary`/etc. with reads of `domain.ContainerInfo`/`domain.ImageInfo`. Drop the `docker/docker/*` imports from `pkg/gui`.

After 1e, **only `pkg/runtime/docker` imports `github.com/docker/docker`**. That is the gate that unlocks Phases 2 → 6.

## Consequences

- **Pro**: clean separation; `pkg/gui` becomes runtime-agnostic; testing the GUI no longer requires a Docker mock — a fake `ContainerRuntime` is enough.
- **Pro**: each sub-phase is a small reviewable PR. No PR breaks the build or removes functionality.
- **Pro**: the cost of adding the Podman backend (Phase 3) collapses to "implement the same interface against `pkg/bindings`" — no GUI changes needed.
- **Con**: temporary code duplication during 1c-1d (the old `pkg/commands` and the new `pkg/runtime/docker` exist side by side).
- **Con**: the inventory of "fields actually read by the GUI" must stay current — any GUI code reading a new SDK field during the transition will need a domain-type update.

## Open questions

- Should `OSCommand` move out of `pkg/commands` into its own `pkg/osexec` package? It is logically unrelated to container runtimes. **Recommendation**: yes, in a separate cleanup PR — not gated on this ADR.
- `Project` and `Service` (compose grouping) are currently in `pkg/commands`. Podman has `podman compose` but the semantics differ. **Recommendation**: keep them in `pkg/commands` for now; revisit when adding compose support to the Podman backend in Phase 5.
- Should domain types implement `fmt.Stringer` and rendering helpers, or stay as pure data? **Recommendation**: pure data — rendering belongs to `pkg/gui/presentation`.

## References

- Inventory: `git log -p` of this commit (the `Found in` columns above are reproducible from the listed greps).
- Strangler-fig pattern: see [ADR 0002](0002-port-docker-sdk-to-podman.md#references).
