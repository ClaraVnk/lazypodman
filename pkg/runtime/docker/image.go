package docker

import (
	"context"

	dockerimage "github.com/docker/docker/api/types/image"

	"github.com/ClaraVnk/lazypodman/pkg/domain"
	"github.com/ClaraVnk/lazypodman/pkg/runtime"
)

// ListImages returns every image known to the daemon.
func (r *Runtime) ListImages(ctx context.Context) ([]domain.ImageInfo, error) {
	summaries, err := r.cli.ImageList(ctx, dockerimage.ListOptions{All: true})
	if err != nil {
		return nil, mapErr("list images", err)
	}
	out := make([]domain.ImageInfo, 0, len(summaries))
	for _, s := range summaries {
		out = append(out, imageSummaryToInfo(s))
	}
	return out, nil
}

// RemoveImage deletes an image.
func (r *Runtime) RemoveImage(ctx context.Context, id string, opts runtime.RemoveImageOptions) error {
	_, err := r.cli.ImageRemove(ctx, id, dockerimage.RemoveOptions{
		Force:         opts.Force,
		PruneChildren: !opts.NoPrune,
	})
	return mapErr("remove image", err)
}

// ImageHistory returns the build history of an image.
func (r *Runtime) ImageHistory(ctx context.Context, id string) ([]domain.ImageHistoryItem, error) {
	items, err := r.cli.ImageHistory(ctx, id)
	if err != nil {
		return nil, mapErr("image history", err)
	}
	out := make([]domain.ImageHistoryItem, 0, len(items))
	for _, h := range items {
		out = append(out, imageHistoryToDomain(h))
	}
	return out, nil
}

// PruneImages removes dangling images.
func (r *Runtime) PruneImages(ctx context.Context) (domain.PruneReport, error) {
	report, err := r.cli.ImagesPrune(ctx, dockerFilters())
	if err != nil {
		return domain.PruneReport{}, mapErr("prune images", err)
	}
	out := domain.PruneReport{
		SpaceReclaimed: report.SpaceReclaimed,
	}
	for _, d := range report.ImagesDeleted {
		switch {
		case d.Deleted != "":
			out.ItemsDeleted = append(out.ItemsDeleted, d.Deleted)
		case d.Untagged != "":
			out.ItemsDeleted = append(out.ItemsDeleted, d.Untagged)
		}
	}
	return out, nil
}
