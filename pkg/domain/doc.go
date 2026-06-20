// Package domain contains the value types exchanged between the runtime
// abstraction (pkg/runtime) and the rest of the application (pkg/commands,
// pkg/gui).
//
// These types are deliberately plain data with no methods that touch I/O,
// no dependency on any container runtime SDK, and no logging. They exist
// so that the UI and the orchestration layer never have to import a
// runtime-specific type (Docker SDK, Podman bindings, etc.).
//
// Field sets are derived from what the GUI actually renders today. Add
// fields only when there is a concrete reader for them.
package domain
