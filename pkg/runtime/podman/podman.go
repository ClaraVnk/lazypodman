package podman

import (
	"context"
	"fmt"

	"github.com/jesseduffield/lazydocker/pkg/domain"
	"github.com/jesseduffield/lazydocker/pkg/runtime"
)

// Runtime is the native Podman implementation of
// runtime.ContainerRuntime, built on Podman's Go bindings. The container
// method group is implemented (Phase 3b); the remaining groups land in
// Phases 3c–3e and report runtime.ErrUnsupported until then.
type Runtime struct {
	// conn is the bindings connection context returned by
	// bindings.NewConnection; it carries the HTTP client to the Podman
	// socket and is passed to every bindings call.
	conn context.Context
}

// Compile-time check that Runtime satisfies the interface.
var _ runtime.ContainerRuntime = (*Runtime)(nil)

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
