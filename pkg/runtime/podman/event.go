package podman

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/containers/podman/v5/pkg/bindings/system"
	entitiesTypes "github.com/containers/podman/v5/pkg/domain/entities/types"

	"github.com/ClaraVnk/lazypodman/pkg/domain"
)

// Events streams engine events. Podman's system.Events is a blocking call
// that writes to a channel it closes when the stream ends, and stops when
// signaled on a cancel channel. We run it in a goroutine and translate
// into the runtime's (events, errors) channel pair.
func (r *Runtime) Events(ctx context.Context, since time.Time) (<-chan domain.Event, <-chan error) {
	events := make(chan domain.Event)
	errs := make(chan error, 1)

	conn, err := r.client()
	if err != nil {
		errs <- err
		close(events)
		close(errs)
		return events, errs
	}

	src := make(chan entitiesTypes.Event)
	cancel := make(chan bool, 1)
	streamErr := make(chan error, 1)

	opts := new(system.EventsOptions).WithStream(true)
	if !since.IsZero() {
		opts = opts.WithSince(strconv.FormatInt(since.Unix(), 10))
	}

	go func() {
		// system.Events closes src on return.
		streamErr <- system.Events(conn, src, cancel, opts)
	}()

	go func() {
		defer close(events)
		defer close(errs)
		for {
			select {
			case <-ctx.Done():
				signalCancel(cancel)
				return
			case ev, ok := <-src:
				if !ok {
					if err := <-streamErr; err != nil {
						errs <- mapErr("events stream", err)
					}
					return
				}
				select {
				case events <- podmanEventToDomain(ev):
				case <-ctx.Done():
					signalCancel(cancel)
					return
				}
			}
		}
	}()

	return events, errs
}

// signalCancel pokes the cancel channel without blocking if it is already
// signaled.
func signalCancel(cancel chan bool) {
	select {
	case cancel <- true:
	default:
	}
}

func podmanEventToDomain(e entitiesTypes.Event) domain.Event {
	return domain.Event{
		Type:    mapEventType(string(e.Type)),
		Action:  string(e.Action),
		ActorID: e.Actor.ID,
		Actor:   e.Actor.Attributes["name"],
		Scope:   e.Scope,
		Time:    time.Unix(e.Time, e.TimeNano%int64(time.Second)),
		Attrs:   e.Actor.Attributes,
	}
}

func mapEventType(s string) domain.EventType {
	switch strings.ToLower(s) {
	case "container":
		return domain.EventTypeContainer
	case "image":
		return domain.EventTypeImage
	case "network":
		return domain.EventTypeNetwork
	case "volume":
		return domain.EventTypeVolume
	case "pod":
		return domain.EventTypePod
	default:
		return domain.EventTypeSystem
	}
}
