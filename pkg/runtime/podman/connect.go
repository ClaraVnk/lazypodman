package podman

import (
	"os"
	"path"
	"strings"
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
		// path.Join (not filepath.Join) so the socket URI always uses
		// forward slashes, including on Windows.
		return "unix://" + path.Join(xdg, "podman", "podman.sock")
	}
	return "unix:///run/podman/podman.sock"
}

// NewFromEnv returns a Runtime targeting the Podman service the
// environment points at. It does not connect yet — the socket is dialed
// lazily on first use (see Runtime.client) so a down engine never fails
// construction. The error return is kept for signature symmetry with the
// Docker constructor and future eager-validation needs.
func NewFromEnv() (*Runtime, error) {
	return &Runtime{uri: ResolveSocketURI()}, nil
}
