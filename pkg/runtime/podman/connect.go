package podman

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/containers/podman/v5/pkg/bindings"

	"github.com/jesseduffield/lazydocker/pkg/runtime"
)

const containerHostEnvKey = "CONTAINER_HOST"

// ResolveSocketURI reproduces Podman's socket resolution, in decreasing
// precedence:
//   - $CONTAINER_HOST when set;
//   - the rootless default unix://$XDG_RUNTIME_DIR/podman/podman.sock;
//   - the rootful default unix:///run/podman/podman.sock.
func ResolveSocketURI() string {
	if h := strings.TrimSpace(os.Getenv(containerHostEnvKey)); h != "" {
		return h
	}
	if xdg := strings.TrimSpace(os.Getenv("XDG_RUNTIME_DIR")); xdg != "" && os.Geteuid() != 0 {
		return "unix://" + filepath.Join(xdg, "podman", "podman.sock")
	}
	return "unix:///run/podman/podman.sock"
}

// NewFromEnv connects to the Podman service the environment points at and
// returns a ready Runtime. A connection failure is mapped to
// runtime.ErrUnavailable so callers can react with errors.Is.
func NewFromEnv() (*Runtime, error) {
	conn, err := bindings.NewConnection(context.Background(), ResolveSocketURI())
	if err != nil {
		return nil, fmt.Errorf("podman: connect: %w: %s", runtime.ErrUnavailable, err.Error())
	}
	return &Runtime{conn: conn}, nil
}
