package podman

import (
	"context"
	"io"

	"github.com/containers/podman/v5/pkg/bindings/containers"

	"github.com/jesseduffield/lazydocker/pkg/runtime"
)

// ContainerLogs streams a container's logs. Podman's bindings deliver
// already-demuxed stdout/stderr as lines on two channels (no stdcopy
// framing, unlike Docker), so we bridge them into a single io.ReadCloser.
func (r *Runtime) ContainerLogs(ctx context.Context, id string, opts runtime.LogOptions) (io.ReadCloser, error) {
	o := new(containers.LogOptions).
		WithStdout(true).
		WithStderr(true).
		WithFollow(opts.Follow).
		WithTimestamps(opts.Timestamps)
	if opts.Tail != "" {
		o = o.WithTail(opts.Tail)
	}
	if opts.Since != "" {
		o = o.WithSince(opts.Since)
	}
	if opts.Until != "" {
		o = o.WithUntil(opts.Until)
	}

	stdoutCh := make(chan string, 64)
	stderrCh := make(chan string, 64)
	pr, pw := io.Pipe()
	// Derive from r.conn so the call keeps the bindings client but can be
	// cancelled when the reader is closed.
	logCtx, cancel := context.WithCancel(r.conn)

	// Producer: the blocking Logs call. It does not close the line
	// channels, so we do once it returns.
	go func() {
		err := containers.Logs(logCtx, id, o, stdoutCh, stderrCh)
		close(stdoutCh)
		close(stderrCh)
		_ = err // surfaced to the reader as EOF when the pipe closes
	}()

	// Bridge: drain both channels into the pipe until both close.
	go func() {
		defer pw.Close()
		for stdoutCh != nil || stderrCh != nil {
			select {
			case line, ok := <-stdoutCh:
				if !ok {
					stdoutCh = nil
					continue
				}
				_, _ = io.WriteString(pw, line+"\n")
			case line, ok := <-stderrCh:
				if !ok {
					stderrCh = nil
					continue
				}
				_, _ = io.WriteString(pw, line+"\n")
			}
		}
	}()

	return &logReadCloser{r: pr, cancel: cancel}, nil
}

// logReadCloser cancels the underlying Logs subscription on Close.
type logReadCloser struct {
	r      *io.PipeReader
	cancel context.CancelFunc
}

func (l *logReadCloser) Read(p []byte) (int, error) { return l.r.Read(p) }

func (l *logReadCloser) Close() error {
	l.cancel()
	return l.r.Close()
}
