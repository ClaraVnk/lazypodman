package docker

import (
	"context"

	dockervolume "github.com/docker/docker/api/types/volume"

	"github.com/ClaraVnk/lazypodman/pkg/domain"
)

// ListVolumes returns every volume known to the daemon.
func (r *Runtime) ListVolumes(ctx context.Context) ([]domain.VolumeInfo, error) {
	resp, err := r.cli.VolumeList(ctx, dockervolume.ListOptions{})
	if err != nil {
		return nil, mapErr("list volumes", err)
	}
	out := make([]domain.VolumeInfo, 0, len(resp.Volumes))
	for _, v := range resp.Volumes {
		out = append(out, volumeToDomain(v))
	}
	return out, nil
}

// RemoveVolume deletes a volume.
func (r *Runtime) RemoveVolume(ctx context.Context, name string, force bool) error {
	return mapErr("remove volume", r.cli.VolumeRemove(ctx, name, force))
}

// PruneVolumes removes all unused volumes.
func (r *Runtime) PruneVolumes(ctx context.Context) (domain.PruneReport, error) {
	report, err := r.cli.VolumesPrune(ctx, dockerFilters())
	if err != nil {
		return domain.PruneReport{}, mapErr("prune volumes", err)
	}
	return domain.PruneReport{
		ItemsDeleted:   append([]string(nil), report.VolumesDeleted...),
		SpaceReclaimed: report.SpaceReclaimed,
	}, nil
}
