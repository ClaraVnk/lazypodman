package podman

import (
	"bufio"
	"context"
	"os"

	"github.com/containers/podman/v5/pkg/api/handlers"
	"github.com/containers/podman/v5/pkg/bindings/containers"
	dockerContainer "github.com/docker/docker/api/types/container"

	"github.com/ClaraVnk/lazypodman/pkg/runtime"
)

// ExecShell runs an interactive command inside a container, attached to the
// process's terminal. It creates an exec session and starts+attaches it via
// the Podman bindings.
func (r *Runtime) ExecShell(ctx context.Context, id string, opts runtime.ExecShellOptions) error {
	conn, err := r.client()
	if err != nil {
		return err
	}

	sessionID, err := containers.ExecCreate(conn, id, &handlers.ExecCreateConfig{
		ExecOptions: dockerContainer.ExecOptions{
			Cmd:          opts.Command,
			Tty:          opts.Tty,
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
		},
	})
	if err != nil {
		return mapErr("exec create", err)
	}

	// INFO: Wire the process terminal into the exec session. ExecStartAndAttach puts
	//       os.Stdin into raw mode and forwards keystrokes to the container, and
	//       returns when the session ends.
	attach := new(containers.ExecStartAndAttachOptions).
				WithInputStream(*bufio.NewReader(os.Stdin)).
				WithOutputStream(os.Stdout).
				WithErrorStream(os.Stderr).
				WithAttachInput(true).
				WithAttachOutput(true).
				WithAttachError(true)

	return mapErr("exec start", containers.ExecStartAndAttach(conn, sessionID, attach))
}
