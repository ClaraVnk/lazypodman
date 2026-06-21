package podman

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jesseduffield/lazydocker/pkg/runtime"
)

// TestUnimplementedGroupsReportUnsupported confirms the method groups not
// yet implemented (images, networks, volumes, events, inspect) still
// report runtime.ErrUnsupported. These stubs do not touch the connection,
// so a zero-value Runtime is enough.
func TestUnimplementedGroupsReportUnsupported(t *testing.T) {
	r := &Runtime{}
	ctx := context.Background()

	checks := map[string]error{
		"InspectContainer": errOf(r.InspectContainer(ctx, "x")),
		"ListNetworks":     errOf(r.ListNetworks(ctx)),
		"RemoveNetwork":    r.RemoveNetwork(ctx, "x"),
		"ListVolumes":      errOf(r.ListVolumes(ctx)),
		"RemoveVolume":     r.RemoveVolume(ctx, "x", false),
		"ContainerStats":   errOf(r.ContainerStats(ctx, "x")),
	}
	for name, err := range checks {
		if !errors.Is(err, runtime.ErrUnsupported) {
			t.Errorf("%s: got %v, want errors.Is(..., ErrUnsupported)", name, err)
		}
	}

	if err := r.Close(); err != nil {
		t.Errorf("Close: got %v, want nil", err)
	}
}

// TestEventsClosesWithError confirms the still-stubbed Events emits one
// ErrUnsupported and then closes both channels.
func TestEventsClosesWithError(t *testing.T) {
	r := &Runtime{}
	evCh, errCh := r.Events(context.Background(), time.Time{})

	if err := <-errCh; !errors.Is(err, runtime.ErrUnsupported) {
		t.Errorf("Events err: got %v, want ErrUnsupported", err)
	}
	if _, ok := <-errCh; ok {
		t.Error("error channel should be closed after the single error")
	}
	if _, ok := <-evCh; ok {
		t.Error("events channel should be closed")
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
