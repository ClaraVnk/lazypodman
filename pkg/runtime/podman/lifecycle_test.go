package podman

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/jesseduffield/lazydocker/pkg/runtime"
)

// unreachableURI points at a unix socket that does not exist, so
// bindings.NewConnection fails fast without touching a real engine.
const unreachableURI = "unix:///nonexistent/lazypodman-test.sock"

// TestClientAfterCloseUnavailable verifies that once the Runtime is closed,
// client() reports runtime.ErrUnavailable instead of attempting to redial.
func TestClientAfterCloseUnavailable(t *testing.T) {
	r := &Runtime{uri: unreachableURI}

	if err := r.Close(); err != nil {
		t.Fatalf("Close: got %v, want nil", err)
	}

	_, err := r.client()
	if !errors.Is(err, runtime.ErrUnavailable) {
		t.Errorf("client after Close: got %v, want errors.Is(..., ErrUnavailable)", err)
	}

	// A method that goes through client() must surface the same error.
	if _, err := r.ListContainers(context.Background()); !errors.Is(err, runtime.ErrUnavailable) {
		t.Errorf("ListContainers after Close: got %v, want errors.Is(..., ErrUnavailable)", err)
	}
}

// TestClientFailureNotCached verifies that a failed connection is not cached:
// every call retries, so the engine coming back up can recover. With an
// unreachable socket both attempts fail with ErrUnavailable.
func TestClientFailureNotCached(t *testing.T) {
	r := &Runtime{uri: unreachableURI}

	for i, attempt := 0, 2; i < attempt; i++ {
		if _, err := r.client(); !errors.Is(err, runtime.ErrUnavailable) {
			t.Errorf("client() attempt %d: got %v, want errors.Is(..., ErrUnavailable)", i+1, err)
		}
	}
}

// TestCloseConcurrencySafe exercises the documented concurrency contract:
// Close may run concurrently with the streaming/listing methods. Run under
// -race (test.sh does) to catch unsynchronized access to the connection
// state. Close is also expected to be idempotent.
func TestCloseConcurrencySafe(t *testing.T) {
	r := &Runtime{uri: unreachableURI}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			_, _ = r.client()
		}()
		go func() {
			defer wg.Done()
			_ = r.Close()
		}()
	}
	wg.Wait()

	// After all the churn the Runtime must be closed and stay unavailable.
	if _, err := r.client(); !errors.Is(err, runtime.ErrUnavailable) {
		t.Errorf("client after concurrent Close: got %v, want errors.Is(..., ErrUnavailable)", err)
	}
}
