// Package runtime defines the abstraction that lazypodman uses to talk to
// a container engine. The interface is consumed by pkg/commands and
// pkg/gui; concrete implementations live in pkg/runtime/docker and
// pkg/runtime/podman.
//
// The interface trades exclusively in pkg/domain types — no Docker SDK
// type, no Podman binding type ever crosses this boundary.
package runtime
