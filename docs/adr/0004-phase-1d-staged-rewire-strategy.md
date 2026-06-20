# 0004 — Phase 1d: staged rewire strategy for pkg/commands

- **Date** : 2026-06-20
- **Status** : Accepted
- **Relates to** : [ADR 0003 — Runtime interface and domain types](0003-runtime-interface-and-domain-types.md) (refines its Phase 1d)

## Context

After delivering Phases 1a (domain types), 1b (interface), and 1c (Docker adapter), Phase 1d as originally sketched in ADR 0003 reads as: *"in pkg/commands, replace the Docker SDK calls with calls through runtime.ContainerRuntime"*. A closer look at the surface revealed three reasons that step is **not** a single tractable PR.

### 1. The struct fields are leaky

`pkg/commands.Container`, `Image`, `Network`, `Volume` all embed Docker SDK types as exported fields:

| Struct field | Current type (Docker SDK) | Domain equivalent |
|---|---|---|
| `Container.Container` | `container.Summary` | `domain.ContainerInfo` |
| `Container.Details` | `container.InspectResponse` | `domain.ContainerDetails` |
| `Container.Client` | `*client.Client` | not exposed; use the runtime through `DockerCommand` |
| `Image.Image` | `image.Summary` | `domain.ImageInfo` |
| `Network.Network` | `network.Inspect` | `domain.NetworkInfo` |
| `Volume.Volume` | `*volume.Volume` | `domain.VolumeInfo` |

`pkg/gui` reads ~30 distinct fields off these (inventoried in [`pkg/gui/containers_panel.go`](../../pkg/gui/containers_panel.go), [`pkg/gui/images_panel.go`](../../pkg/gui/images_panel.go), [`pkg/gui/presentation/containers.go`](../../pkg/gui/presentation/containers.go), [`pkg/gui/networks_panel.go`](../../pkg/gui/networks_panel.go), [`pkg/gui/volumes_panel.go`](../../pkg/gui/volumes_panel.go)). Changing one struct field's type cascades into the GUI immediately — there is no "1d without 1e" if the field types change.

### 2. Logs are multiplexed at the wire level

`pkg/gui/container_logs.go` calls `stdcopy.StdCopy` to demux the multiplexed Docker log stream. That is a Docker-specific protocol detail that the GUI should not know about. Either the runtime adapter performs the demux before handing the stream to callers, or we add a `domain.LogStream` type that abstracts it. **Decision (this ADR)**: the adapter demuxes; `ContainerLogs` already returns `io.ReadCloser` and the Docker implementation will switch to a `stdcopy.StdCopy`-driven pipe in 1d.

### 3. Attach is CLI-based, not SDK-based

`(*Container).Attach()` in upstream returns an `*exec.Cmd` running `docker attach`. That is an `exec` concern, not a runtime concern. **Decision (this ADR)**: `runtime.ContainerRuntime.AttachContainer` is kept as a future affordance for SDK-native attach, but pkg/commands keeps the CLI-based attach path during 1d (calling `docker` for the Docker backend, `podman attach` for the Podman backend in Phase 3+). A follow-up ADR will revisit once we have a use case requiring SDK-native attach.

## Decision

Split Phase 1d into **five micro-PRs**, each safely reviewable, build-passing and test-passing.

### 1d.0 — Connection setup foundation (this PR)

- Move `determineDockerHost`, `newDockerClient` and platform `defaultDockerHost` into `pkg/runtime/docker`:
  - `NewFromEnv() (*Runtime, error)` — one-call construction from environment.
  - `ResolveDockerHost() (string, error)` — exposed for callers that need the host string (SSH tunnel setup).
- `pkg/commands.NewDockerCommand` continues to use its own local copies for now — no functional change, no field added to `DockerCommand` yet.
- **Why this is the first step**: it gives 1d.1+ a single, importable entry point and surfaces no risk because the new functions have no caller.

### 1d.1 — DockerCommand acquires a runtime field

- Add `Runtime runtime.ContainerRuntime` field on `DockerCommand`.
- Construct it in `NewDockerCommand` using `docker.New(cli)` (reusing the already-built `*client.Client`).
- No method rewires. The runtime is unused; both fields coexist.
- **Why**: gives every subsequent micro-PR something to call through; risk = none.

### 1d.2 — Container method group rewire

- Rewrite `(*Container).Start/Stop/Restart/Pause/Unpause/Remove/Top/Inspect/RenderTop` to call through `c.DockerCommand.Runtime` instead of `c.Client`.
- `(*Container).Inspect()` now returns `domain.ContainerDetails` (breaking change for `pkg/gui/presentation/containers.go`). pkg/gui updated atomically in the same PR — field accesses are 1:1 because we designed `domain.ContainerDetails` after the GUI's inventory (ADR 0003).
- `(*DockerCommand).RefreshContainersAndServices`, `RefreshContainerDetails`, `GetContainers`, `PruneContainers` are rewired to use the runtime.
- `Container.Container container.Summary` is replaced by `Container.Info domain.ContainerInfo`. `Container.Details container.InspectResponse` is replaced by `Container.Details domain.ContainerDetails`. `Container.Client` is removed.
- pkg/gui field accesses (`.Container.State`, `.Container.Ports`, `.Details.Config.*`, `.Details.Mounts`, `.Details.NetworkSettings.*`) are updated in the same PR.
- **Risk**: medium. Mitigation: the existing `pkg/gui/sort_container_test.go` exercises the State / ID fields; we keep it passing. Visual smoke test (manual run against a real Docker socket) is the acceptance gate.

### 1d.3 — Image method group rewire

- Same pattern: `(*Image).Remove/RenderHistory` and `(*DockerCommand).RefreshImages/PruneImages` go through the runtime.
- `Image.Image image.Summary` becomes `Image.Info domain.ImageInfo`. `Image.Client` removed.
- pkg/gui (`pkg/gui/images_panel.go`, `pkg/gui/presentation/images.go`) updated atomically.
- **Risk**: low — smaller surface than containers.

### 1d.4 — Network + Volume method groups + events + stats

- Network: `(*Network).Remove`, `RefreshNetworks`, `PruneNetworks`.
- Volume: `(*Volume).Remove`, `RefreshVolumes`, `PruneVolumes`.
- Events: `(*DockerCommand).listenForEvents` (called from `pkg/gui/gui.go`) uses `runtime.Events`; drop the `docker/docker/api/types/events` import from `pkg/gui`.
- Stats: `(*DockerCommand).CreateClientStatMonitor` uses `runtime.ContainerStats`. The existing `ContainerStats`/`RecordedStats` types in `pkg/commands` get rewritten to accept `domain.Stats`.
- pkg/gui field accesses for networks and volumes updated atomically.
- **Risk**: low — networks/volumes have small field sets; stats path is well-isolated.

### 1d.5 — Cleanup

- Drop the `Docker *client.Client` field from `DockerCommand`.
- Drop the `docker/docker` import from `pkg/commands/docker.go`, `pkg/commands/container.go`, `pkg/commands/image.go`, `pkg/commands/network.go`, `pkg/commands/volume.go`.
- Drop the `docker/docker` imports from `pkg/gui/*`.
- After this PR, **only `pkg/runtime/docker` imports `github.com/docker/docker`**. Verifiable with one grep.

## Consequences

- **Pro**: every PR is small, reviewable, and ships a runnable build. The branch tip is always green.
- **Pro**: a manual GUI smoke test happens at every step, not just at the end. If a panel breaks, we know in which PR.
- **Pro**: each method group conversion is a self-contained mapping problem with no risk of cross-contamination.
- **Con**: the migration spans 5 PRs instead of 1. Slower elapsed time, but lower variance and a much safer per-step blast radius.
- **Con**: pkg/commands carries both `Client` and `Runtime` fields between 1d.1 and 1d.5 — temporary noise.

## Verification gate at every step

- `GOFLAGS=-mod=vendor go build ./...` — succeeds
- `GOFLAGS=-mod=vendor go vet ./...` — clean
- `gofmt -l pkg/commands pkg/gui pkg/runtime pkg/domain` — clean
- `GOFLAGS=-mod=vendor go test ./...` — green
- **Manual smoke**: `go run main.go` against a real Docker socket; navigate containers/images/networks/volumes panels; trigger logs, inspect, stop, restart, prune; confirm no panel renders differently from upstream.

That last point — the manual smoke — is the reason 1d.2 onwards needs a session where the operator can actually look at a terminal. It is not a CI gate; it is the gate that catches the bugs CI cannot see.
