package docker

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"testing"
	"time"
)

// frame writes one Docker multiplexed-stream frame (header + payload) to buf.
// The frame format is: 1 byte stream id (1=stdout, 2=stderr), 3 bytes
// padding, 4 bytes big-endian length, then payload.
func frame(buf *bytes.Buffer, streamID byte, payload []byte) {
	header := [8]byte{}
	header[0] = streamID
	binary.BigEndian.PutUint32(header[4:], uint32(len(payload)))
	buf.Write(header[:])
	buf.Write(payload)
}

func TestDemuxedLogStream_MergesStdoutAndStderr(t *testing.T) {
	src := &bytes.Buffer{}
	frame(src, 1, []byte("hello "))
	frame(src, 2, []byte("world\n"))
	frame(src, 1, []byte("bye"))

	rc := demuxedLogStream(io.NopCloser(src))
	defer rc.Close()

	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(got) != "hello world\nbye" {
		t.Errorf("output = %q, want %q", string(got), "hello world\nbye")
	}
}

func TestDemuxedLogStream_EmptySource(t *testing.T) {
	rc := demuxedLogStream(io.NopCloser(&bytes.Buffer{}))
	defer rc.Close()

	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty output, got %q", string(got))
	}
}

// closeTracker reports whether Close was called on the underlying source.
type closeTracker struct {
	io.Reader
	closed bool
}

func (c *closeTracker) Close() error {
	c.closed = true
	return nil
}

func TestDemuxedLogStream_CloseStopsGoroutine(t *testing.T) {
	src := &closeTracker{Reader: bytes.NewReader(nil)}
	rc := demuxedLogStream(src)

	if err := rc.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if !src.closed {
		t.Error("Close did not propagate to the underlying source")
	}

	// Give the goroutine a moment to exit; a leak would not be fatal here
	// but the test ensures the close path is wired.
	time.Sleep(10 * time.Millisecond)
}

// errReader returns the given error on the first Read.
type errReader struct {
	err error
}

func (e *errReader) Read(_ []byte) (int, error) { return 0, e.err }
func (e *errReader) Close() error               { return nil }

func TestDemuxedLogStream_SourceErrorSurfacesToReader(t *testing.T) {
	want := errors.New("upstream boom")
	rc := demuxedLogStream(&errReader{err: want})
	defer rc.Close()

	_, err := io.ReadAll(rc)
	if !errors.Is(err, want) {
		t.Fatalf("ReadAll err = %v, want chain containing %v", err, want)
	}
}
