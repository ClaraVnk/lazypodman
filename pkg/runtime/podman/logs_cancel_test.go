package podman

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/jesseduffield/lazydocker/pkg/domain"
	"github.com/jesseduffield/lazydocker/pkg/runtime"
)

// TestContainerLogsCancelUnblocks is an integration regression test for the
// container-switching freeze: a followed log stream must tear down when the
// caller cancels its context, not only when the reader is closed. Before the
// fix, ContainerLogs derived its stream context solely from the connection,
// so cancelling the caller's ctx left the GUI's io.Copy blocked forever,
// which wedged the task manager and froze every panel on the next selection.
//
// Gated behind LAZYPODMAN_INTEGRATION=1 (like the compliance suite) and
// skipped unless a Podman socket is reachable and a running container exists.
func TestContainerLogsCancelUnblocks(t *testing.T) {
	if os.Getenv("LAZYPODMAN_INTEGRATION") != "1" {
		t.Skip("set LAZYPODMAN_INTEGRATION=1 to run the Podman logs-cancellation integration test")
	}

	r, err := NewFromEnv()
	if err != nil {
		t.Skipf("no podman runtime: %v", err)
	}
	defer r.Close()

	ctx := context.Background()
	containers, err := r.ListContainers(ctx)
	if err != nil {
		t.Skipf("no podman socket reachable: %v", err)
	}
	var id string
	for _, c := range containers {
		if c.State == domain.ContainerStateRunning {
			id = c.ID
			break
		}
	}
	if id == "" {
		t.Skip("no running container to stream logs from")
	}

	streamCtx, cancel := context.WithCancel(ctx)
	rc, err := r.ContainerLogs(streamCtx, id, runtime.LogOptions{Follow: true, Tail: "5"})
	if err != nil {
		cancel()
		t.Fatalf("ContainerLogs: %v", err)
	}

	// Drain in the background; cancelling streamCtx must make this return.
	done := make(chan error, 1)
	go func() {
		_, e := io.Copy(io.Discard, rc)
		done <- e
	}()

	// Let the stream establish, then cancel and require a prompt unblock.
	time.Sleep(500 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// Reader unblocked on ctx cancel — fix is in place.
	case <-time.After(3 * time.Second):
		t.Fatal("ContainerLogs did not unblock within 3s of ctx cancel (container-switch freeze regression)")
	}
	_ = rc.Close()
}
