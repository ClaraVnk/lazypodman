package podman

import "fmt"

// mapErr wraps a Podman bindings error with the operation name. Richer
// mapping onto the runtime sentinel errors (ErrNotFound, ErrConflict, …)
// is added as the call sites that need it land in later phases.
func mapErr(op string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("podman: %s: %w", op, err)
}
