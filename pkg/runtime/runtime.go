package runtime

import (
	"context"
	"io"
	"time"

	"github.com/ClaraVnk/lazypodman/pkg/domain"
)

// ContainerRuntime is everything lazypodman needs from a container
// engine. Implementations must be safe for concurrent use by multiple
// goroutines (the GUI runs refresh loops in parallel).
//
// Method-level guarantees:
//   - Every method respects context cancellation.
//   - Listing methods return an empty slice (not nil) when the runtime
//     has no objects, and never return a nil-but-non-error result.
//   - Single-object methods return a wrapped ErrNotFound when the ID is
//     unknown — callers test with errors.Is(err, runtime.ErrNotFound).
type ContainerRuntime interface {
	// Close releases the underlying connection and any background
	// goroutines. After Close, all other methods return ErrUnavailable.
	Close() error

	// ----- Containers -----

	ListContainers(ctx context.Context) ([]domain.ContainerInfo, error)
	InspectContainer(ctx context.Context, id string) (domain.ContainerDetails, error)

	StartContainer(ctx context.Context, id string) error
	StopContainer(ctx context.Context, id string, timeout *time.Duration) error
	RestartContainer(ctx context.Context, id string, timeout *time.Duration) error
	PauseContainer(ctx context.Context, id string) error
	UnpauseContainer(ctx context.Context, id string) error
	RemoveContainer(ctx context.Context, id string, opts RemoveContainerOptions) error

	ContainerLogs(ctx context.Context, id string, opts LogOptions) (io.ReadCloser, error)
	ContainerTop(ctx context.Context, id string) (domain.TopOutput, error)
	ContainerStats(ctx context.Context, id string) (<-chan domain.Stats, error)
	AttachContainer(ctx context.Context, id string, opts AttachOptions) (domain.AttachStream, error)

	PruneContainers(ctx context.Context) (domain.PruneReport, error)

	// ----- Images -----

	ListImages(ctx context.Context) ([]domain.ImageInfo, error)
	RemoveImage(ctx context.Context, id string, opts RemoveImageOptions) error
	ImageHistory(ctx context.Context, id string) ([]domain.ImageHistoryItem, error)
	PruneImages(ctx context.Context) (domain.PruneReport, error)

	// ----- Networks -----

	ListNetworks(ctx context.Context) ([]domain.NetworkInfo, error)
	RemoveNetwork(ctx context.Context, id string) error
	PruneNetworks(ctx context.Context) (domain.PruneReport, error)

	// ----- Volumes -----

	ListVolumes(ctx context.Context) ([]domain.VolumeInfo, error)
	RemoveVolume(ctx context.Context, name string, force bool) error
	PruneVolumes(ctx context.Context) (domain.PruneReport, error)

	// ----- Events -----

	// Events returns a channel that emits engine events as they happen.
	// The errors channel is closed when the events channel is closed.
	// Callers must consume errors or risk blocking the producer.
	Events(ctx context.Context, since time.Time) (<-chan domain.Event, <-chan error)
}

// RemoveContainerOptions are the toggles accepted by RemoveContainer.
type RemoveContainerOptions struct {
	// Force removes the container even if it is running (SIGKILL first).
	Force bool
	// RemoveVolumes deletes the anonymous volumes associated with the
	// container. Named volumes are never deleted.
	RemoveVolumes bool
	// RemoveLinks deletes the legacy --link references pointing at this
	// container. Ignored on backends that do not support links (Podman).
	RemoveLinks bool
}

// RemoveImageOptions are the toggles accepted by RemoveImage.
type RemoveImageOptions struct {
	// Force removes the image even if it has multiple tags or is used
	// by stopped containers.
	Force bool
	// NoPrune disables automatic deletion of untagged parents.
	NoPrune bool
}

// LogOptions controls how ContainerLogs streams the log output.
//
// Since/Until accept any value the underlying runtime accepts as a time
// spec: a Unix timestamp ("1700000000"), an RFC3339 timestamp, or a
// relative duration ("10m", "1h30m"). Empty means "unrestricted".
type LogOptions struct {
	// Follow keeps the stream open and emits new entries as they
	// appear. The caller must Close the reader to stop following.
	Follow bool
	// Tail is the maximum number of historical lines to return, or "all"
	// for the full history. Empty means "all".
	Tail string
	// Since restricts the stream to entries newer than this point.
	Since string
	// Until restricts the stream to entries older than this point.
	Until string
	// Timestamps prefixes each line with its RFC3339 timestamp.
	Timestamps bool
	// TTY tells the runtime whether the container was created with a
	// pseudo-TTY. When true the stream is plain UTF-8; when false the
	// stream is multiplexed at the wire level (stdout + stderr framed)
	// and the runtime adapter is responsible for demuxing into a plain
	// stream — callers never see the multiplexed bytes.
	TTY bool
}

// AttachOptions controls how AttachContainer wires the streams.
type AttachOptions struct {
	// Stdin requests an interactive input stream.
	Stdin bool
	// Stdout requests the standard output stream.
	Stdout bool
	// Stderr requests the standard error stream.
	Stderr bool
	// DetachKeys is the key sequence that the user types to detach,
	// e.g. "ctrl-p,ctrl-q". Empty means "use the runtime default".
	DetachKeys string
	// TTY allocates a pseudo-TTY for the attachment.
	TTY bool
}
