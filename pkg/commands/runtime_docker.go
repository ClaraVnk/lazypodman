//go:build docker

package commands

import (
	"io"
	ogLog "log"
	"os"
	"strings"

	"github.com/ClaraVnk/lazypodman/pkg/commands/ssh"
	"github.com/ClaraVnk/lazypodman/pkg/runtime"
	dockerruntime "github.com/ClaraVnk/lazypodman/pkg/runtime/docker"
)

// dockerHostEnvKey is the standard Docker host environment variable, read and
// written here to hand the resolved host to the SSH tunnel helper.
const dockerHostEnvKey = "DOCKER_HOST"

// newDockerBackend builds the Docker runtime, resolving the host (honouring an
// SSH tunnel). It is compiled only with -tags docker; the default Podman-only
// build uses the stub in runtime_nodocker.go, which keeps the Docker SDK (and
// its CVEs) out of the default binary.
func newDockerBackend(osCommand *OSCommand) (runtime.ContainerRuntime, []io.Closer, error) {
	dockerHost, err := dockerruntime.ResolveDockerHost()
	if err != nil {
		ogLog.Printf("> could not determine host %v", err)
	}

	// Inject the resolved host into the environment so HandleSSHDockerHost can
	// create a local unix socket tunnelled over SSH to the specified host.
	if strings.HasPrefix(dockerHost, "ssh://") {
		os.Setenv(dockerHostEnvKey, dockerHost)
	}

	tunnelCloser, err := ssh.NewSSHHandler(osCommand).HandleSSHDockerHost()
	if err != nil {
		ogLog.Fatal(err)
	}

	// HandleSSHDockerHost may have overridden DOCKER_HOST in the environment.
	if h := os.Getenv(dockerHostEnvKey); h != "" {
		dockerHost = h
	}

	rt, err := dockerruntime.NewWithHost(dockerHost)
	if err != nil {
		ogLog.Fatal(err)
	}
	return rt, []io.Closer{tunnelCloser, rt}, nil
}
