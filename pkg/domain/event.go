package domain

import "time"

// EventType is the kind of object an event is about.
type EventType string

const (
	EventTypeContainer EventType = "container"
	EventTypeImage     EventType = "image"
	EventTypeNetwork   EventType = "network"
	EventTypeVolume    EventType = "volume"
	EventTypePod       EventType = "pod" // Podman-only, irrelevant on the Docker backend
	EventTypeSystem    EventType = "system"
)

// Event is a runtime event (start/stop/destroy/...) emitted by the
// container engine. Equivalent to docker's events.Message and podman's
// Event.
type Event struct {
	Type    EventType
	Action  string // "start" | "stop" | "die" | "destroy" | "create" | ...
	ActorID string // ID of the object the event is about
	Actor   string // human-readable identifier (container name, image tag...)
	Scope   string // "local" | "swarm"
	Time    time.Time
	Attrs   map[string]string
}
