package runtime

import "errors"

// Sentinel errors returned by ContainerRuntime implementations. Callers
// should use errors.Is to test for them — backends are free to wrap them
// with extra context.
var (
	// ErrNotFound is returned when the requested object (container,
	// image, network, volume...) does not exist on the runtime.
	ErrNotFound = errors.New("not found")

	// ErrConflict is returned when an operation would conflict with the
	// current state (e.g. stopping an already-stopped container,
	// removing an in-use volume).
	ErrConflict = errors.New("conflict")

	// ErrUnauthorized is returned when the caller lacks the permission
	// to perform the operation on the runtime (rootless socket missing,
	// registry login required, etc.).
	ErrUnauthorized = errors.New("unauthorized")

	// ErrUnsupported is returned when the backend does not support the
	// requested operation (typically Podman-specific calls on the
	// Docker backend, or vice-versa).
	ErrUnsupported = errors.New("unsupported by this runtime")

	// ErrUnavailable is returned when the runtime daemon/socket is
	// unreachable.
	ErrUnavailable = errors.New("runtime unavailable")
)
