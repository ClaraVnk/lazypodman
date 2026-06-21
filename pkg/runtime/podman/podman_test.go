package podman

import (
	"context"
	"errors"
	"testing"

	"github.com/jesseduffield/lazydocker/pkg/runtime"
)

// TestUnimplementedReportsUnsupported confirms the one operation not yet
// implemented (AttachContainer) still reports runtime.ErrUnsupported. The
// stub does not touch the connection, so a zero-value Runtime is enough.
func TestUnimplementedReportsUnsupported(t *testing.T) {
	r := &Runtime{}
	ctx := context.Background()

	if err := errOf(r.AttachContainer(ctx, "x", runtime.AttachOptions{})); !errors.Is(err, runtime.ErrUnsupported) {
		t.Errorf("AttachContainer: got %v, want errors.Is(..., ErrUnsupported)", err)
	}

	if err := r.Close(); err != nil {
		t.Errorf("Close: got %v, want nil", err)
	}
}

// TestResolveSocketURI covers the precedence of CONTAINER_HOST over the
// rootless default.
func TestResolveSocketURI(t *testing.T) {
	t.Run("CONTAINER_HOST wins", func(t *testing.T) {
		t.Setenv(containerHostEnvKey, "unix:///custom/podman.sock")
		if got := ResolveSocketURI(); got != "unix:///custom/podman.sock" {
			t.Errorf("got %q, want the CONTAINER_HOST value", got)
		}
	})

	t.Run("rootless default from XDG_RUNTIME_DIR", func(t *testing.T) {
		t.Setenv(containerHostEnvKey, "")
		t.Setenv("XDG_RUNTIME_DIR", "/run/user/1234")
		got := ResolveSocketURI()
		// Skip on root (and on platforms reporting euid 0) where the rootful
		// default is returned instead.
		if got == "unix:///run/podman/podman.sock" {
			t.Skip("running as root; rootful default returned")
		}
		if got != "unix:///run/user/1234/podman/podman.sock" {
			t.Errorf("got %q, want the rootless XDG socket", got)
		}
	})
}

// errOf discards a method's first return value and keeps the error.
func errOf[T any](_ T, err error) error { return err }
