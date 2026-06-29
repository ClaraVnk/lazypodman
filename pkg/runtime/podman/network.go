package podman

import (
	"context"

	"github.com/containers/podman/v5/pkg/bindings/network"

	"github.com/ClaraVnk/lazypodman/pkg/domain"
)

// ListNetworks returns every network known to Podman.
func (r *Runtime) ListNetworks(ctx context.Context) ([]domain.NetworkInfo, error) {
	conn, err := r.client()
	if err != nil {
		return nil, err
	}
	list, err := network.List(conn, nil)
	if err != nil {
		return nil, mapErr("list networks", err)
	}
	out := make([]domain.NetworkInfo, 0, len(list))
	for i := range list {
		out = append(out, networkToDomain(list[i]))
	}
	return out, nil
}

// RemoveNetwork deletes a network.
func (r *Runtime) RemoveNetwork(ctx context.Context, id string) error {
	conn, err := r.client()
	if err != nil {
		return err
	}
	_, err = network.Remove(conn, id, nil)
	return mapErr("remove network", err)
}

// PruneNetworks removes all unused networks.
func (r *Runtime) PruneNetworks(ctx context.Context) (domain.PruneReport, error) {
	conn, err := r.client()
	if err != nil {
		return domain.PruneReport{}, err
	}
	reps, err := network.Prune(conn, nil)
	if err != nil {
		return domain.PruneReport{}, mapErr("prune networks", err)
	}
	var out domain.PruneReport
	for _, p := range reps {
		if p == nil {
			continue
		}
		out.ItemsDeleted = append(out.ItemsDeleted, p.Name)
	}
	return out, nil
}
