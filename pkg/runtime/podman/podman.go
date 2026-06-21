package podman

import (
	"context"
	"fmt"
	"io"
	"time"

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

// InspectContainer lands in a follow-up increment (the ContainerDetails
// mapper); until then it reports unsupported.
func (r *Runtime) InspectContainer(ctx context.Context, id string) (domain.ContainerDetails, error) {
	return domain.ContainerDetails{}, unsupported("inspect container")
}

// ----- Images ----- (implemented in image.go)

// ----- Networks ----- (implemented in network.go)

// ----- Volumes ----- (implemented in volume.go)

// ----- Logs / Stats / Attach ----- (Phase 3e)

func (r *Runtime) ContainerLogs(ctx context.Context, id string, opts runtime.LogOptions) (io.ReadCloser, error) {
	return nil, unsupported("container logs")
}

func (r *Runtime) ContainerStats(ctx context.Context, id string) (<-chan domain.Stats, error) {
	return nil, unsupported("container stats")
}

func (r *Runtime) AttachContainer(ctx context.Context, id string, opts runtime.AttachOptions) (domain.AttachStream, error) {
	return nil, unsupported("attach container")
}

// ----- Events ----- (Phase 3e)

func (r *Runtime) Events(ctx context.Context, since time.Time) (<-chan domain.Event, <-chan error) {
	events := make(chan domain.Event)
	errs := make(chan error, 1)
	errs <- unsupported("events")
	close(events)
	close(errs)
	return events, errs
}
