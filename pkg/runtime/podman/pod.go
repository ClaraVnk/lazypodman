package podman

import (
	"context"
	"strings"
	"time"

	"github.com/containers/podman/v5/pkg/bindings/pods"
	entitiesTypes "github.com/containers/podman/v5/pkg/domain/entities/types"

	"github.com/jesseduffield/lazydocker/pkg/domain"
	"github.com/jesseduffield/lazydocker/pkg/runtime"
)

// Compile-time check that the Podman runtime advertises the optional
// PodRuntime capability.
var _ runtime.PodRuntime = (*Runtime)(nil)

// ListPods returns every pod known to Podman.
func (r *Runtime) ListPods(ctx context.Context) ([]domain.PodInfo, error) {
	conn, err := r.client()
	if err != nil {
		return nil, err
	}
	list, err := pods.List(conn, nil)
	if err != nil {
		return nil, mapErr("list pods", err)
	}
	out := make([]domain.PodInfo, 0, len(list))
	for _, p := range list {
		if p == nil {
			continue
		}
		out = append(out, podToDomain(p))
	}
	return out, nil
}

func (r *Runtime) StartPod(ctx context.Context, id string) error {
	conn, err := r.client()
	if err != nil {
		return err
	}
	_, err = pods.Start(conn, id, nil)
	return mapErr("start pod", err)
}

func (r *Runtime) StopPod(ctx context.Context, id string, timeout *time.Duration) error {
	conn, err := r.client()
	if err != nil {
		return err
	}
	opts := new(pods.StopOptions)
	if timeout != nil {
		opts = opts.WithTimeout(int(timeout.Seconds()))
	}
	_, err = pods.Stop(conn, id, opts)
	return mapErr("stop pod", err)
}

func (r *Runtime) RestartPod(ctx context.Context, id string) error {
	conn, err := r.client()
	if err != nil {
		return err
	}
	_, err = pods.Restart(conn, id, nil)
	return mapErr("restart pod", err)
}

func (r *Runtime) RemovePod(ctx context.Context, id string, force bool) error {
	conn, err := r.client()
	if err != nil {
		return err
	}
	_, err = pods.Remove(conn, id, new(pods.RemoveOptions).WithForce(force))
	return mapErr("remove pod", err)
}

func (r *Runtime) PrunePods(ctx context.Context) (domain.PruneReport, error) {
	conn, err := r.client()
	if err != nil {
		return domain.PruneReport{}, err
	}
	reps, err := pods.Prune(conn, nil)
	if err != nil {
		return domain.PruneReport{}, mapErr("prune pods", err)
	}
	var out domain.PruneReport
	for _, p := range reps {
		if p == nil {
			continue
		}
		out.ItemsDeleted = append(out.ItemsDeleted, p.Id)
	}
	return out, nil
}

func podToDomain(p *entitiesTypes.ListPodsReport) domain.PodInfo {
	info := domain.PodInfo{
		ID:        p.Id,
		Name:      p.Name,
		Namespace: p.Namespace,
		Status:    mapPodStatus(p.Status),
		Created:   p.Created,
		InfraID:   p.InfraId,
		Labels:    p.Labels,
		Networks:  p.Networks,
	}
	for _, c := range p.Containers {
		if c == nil {
			continue
		}
		info.Containers = append(info.Containers, domain.PodContainer{
			ID:     c.Id,
			Name:   c.Names,
			Status: c.Status,
		})
	}
	return info
}

func mapPodStatus(s string) domain.PodStatus {
	switch strings.ToLower(s) {
	case "created":
		return domain.PodStatusCreated
	case "running":
		return domain.PodStatusRunning
	case "stopped":
		return domain.PodStatusStopped
	case "exited":
		return domain.PodStatusExited
	case "paused":
		return domain.PodStatusPaused
	case "degraded":
		return domain.PodStatusDegraded
	case "dead":
		return domain.PodStatusDead
	default:
		return domain.PodStatusUnknown
	}
}
