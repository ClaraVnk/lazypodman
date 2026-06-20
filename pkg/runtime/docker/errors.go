package docker

import (
	"errors"
	"fmt"

	dockerclient "github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"

	"github.com/jesseduffield/lazydocker/pkg/runtime"
)

// mapErr converts an error returned by the Docker SDK into one of the
// runtime sentinel errors, preserving the original message as context.
// Non-Docker errors are returned unchanged.
func mapErr(op string, err error) error {
	if err == nil {
		return nil
	}
	var sentinel error
	switch {
	case errdefs.IsNotFound(err) || dockerclient.IsErrNotFound(err):
		sentinel = runtime.ErrNotFound
	case errdefs.IsConflict(err):
		sentinel = runtime.ErrConflict
	case errdefs.IsUnauthorized(err) || errdefs.IsForbidden(err):
		sentinel = runtime.ErrUnauthorized
	case errdefs.IsNotImplemented(err):
		sentinel = runtime.ErrUnsupported
	case errdefs.IsUnavailable(err) || dockerclient.IsErrConnectionFailed(err):
		sentinel = runtime.ErrUnavailable
	default:
		return fmt.Errorf("%s: %w", op, err)
	}
	return fmt.Errorf("%s: %w: %s", op, sentinel, err.Error())
}

// asSentinel exposes the wrapped sentinel error for tests.
func asSentinel(err error) error {
	for _, target := range []error{
		runtime.ErrNotFound,
		runtime.ErrConflict,
		runtime.ErrUnauthorized,
		runtime.ErrUnsupported,
		runtime.ErrUnavailable,
	} {
		if errors.Is(err, target) {
			return target
		}
	}
	return nil
}
