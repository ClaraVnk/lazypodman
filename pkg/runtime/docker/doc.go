// Package docker implements runtime.ContainerRuntime against the Docker
// Engine API via the official Docker Go SDK
// (github.com/docker/docker/client).
//
// Construction is intentionally narrow: New takes an already-built
// *client.Client and trusts the caller to have negotiated the API
// version, resolved DOCKER_HOST, and set up any SSH tunnel. Centralising
// connection logic stays in pkg/commands until Phase 1d.
package docker
