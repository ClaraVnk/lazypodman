package runtime

import (
	"context"
	"time"

	"github.com/ClaraVnk/lazypodman/pkg/domain"
)

// PodRuntime is an optional capability implemented by backends that
// support Podman-style pods. The Docker backend deliberately does not
// implement it. Callers discover the capability with a type assertion:
//
//	if pr, ok := rt.(runtime.PodRuntime); ok {
//		pods, err := pr.ListPods(ctx)
//	}
//
// See docs/adr/0006-podman-pods-and-capabilities.md.
type PodRuntime interface {
	ListPods(ctx context.Context) ([]domain.PodInfo, error)
	StartPod(ctx context.Context, id string) error
	StopPod(ctx context.Context, id string, timeout *time.Duration) error
	RestartPod(ctx context.Context, id string) error
	RemovePod(ctx context.Context, id string, force bool) error
	PrunePods(ctx context.Context) (domain.PruneReport, error)
}
