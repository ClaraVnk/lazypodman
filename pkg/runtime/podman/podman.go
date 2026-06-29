package podman

import (
	"context"
	"fmt"
	"sync"

	"github.com/containers/podman/v5/pkg/bindings"

	"github.com/ClaraVnk/lazypodman/pkg/domain"
	"github.com/ClaraVnk/lazypodman/pkg/runtime"
)

// Runtime is the native Podman implementation of runtime.ContainerRuntime,
// built on Podman's Go bindings.
//
// The connection to the Podman socket is established lazily on first use,
// not at construction: like the Docker client, building a Runtime never
// fails just because the engine is down. The first call that needs the
// socket surfaces runtime.ErrUnavailable, which the GUI renders as a
// connection error instead of crashing at startup.
//
// The bindings connection is rooted at a lifetime context owned by the
// Runtime. Close cancels it, which tears down every in-flight streaming
// call (logs, stats, events) derived from the connection. All access to
// the connection state goes through the mutex so Close is safe to call
// concurrently with the listing/streaming methods, per the
// runtime.ContainerRuntime concurrency contract.
type Runtime struct {
	uri string

	mu sync.Mutex
	// lifeCancel cancels lifeCtx, the parent of the bindings connection.
	lifeCancel context.CancelFunc
	conn       context.Context
	closed     bool

	// runCommand runs an external command (used for quadlets, which are
	// managed via systemctl, not the API). Nil means use os/exec; tests
	// inject a fake. See quadlet.go.
	runCommand func(ctx context.Context, name string, args ...string) ([]byte, error)
}

// Compile-time check that Runtime satisfies the interface.
var _ runtime.ContainerRuntime = (*Runtime)(nil)

// client returns the bindings connection context, establishing it on first
// use. A connection failure is reported as runtime.ErrUnavailable and is
// not cached: the next call retries, so the engine coming back up (or the
// GUI's reconnect loop) can recover. After Close it always reports
// ErrUnavailable.
func (r *Runtime) client() (context.Context, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil, fmt.Errorf("podman: client: %w", runtime.ErrUnavailable)
	}
	if r.conn != nil {
		return r.conn, nil
	}

	lifeCtx, lifeCancel := context.WithCancel(context.Background())
	conn, err := bindings.NewConnection(lifeCtx, r.uri)
	if err != nil {
		lifeCancel()
		return nil, fmt.Errorf("podman: connect: %w: %s", runtime.ErrUnavailable, err.Error())
	}
	r.lifeCancel, r.conn = lifeCancel, conn
	return r.conn, nil
}

// unsupported wraps runtime.ErrUnsupported with the operation name so
// callers get a clear message while errors.Is keeps working.
func unsupported(op string) error {
	return fmt.Errorf("podman: %s: %w", op, runtime.ErrUnsupported)
}

// Close cancels the lifetime context, which tears down any in-flight
// streaming calls (logs, stats, events) rooted at the connection, and
// marks the Runtime closed so later calls report ErrUnavailable. It is
// idempotent and safe to call concurrently with other methods.
func (r *Runtime) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.closed = true
	if r.lifeCancel != nil {
		r.lifeCancel()
		r.lifeCancel = nil
	}
	r.conn = nil
	return nil
}

// ----- Containers ----- (implemented in container.go)

// ----- Images ----- (implemented in image.go)

// ----- Networks ----- (implemented in network.go)

// ----- Volumes ----- (implemented in volume.go)

// ----- Logs (logs.go) / Stats (stats.go) / Events (event.go) -----

// AttachContainer stays a follow-up: upstream lazydocker attaches via the
// CLI (exec) path in pkg/commands, not the SDK/bindings; the runtime
// affordance is revisited once that flow is ported.
func (r *Runtime) AttachContainer(ctx context.Context, id string, opts runtime.AttachOptions) (domain.AttachStream, error) {
	return nil, unsupported("attach container")
}
