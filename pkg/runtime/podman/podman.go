package podman

import (
	"context"
	"fmt"
	"sync"

	"github.com/containers/podman/v5/pkg/bindings"

	"github.com/jesseduffield/lazydocker/pkg/domain"
	"github.com/jesseduffield/lazydocker/pkg/runtime"
)

// Runtime is the native Podman implementation of runtime.ContainerRuntime,
// built on Podman's Go bindings.
//
// The connection to the Podman socket is established lazily on first use,
// not at construction: like the Docker client, building a Runtime never
// fails just because the engine is down. The first call that needs the
// socket surfaces runtime.ErrUnavailable, which the GUI renders as a
// connection error instead of crashing at startup.
type Runtime struct {
	uri     string
	once    sync.Once
	conn    context.Context
	connErr error
}

// Compile-time check that Runtime satisfies the interface.
var _ runtime.ContainerRuntime = (*Runtime)(nil)

// client returns the bindings connection context, establishing it once on
// first use. A connection failure is reported as runtime.ErrUnavailable.
func (r *Runtime) client() (context.Context, error) {
	r.once.Do(func() {
		conn, err := bindings.NewConnection(context.Background(), r.uri)
		if err != nil {
			r.connErr = fmt.Errorf("podman: connect: %w: %s", runtime.ErrUnavailable, err.Error())
			return
		}
		r.conn = conn
	})
	return r.conn, r.connErr
}

// unsupported wraps runtime.ErrUnsupported with the operation name so
// callers get a clear message while errors.Is keeps working.
func unsupported(op string) error {
	return fmt.Errorf("podman: %s: %w", op, runtime.ErrUnsupported)
}

// Close releases the connection. The bindings client has no explicit
// close; dropping the reference is enough.
func (r *Runtime) Close() error {
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
