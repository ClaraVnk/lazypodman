// Package compliance holds the dual-backend conformance suite: the same
// read-contract assertions run against both the Docker and Podman
// implementations of runtime.ContainerRuntime, so we have an objective,
// reusable definition of "parity" before flipping the default backend
// (ADR 0005, Phase 4).
//
// The suite is opt-in: it talks to a live engine socket, so it is skipped
// unless LAZYPODMAN_INTEGRATION=1. Each backend is skipped individually if
// its socket cannot be reached. Build with the Podman client tags:
//
//	LAZYPODMAN_INTEGRATION=1 go test -tags \
//	  'containers_image_openpgp exclude_graphdriver_btrfs exclude_graphdriver_devicemapper remote' \
//	  ./pkg/runtime/compliance/...
package compliance

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/ClaraVnk/lazypodman/pkg/runtime"
	dockerruntime "github.com/ClaraVnk/lazypodman/pkg/runtime/docker"
	podmanruntime "github.com/ClaraVnk/lazypodman/pkg/runtime/podman"
)

func TestCompliance(t *testing.T) {
	if os.Getenv("LAZYPODMAN_INTEGRATION") != "1" {
		t.Skip("set LAZYPODMAN_INTEGRATION=1 to run the dual-backend compliance suite")
	}

	backends := map[string]func() (runtime.ContainerRuntime, error){
		"docker": func() (runtime.ContainerRuntime, error) { return dockerruntime.NewFromEnv() },
		"podman": func() (runtime.ContainerRuntime, error) { return podmanruntime.NewFromEnv() },
	}

	ran := 0
	for name, build := range backends {
		rt, err := build()
		if err != nil {
			t.Logf("%s: backend unavailable, skipping (%v)", name, err)
			continue
		}
		// Construction is lazy for both backends, so probe the socket and
		// skip the backend when the engine is not reachable.
		probeCtx, probeCancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, probeErr := rt.ListContainers(probeCtx)
		probeCancel()
		if errors.Is(probeErr, runtime.ErrUnavailable) {
			rt.Close()
			t.Logf("%s: socket unreachable, skipping (%v)", name, probeErr)
			continue
		}
		ran++
		t.Run(name, func(t *testing.T) {
			defer rt.Close()
			runReadContracts(t, rt)
		})
	}
	if ran == 0 {
		t.Skip("no engine socket reachable for either backend")
	}
}

// runReadContracts exercises every read path of the interface and asserts
// the contract documented on ContainerRuntime: listings never error on a
// healthy engine and return a non-nil slice; inspect of a listed container
// succeeds and round-trips its ID.
func runReadContracts(t *testing.T, rt runtime.ContainerRuntime) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	containers, err := rt.ListContainers(ctx)
	if err != nil {
		t.Fatalf("ListContainers: %v", err)
	}
	if containers == nil {
		t.Error("ListContainers returned a nil slice; want non-nil (possibly empty)")
	}

	images, err := rt.ListImages(ctx)
	if err != nil {
		t.Fatalf("ListImages: %v", err)
	}
	if images == nil {
		t.Error("ListImages returned a nil slice")
	}

	if _, err := rt.ListNetworks(ctx); err != nil {
		t.Fatalf("ListNetworks: %v", err)
	}
	if _, err := rt.ListVolumes(ctx); err != nil {
		t.Fatalf("ListVolumes: %v", err)
	}

	// Inspect the first container, if any, and confirm the ID round-trips.
	if len(containers) > 0 {
		id := containers[0].ID
		details, err := rt.InspectContainer(ctx, id)
		if err != nil {
			t.Fatalf("InspectContainer(%s): %v", id, err)
		}
		if details.ID != id {
			t.Errorf("InspectContainer ID = %q, want %q", details.ID, id)
		}
		if hist, err := rt.ImageHistory(ctx, containers[0].ImageID); err != nil && containers[0].ImageID != "" {
			t.Logf("ImageHistory(%s): %v", containers[0].ImageID, err)
			_ = hist
		}
	}

	// Events: the stream must open and shut down cleanly on context cancel.
	ectx, ecancel := context.WithTimeout(ctx, time.Second)
	defer ecancel()
	evCh, errCh := rt.Events(ectx, time.Now().Add(-time.Minute))
	for open := true; open; {
		select {
		case _, ok := <-evCh:
			if !ok {
				open = false
			}
		case <-errCh:
		case <-ectx.Done():
			open = false
		}
	}
}
