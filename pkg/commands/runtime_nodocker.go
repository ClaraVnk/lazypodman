//go:build !docker

package commands

import (
	"fmt"
	"io"

	"github.com/ClaraVnk/lazypodman/pkg/runtime"
)

// newDockerBackend reports that the Docker backend was not compiled into this
// binary. lazypodman is Podman-first; the Docker SDK (and its CVEs) is left
// out of the default build. Rebuild with -tags docker to enable the Docker
// runtime (runtime: docker).
func newDockerBackend(_ *OSCommand) (runtime.ContainerRuntime, []io.Closer, error) {
	return nil, nil, fmt.Errorf("the Docker backend is not built into this binary; rebuild with -tags docker to use the Docker runtime")
}
