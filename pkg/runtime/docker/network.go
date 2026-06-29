//go:build docker

package docker

import (
	"context"

	dockernetwork "github.com/docker/docker/api/types/network"

	"github.com/ClaraVnk/lazypodman/pkg/domain"
)

// ListNetworks returns every network known to the daemon.
func (r *Runtime) ListNetworks(ctx context.Context) ([]domain.NetworkInfo, error) {
	summaries, err := r.cli.NetworkList(ctx, dockernetwork.ListOptions{})
	if err != nil {
		return nil, mapErr("list networks", err)
	}
	out := make([]domain.NetworkInfo, 0, len(summaries))
	for _, n := range summaries {
		out = append(out, networkInspectToDomain(n))
	}
	return out, nil
}

// RemoveNetwork deletes a network.
func (r *Runtime) RemoveNetwork(ctx context.Context, id string) error {
	return mapErr("remove network", r.cli.NetworkRemove(ctx, id))
}

// PruneNetworks removes all unused networks.
func (r *Runtime) PruneNetworks(ctx context.Context) (domain.PruneReport, error) {
	report, err := r.cli.NetworksPrune(ctx, dockerFilters())
	if err != nil {
		return domain.PruneReport{}, mapErr("prune networks", err)
	}
	return domain.PruneReport{
		ItemsDeleted: append([]string(nil), report.NetworksDeleted...),
	}, nil
}
