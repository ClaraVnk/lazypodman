//go:build docker

package docker

import (
	"os"
	"testing"

	dockerclient "github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
)

// TestNewWithHostVersionNegotiation verifies that NewWithHost allows API
// version negotiation even when DOCKER_API_VERSION is set.
//
// Regression test for https://github.com/ClaraVnk/lazypodman/issues/715
// where users got "client version 1.25 is too old" errors because FromEnv()
// includes WithVersionFromEnv(), which sets manualOverride=true and prevents
// API version negotiation. The construction path used to live in
// pkg/commands.newDockerClient; it moved here in Phase 1d.5.
func TestNewWithHostVersionNegotiation(t *testing.T) {
	originalAPIVersion := os.Getenv("DOCKER_API_VERSION")
	defer func() {
		if originalAPIVersion == "" {
			os.Unsetenv("DOCKER_API_VERSION")
		} else {
			os.Setenv("DOCKER_API_VERSION", originalAPIVersion)
		}
	}()

	// An old version that would trigger "client version 1.25 is too old"
	// errors if negotiation were disabled.
	os.Setenv("DOCKER_API_VERSION", "1.25")

	t.Run("FromEnv locks version preventing negotiation", func(t *testing.T) {
		// Demonstrates the behavior we deliberately avoid: FromEnv locks the
		// client version to the env var and disables negotiation.
		cli, err := dockerclient.NewClientWithOpts(
			dockerclient.FromEnv,
			dockerclient.WithAPIVersionNegotiation(),
		)
		assert.NoError(t, err)
		defer cli.Close()

		assert.Equal(t, "1.25", cli.ClientVersion())
	})

	t.Run("NewWithHost allows version negotiation", func(t *testing.T) {
		// DefaultDockerHost is cross-platform (unix socket / named pipe).
		rt, err := NewWithHost(dockerclient.DefaultDockerHost)
		assert.NoError(t, err)
		defer rt.Close()

		// The version is not locked to the env var; it negotiates with the
		// server on first request.
		assert.NotEqual(t, "1.25", rt.cli.ClientVersion(),
			"client version should not be locked to DOCKER_API_VERSION env var")
	})
}
