// Package runtime defines the abstraction that lazypodman uses to talk to
// a container engine. The interface is consumed by pkg/commands and
// pkg/gui; the concrete implementation lives in pkg/runtime/podman.
//
// The interface trades exclusively in pkg/domain types — no Podman binding
// type ever crosses this boundary.
package runtime
