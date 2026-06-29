package docker

import (
	dockerclient "github.com/docker/docker/client"

	"github.com/ClaraVnk/lazypodman/pkg/runtime"
)

// Runtime is the Docker implementation of runtime.ContainerRuntime.
// It is safe for concurrent use; the underlying Docker client already is.
type Runtime struct {
	cli *dockerclient.Client
}

// Compile-time check that Runtime satisfies the interface.
var _ runtime.ContainerRuntime = (*Runtime)(nil)

// New wraps an already-built Docker client. The caller is responsible
// for DOCKER_HOST resolution, API version negotiation and any SSH
// tunnel setup. Pass a client built with
// client.NewClientWithOpts(client.WithAPIVersionNegotiation(), ...).
func New(cli *dockerclient.Client) *Runtime {
	return &Runtime{cli: cli}
}

// Close releases the underlying Docker client.
func (r *Runtime) Close() error {
	if r.cli == nil {
		return nil
	}
	err := r.cli.Close()
	r.cli = nil
	return mapErr("close", err)
}
