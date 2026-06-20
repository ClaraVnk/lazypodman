# lazypodman

A lazier way to manage Podman containers — a hard fork of [jesseduffield/lazydocker](https://github.com/jesseduffield/lazydocker), adapted for Podman.

> **Status: early fork.** The code currently still targets the Docker SDK. The plan is to incrementally swap it for the native Podman Go bindings (`github.com/containers/podman/v5/pkg/bindings`) and add Podman-specific features (pods, quadlets, rootless ergonomics).

## What

A terminal UI to view and operate containers, pods, images, volumes and networks managed by Podman, without leaving the terminal. Same spirit as [lazydocker](https://github.com/jesseduffield/lazydocker), with Podman as the runtime.

## Run

```sh
go run main.go
```

For now the upstream lazydocker build flags apply (the codebase still uses the Docker SDK):

```sh
GOFLAGS=-mod=vendor go run main.go
```

## Test

```sh
GOFLAGS=-mod=vendor go test ./...
```

## Deploy

Pre-built binaries and release artifacts: not yet available. Builds are local-only during the porting phase.

## Architecture

See [`docs/adr/`](docs/adr/) for the design decisions behind this fork:

- [0001 — Hard fork from lazydocker for Podman support](docs/adr/0001-hard-fork-from-lazydocker.md)

Until the port is done, the architecture is identical to upstream lazydocker. Refer to the upstream docs in [`docs/`](docs/) for the current internal layout.

## Credits

This project is a hard fork of [jesseduffield/lazydocker](https://github.com/jesseduffield/lazydocker) (MIT). All credit for the original design, TUI architecture and most of the codebase goes to Jesse Duffield and the lazydocker contributors. See [`LICENSE`](LICENSE).

## License

MIT — see [`LICENSE`](LICENSE).
