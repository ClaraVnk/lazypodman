//go:build docker

package docker

import (
	"io"

	"github.com/docker/docker/pkg/stdcopy"
)

// demuxedLogStream wraps a Docker multiplexed log stream (stdout+stderr
// framed with stdcopy) and exposes a plain UTF-8 io.ReadCloser. A
// background goroutine pumps the framed stream through stdcopy.StdCopy
// into a pipe; the returned reader is the read end of that pipe.
//
// Both stdout and stderr are merged into the same output stream — that
// matches what pkg/gui has always done with logs and avoids forcing
// every caller to handle two channels. A richer dual-stream API can be
// added if a real caller needs it.
func demuxedLogStream(src io.ReadCloser) io.ReadCloser {
	pr, pw := io.Pipe()
	go func() {
		_, err := stdcopy.StdCopy(pw, pw, src)
		// Surface the demux error to the reader so callers see io.EOF
		// only on clean termination.
		_ = pw.CloseWithError(err)
	}()
	return &demuxStream{PipeReader: pr, src: src}
}

// demuxStream is a ReadCloser that closes both the pipe (waking the
// reader) and the underlying source (stopping the demux goroutine).
type demuxStream struct {
	*io.PipeReader
	src io.ReadCloser
}

// Close stops the demux goroutine by closing the source (StdCopy
// returns on EOF) and tears down the pipe.
func (d *demuxStream) Close() error {
	srcErr := d.src.Close()
	pipeErr := d.PipeReader.Close()
	if srcErr != nil {
		return srcErr
	}
	return pipeErr
}
