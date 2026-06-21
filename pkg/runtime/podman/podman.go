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
// runtime.ContainerRuntime. In Phase 3a it is a scaffold: every operation
// returns runtime.ErrUnsupported. Method groups are filled in over
// Phases 3b–3e (containers, images, networks/volumes, events/stats/logs).
type Runtime struct{}

// Compile-time check that Runtime satisfies the interface.
var _ runtime.ContainerRuntime = (*Runtime)(nil)

// New returns a Podman runtime. The connection to the Podman socket is
// established in Phase 3b, when the first real method group needs it; the
// scaffold needs no connection because every call short-circuits to
// ErrUnsupported.
func New() *Runtime {
	return &Runtime{}
}

// unsupported wraps runtime.ErrUnsupported with the operation name so
// callers get a clear message while errors.Is keeps working.
func unsupported(op string) error {
	return fmt.Errorf("podman: %s: %w", op, runtime.ErrUnsupported)
}

// Close releases the runtime. The scaffold holds nothing.
func (r *Runtime) Close() error { return nil }

// ----- Containers -----

func (r *Runtime) ListContainers(ctx context.Context) ([]domain.ContainerInfo, error) {
	return nil, unsupported("list containers")
}

func (r *Runtime) InspectContainer(ctx context.Context, id string) (domain.ContainerDetails, error) {
	return domain.ContainerDetails{}, unsupported("inspect container")
}

func (r *Runtime) StartContainer(ctx context.Context, id string) error {
	return unsupported("start container")
}

func (r *Runtime) StopContainer(ctx context.Context, id string, timeout *time.Duration) error {
	return unsupported("stop container")
}

func (r *Runtime) RestartContainer(ctx context.Context, id string, timeout *time.Duration) error {
	return unsupported("restart container")
}

func (r *Runtime) PauseContainer(ctx context.Context, id string) error {
	return unsupported("pause container")
}

func (r *Runtime) UnpauseContainer(ctx context.Context, id string) error {
	return unsupported("unpause container")
}

func (r *Runtime) RemoveContainer(ctx context.Context, id string, opts runtime.RemoveContainerOptions) error {
	return unsupported("remove container")
}

func (r *Runtime) ContainerLogs(ctx context.Context, id string, opts runtime.LogOptions) (io.ReadCloser, error) {
	return nil, unsupported("container logs")
}

func (r *Runtime) ContainerTop(ctx context.Context, id string) (domain.TopOutput, error) {
	return domain.TopOutput{}, unsupported("container top")
}

func (r *Runtime) ContainerStats(ctx context.Context, id string) (<-chan domain.Stats, error) {
	return nil, unsupported("container stats")
}

func (r *Runtime) AttachContainer(ctx context.Context, id string, opts runtime.AttachOptions) (domain.AttachStream, error) {
	return nil, unsupported("attach container")
}

func (r *Runtime) PruneContainers(ctx context.Context) (domain.PruneReport, error) {
	return domain.PruneReport{}, unsupported("prune containers")
}

// ----- Images -----

func (r *Runtime) ListImages(ctx context.Context) ([]domain.ImageInfo, error) {
	return nil, unsupported("list images")
}

func (r *Runtime) RemoveImage(ctx context.Context, id string, opts runtime.RemoveImageOptions) error {
	return unsupported("remove image")
}

func (r *Runtime) ImageHistory(ctx context.Context, id string) ([]domain.ImageHistoryItem, error) {
	return nil, unsupported("image history")
}

func (r *Runtime) PruneImages(ctx context.Context) (domain.PruneReport, error) {
	return domain.PruneReport{}, unsupported("prune images")
}

// ----- Networks -----

func (r *Runtime) ListNetworks(ctx context.Context) ([]domain.NetworkInfo, error) {
	return nil, unsupported("list networks")
}

func (r *Runtime) RemoveNetwork(ctx context.Context, id string) error {
	return unsupported("remove network")
}

func (r *Runtime) PruneNetworks(ctx context.Context) (domain.PruneReport, error) {
	return domain.PruneReport{}, unsupported("prune networks")
}

// ----- Volumes -----

func (r *Runtime) ListVolumes(ctx context.Context) ([]domain.VolumeInfo, error) {
	return nil, unsupported("list volumes")
}

func (r *Runtime) RemoveVolume(ctx context.Context, name string, force bool) error {
	return unsupported("remove volume")
}

func (r *Runtime) PruneVolumes(ctx context.Context) (domain.PruneReport, error) {
	return domain.PruneReport{}, unsupported("prune volumes")
}

// ----- Events -----

func (r *Runtime) Events(ctx context.Context, since time.Time) (<-chan domain.Event, <-chan error) {
	events := make(chan domain.Event)
	errs := make(chan error, 1)
	errs <- unsupported("events")
	close(events)
	close(errs)
	return events, errs
}
