package podman

import (
	"context"
	"errors"

	"github.com/containers/podman/v5/pkg/bindings/images"

	"github.com/jesseduffield/lazydocker/pkg/domain"
	"github.com/jesseduffield/lazydocker/pkg/runtime"
)

// ListImages returns the final (non-intermediate) images known to Podman.
func (r *Runtime) ListImages(ctx context.Context) ([]domain.ImageInfo, error) {
	conn, err := r.client()
	if err != nil {
		return nil, err
	}
	list, err := images.List(conn, new(images.ListOptions).WithAll(false))
	if err != nil {
		return nil, mapErr("list images", err)
	}
	out := make([]domain.ImageInfo, 0, len(list))
	for _, s := range list {
		if s == nil {
			continue
		}
		out = append(out, imageSummaryToDomain(s))
	}
	return out, nil
}

// RemoveImage deletes a single image. Podman's Remove takes a slice and
// returns a slice of errors, which we join.
func (r *Runtime) RemoveImage(ctx context.Context, id string, opts runtime.RemoveImageOptions) error {
	conn, err := r.client()
	if err != nil {
		return err
	}
	o := new(images.RemoveOptions).WithForce(opts.Force)
	_, errs := images.Remove(conn, []string{id}, o)
	return mapErr("remove image", errors.Join(errs...))
}

// ImageHistory returns the build history of an image.
func (r *Runtime) ImageHistory(ctx context.Context, id string) ([]domain.ImageHistoryItem, error) {
	conn, err := r.client()
	if err != nil {
		return nil, err
	}
	hist, err := images.History(conn, id, nil)
	if err != nil {
		return nil, mapErr("image history", err)
	}
	out := make([]domain.ImageHistoryItem, 0, len(hist))
	for _, h := range hist {
		if h == nil {
			continue
		}
		out = append(out, historyToDomain(h))
	}
	return out, nil
}

// PruneImages removes dangling images.
func (r *Runtime) PruneImages(ctx context.Context) (domain.PruneReport, error) {
	conn, err := r.client()
	if err != nil {
		return domain.PruneReport{}, err
	}
	reps, err := images.Prune(conn, nil)
	if err != nil {
		return domain.PruneReport{}, mapErr("prune images", err)
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
