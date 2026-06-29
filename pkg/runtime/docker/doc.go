//go:build docker

// Package docker implements runtime.ContainerRuntime against the Docker
// Engine API via the official Docker Go SDK
// (github.com/docker/docker/client).
//
// Construction is intentionally narrow: New takes an already-built
// *client.Client and trusts the caller to have negotiated the API
// version, resolved DOCKER_HOST, and set up any SSH tunnel — newDockerBackend
// in pkg/commands orchestrates that before calling NewWithHost.
package docker
