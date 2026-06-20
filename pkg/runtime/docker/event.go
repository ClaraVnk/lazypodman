package docker

import (
	"context"
	"strconv"
	"time"

	dockerevents "github.com/docker/docker/api/types/events"

	"github.com/jesseduffield/lazydocker/pkg/domain"
)

// Events streams runtime events. The returned channels are closed when
// ctx is cancelled or the underlying stream ends.
func (r *Runtime) Events(ctx context.Context, since time.Time) (<-chan domain.Event, <-chan error) {
	events := make(chan domain.Event)
	errs := make(chan error, 1)

	opts := dockerevents.ListOptions{}
	if !since.IsZero() {
		opts.Since = strconv.FormatInt(since.Unix(), 10)
	}

	src, srcErrs := r.cli.Events(ctx, opts)

	go func() {
		defer close(events)
		defer close(errs)
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-src:
				if !ok {
					return
				}
				select {
				case events <- eventToDomain(msg):
				case <-ctx.Done():
					return
				}
			case err, ok := <-srcErrs:
				if !ok {
					return
				}
				if err != nil {
					select {
					case errs <- mapErr("events stream", err):
					case <-ctx.Done():
					}
					return
				}
			}
		}
	}()

	return events, errs
}
