package podman

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jesseduffield/lazydocker/pkg/runtime"
)

// TestStubReportsUnsupported confirms the Phase 3a scaffold satisfies the
// interface and reports runtime.ErrUnsupported (via errors.Is) for every
// operation, so callers degrade predictably until the real methods land.
func TestStubReportsUnsupported(t *testing.T) {
	r := New()
	ctx := context.Background()

	checks := map[string]error{
		"ListContainers":   errOf(r.ListContainers(ctx)),
		"InspectContainer": errOf(r.InspectContainer(ctx, "x")),
		"StartContainer":   r.StartContainer(ctx, "x"),
		"StopContainer":    r.StopContainer(ctx, "x", nil),
		"RemoveContainer":  r.RemoveContainer(ctx, "x", runtime.RemoveContainerOptions{}),
		"ContainerStats":   errOf(r.ContainerStats(ctx, "x")),
		"ListImages":       errOf(r.ListImages(ctx)),
		"ListNetworks":     errOf(r.ListNetworks(ctx)),
		"ListVolumes":      errOf(r.ListVolumes(ctx)),
		"RemoveVolume":     r.RemoveVolume(ctx, "x", false),
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

// TestEventsClosesWithError confirms Events emits one ErrUnsupported and
// then closes both channels, matching the interface contract.
func TestEventsClosesWithError(t *testing.T) {
	r := New()
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

// errOf discards a method's first return value and keeps the error, so a
// table can mix methods with different value types.
func errOf[T any](_ T, err error) error { return err }
