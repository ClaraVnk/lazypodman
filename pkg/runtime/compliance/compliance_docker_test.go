//go:build docker

package compliance

import (
	"github.com/ClaraVnk/lazypodman/pkg/runtime"
	dockerruntime "github.com/ClaraVnk/lazypodman/pkg/runtime/docker"
)

// Register the Docker backend in the dual-backend compliance suite. Compiled
// only with -tags docker, matching the rest of the Docker runtime; the default
// Podman-only build runs the suite against Podman alone.
func init() {
	complianceBackends["docker"] = func() (runtime.ContainerRuntime, error) {
		return dockerruntime.NewFromEnv()
	}
}
