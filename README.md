<div align="center">

# lazypodman

**A lazier way to manage Podman from your terminal.**

A fast, keyboard-driven terminal UI for Podman — containers, pods, images, volumes, networks, logs, stats and more, without leaving the shell.

[![CI](https://github.com/ClaraVnk/lazypodman/actions/workflows/ci.yml/badge.svg)](https://github.com/ClaraVnk/lazypodman/actions/workflows/ci.yml)
[![Security](https://github.com/ClaraVnk/lazypodman/actions/workflows/security.yml/badge.svg)](https://github.com/ClaraVnk/lazypodman/actions/workflows/security.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25%2B-00ADD8?logo=go&logoColor=white)](go.mod)
[![Podman](https://img.shields.io/badge/Podman-native-892CA0?logo=podman&logoColor=white)](https://podman.io)
[![PRs welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](#contributing)

</div>

---

## What is this?

`lazypodman` is a hard fork of the excellent [jesseduffield/lazydocker](https://github.com/jesseduffield/lazydocker), re-engineered to talk to **Podman natively** via its Go bindings (`github.com/containers/podman/v5/pkg/bindings`) — not through the Docker-compatibility shim. It keeps lazydocker's snappy TUI and adds the features that make Podman worth using on its own terms: **pods**, **`generate kube`**, and **quadlets**.

Podman is the default backend. Docker remains available as a fallback for mixed environments.

## Highlights

- 🦭 **Native Podman backend** — speaks the Podman API directly (rootless socket auto-discovered); no `DOCKER_HOST` shim required.
- 📦 **Pods panel** — list, inspect, remove and prune pods; export any pod to Kubernetes YAML (`podman generate kube`) from a dedicated tab.
- ⚙️ **Quadlets** — see your `systemd` quadlet units and start/stop/restart them.
- 🐳 **Docker fallback** — set `runtime: docker` (or `LAZYPODMAN_RUNTIME=docker`) to drive a Docker daemon instead.
- 🧰 **Everything you expect** — containers, images, volumes, networks, logs, stats, exec, inspect, prune, custom & bulk commands.
- 🪶 **Single static binary** — builds CGO-free; nothing to install on the host but the engine.

Podman-only features (pods, kube export, quadlets) appear in the UI **only when the active backend supports them** — they are exposed through optional capability interfaces and stay hidden on Docker.

## Requirements

- **Go 1.25+** (to build from source)
- **Podman 5.x** with the user API socket running, e.g.:
  ```sh
  systemctl --user enable --now podman.socket
  ```
  lazypodman auto-discovers `$XDG_RUNTIME_DIR/podman/podman.sock` (rootless) or `/run/podman/podman.sock` (rootful), and honours `CONTAINER_HOST`.
- *(optional)* a Docker daemon, if you use the `docker` backend.

## Install

### Released binary (Linux/macOS)

Once a release is published, install or update the latest binary into `~/.local/bin`:

```sh
curl -fsSL https://raw.githubusercontent.com/ClaraVnk/lazypodman/main/scripts/install_update_linux.sh | bash
```

Pre-built archives for Linux, macOS and Windows (amd64/arm64) are attached to each [GitHub Release](https://github.com/ClaraVnk/lazypodman/releases).

### Build from source

```sh
git clone https://github.com/ClaraVnk/lazypodman.git
cd lazypodman

CGO_ENABLED=0 go build -mod=vendor \
  -tags 'containers_image_openpgp exclude_graphdriver_btrfs exclude_graphdriver_devicemapper remote' \
  -o lazypodman .

./lazypodman
```

> The build tags are **required**: they build the Podman bindings tree without the storage graph drivers that need C libraries, keeping the binary CGO-free and portable.

## Usage

```sh
./lazypodman                             # native Podman (default)
LAZYPODMAN_RUNTIME=docker ./lazypodman   # Docker backend
```

Navigate panels with the number keys (`[1]`–`[7]`) or the arrows; press `x` for the menu and `b` for bulk commands. Full keybinding cheatsheets live in [`docs/keybindings/`](docs/keybindings/).

### Backend selection

| Source | Value | Notes |
|---|---|---|
| Env var `LAZYPODMAN_RUNTIME` | `podman` \| `docker` | highest precedence |
| Config `runtime:` | `podman` \| `docker` | in your lazypodman config |
| *(default)* | `podman` | |

## Architecture

The TUI never talks to an engine SDK directly — it goes through a `runtime.ContainerRuntime` interface that deals in lazypodman-owned domain types. Two backends implement it (`pkg/runtime/podman`, `pkg/runtime/docker`), and Podman-only features are layered on as optional capability interfaces (`PodRuntime`, `KubeGenerator`, `QuadletManager`) that callers discover via type assertions.

The design decisions are recorded as ADRs in [`docs/adr/`](docs/adr/):

| ADR | Topic |
|---|---|
| [0001](docs/adr/0001-hard-fork-from-lazydocker.md) | Hard fork from lazydocker |
| [0002](docs/adr/0002-port-docker-sdk-to-podman.md) | Phased port: Docker SDK → Podman bindings |
| [0003](docs/adr/0003-runtime-interface-and-domain-types.md) | Runtime interface & domain types |
| [0004](docs/adr/0004-phase-1d-staged-rewire-strategy.md) | Staged rewire of `pkg/commands` |
| [0005](docs/adr/0005-podman-native-backend.md) | Native Podman backend |
| [0006](docs/adr/0006-podman-pods-and-capabilities.md) | Pods & optional capability interfaces |
| [0007](docs/adr/0007-quadlets.md) | Quadlets management |

## Project status

The port is functionally complete: the runtime abstraction, the native Podman backend, the default flip to Podman, and the Podman-native features (pods + kube export in the UI, quadlets at the runtime layer) are all in. A dual-backend compliance suite gates parity in CI.

Roadmap odds and ends: a Quadlets UI panel, quadlet enable/disable (autostart), published release binaries, and eventually renaming the Go module once Podman has been the default for a release cycle.

## Contributing

Issues and PRs are welcome. Before opening a PR:

```sh
TAGS='containers_image_openpgp exclude_graphdriver_btrfs exclude_graphdriver_devicemapper remote'
GOFLAGS="-mod=vendor -tags=$TAGS" go build ./...
GOFLAGS="-mod=vendor -tags=$TAGS" go test ./...
bash ./test.sh        # race + coverage
```

CI runs lint (golangci-lint v2), format, tests on Linux + Windows, a build matrix, `govulncheck`, `gitleaks`, and the dual-backend compliance suite. Keep commits atomic and use [Conventional Commits](https://www.conventionalcommits.org/).

## Credits

A hard fork of [jesseduffield/lazydocker](https://github.com/jesseduffield/lazydocker) (MIT). All credit for the original TUI, design and the bulk of the codebase goes to Jesse Duffield and the lazydocker contributors. 💛

### Contributors

Thanks to everyone who has contributed to lazypodman:

- [@kallioli](https://github.com/kallioli) — Podman runtime fixes and test coverage

## License

[MIT](LICENSE) © 2026 Clara Vanacker — and the original lazydocker authors.
