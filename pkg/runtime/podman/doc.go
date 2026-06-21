// Package podman is the native Podman implementation of
// runtime.ContainerRuntime, built on Podman's Go bindings
// (github.com/containers/podman/v5/pkg/bindings).
//
// The container method group is implemented (Phase 3b); the image,
// network, volume, event, stats and log groups land in Phases 3c–3e and
// report runtime.ErrUnsupported until then. Backend selection lives in
// pkg/commands (config `runtime:` / LAZYPODMAN_RUNTIME). See
// docs/adr/0005-podman-native-backend.md.
//
// The package builds CGO-free with the standard Podman client build tags
// (containers_image_openpgp, exclude_graphdriver_btrfs,
// exclude_graphdriver_devicemapper, remote).
package podman
