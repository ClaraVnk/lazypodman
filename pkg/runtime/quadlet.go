package runtime

import (
	"context"

	"github.com/jesseduffield/lazydocker/pkg/domain"
)

// QuadletManager is an optional capability for backends that manage Podman
// quadlets (systemd units). It is implemented over `systemctl --user`, not
// the Podman API. The Docker backend does not implement it. Callers
// discover it with a type assertion:
//
//	if qm, ok := rt.(runtime.QuadletManager); ok {
//		quadlets, err := qm.ListQuadlets(ctx)
//	}
//
// Enable/disable (autostart) is intentionally absent in v1 — for a
// generated unit it is a source-file [Install] edit, not a systemctl call.
// See docs/adr/0007-quadlets.md.
type QuadletManager interface {
	ListQuadlets(ctx context.Context) ([]domain.Quadlet, error)
	StartQuadlet(ctx context.Context, unit string) error
	StopQuadlet(ctx context.Context, unit string) error
	RestartQuadlet(ctx context.Context, unit string) error
}
