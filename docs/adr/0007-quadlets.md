# 0007 — Quadlets management (Phase 5d)

- **Date** : 2026-06-21
- **Status** : Accepted
- **Relates to** : [ADR 0006 — Podman-native features as optional capabilities](0006-podman-pods-and-capabilities.md)

## Context

Quadlets let users declare containers/pods/networks/volumes as systemd units via `*.container`, `*.pod`, etc. files under `~/.config/containers/systemd/` (rootless). On `systemctl --user daemon-reload`, Podman's generator turns each into a systemd service (`foo.container` → `foo.service`, `foo.pod` → `foo-pod.service`, …).

Unlike pods and generate-kube, **quadlets have no Podman binding** — they are managed through systemd. So this capability shells out to `systemctl --user`, not the REST API.

## Decision

Add an optional `runtime.QuadletManager` capability (same pattern as `PodRuntime`/`KubeGenerator`), implemented by the Podman backend over `systemctl --user`:

```go
type QuadletManager interface {
    ListQuadlets(ctx context.Context) ([]domain.Quadlet, error)
    StartQuadlet(ctx context.Context, unit string) error
    StopQuadlet(ctx context.Context, unit string) error
    RestartQuadlet(ctx context.Context, unit string) error
}
```

- **ListQuadlets** enumerates the quadlet source files in the user systemd dir, derives each generated unit name, and queries its state with `systemctl --user show <unit>`.
- **Start/Stop/Restart** map directly to `systemctl --user start|stop|restart <unit>` — these work on generated units.

### Why no enable/disable in v1

Empirically, a quadlet-generated unit reports `UnitFileState=generated`, and `systemctl --user enable <unit>` **fails** with *"Unit … is transient or generated"*. Autostart for a quadlet is controlled by the `[Install] WantedBy=` section **inside the source file**, not by `systemctl enable`. Toggling it therefore means editing the user's `*.container` file and reloading — a config-file mutation with its own safety/UX concerns. That is deliberately deferred to a follow-up; v1 ships the safe, reversible systemctl operations (list + lifecycle).

### Testability

The Podman runtime gains an injectable command runner so the quadlet logic (directory scan, unit-name derivation, `systemctl show` parsing) is unit-testable with a fake `systemctl` and a temp source dir, with no real systemd.

## Consequences

- **Pro**: users see their quadlet services and can start/stop/restart them from the TUI; no new dependency (systemd CLI only).
- **Con**: OS-coupled (rootless user systemd dir assumed, matching the default rootless Podman target); the unit-name derivation tracks Podman's generator conventions and may need updates if those change.
- **Deferred**: enable/disable (autostart) via `[Install]` editing; a GUI Quadlets panel (this PR is runtime-only, like 5a/5c-runtime).

## References

- Quadlet (`podman-systemd.unit`): https://docs.podman.io/en/latest/markdown/podman-systemd.unit.5.html
