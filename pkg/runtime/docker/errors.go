package docker

import (
	"fmt"

	cerrdefs "github.com/containerd/errdefs"
	dockerclient "github.com/docker/docker/client"

	"github.com/ClaraVnk/lazypodman/pkg/runtime"
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
	case cerrdefs.IsNotFound(err):
		sentinel = runtime.ErrNotFound
	case cerrdefs.IsConflict(err):
		sentinel = runtime.ErrConflict
	case cerrdefs.IsUnauthorized(err) || cerrdefs.IsPermissionDenied(err):
		sentinel = runtime.ErrUnauthorized
	case cerrdefs.IsNotImplemented(err):
		sentinel = runtime.ErrUnsupported
	case cerrdefs.IsUnavailable(err) || dockerclient.IsErrConnectionFailed(err):
		sentinel = runtime.ErrUnavailable
	default:
		return fmt.Errorf("%s: %w", op, err)
	}
	return fmt.Errorf("%s: %w: %s", op, sentinel, err.Error())
}
