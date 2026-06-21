package podman

import (
	"context"

	"github.com/containers/podman/v5/pkg/bindings/volumes"

	"github.com/jesseduffield/lazydocker/pkg/domain"
)

// ListVolumes returns every volume known to Podman.
func (r *Runtime) ListVolumes(ctx context.Context) ([]domain.VolumeInfo, error) {
	list, err := volumes.List(r.conn, nil)
	if err != nil {
		return nil, mapErr("list volumes", err)
	}
	out := make([]domain.VolumeInfo, 0, len(list))
	for _, v := range list {
		if v == nil {
			continue
		}
		out = append(out, volumeReportToDomain(v))
	}
	return out, nil
}

// RemoveVolume deletes a volume.
func (r *Runtime) RemoveVolume(ctx context.Context, name string, force bool) error {
	o := new(volumes.RemoveOptions).WithForce(force)
	return mapErr("remove volume", volumes.Remove(r.conn, name, o))
}

// PruneVolumes removes all unused volumes.
func (r *Runtime) PruneVolumes(ctx context.Context) (domain.PruneReport, error) {
	reps, err := volumes.Prune(r.conn, nil)
	if err != nil {
		return domain.PruneReport{}, mapErr("prune volumes", err)
	}
	var out domain.PruneReport
	for _, p := range reps {
		if p == nil {
			continue
		}
		out.ItemsDeleted = append(out.ItemsDeleted, p.Id)
		out.SpaceReclaimed += p.Size
	}
	return out, nil
}
