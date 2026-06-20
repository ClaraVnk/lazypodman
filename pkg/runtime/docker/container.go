package docker

import (
	"context"
	"io"
	"time"

	dockercontainer "github.com/docker/docker/api/types/container"

	"github.com/jesseduffield/lazydocker/pkg/domain"
	"github.com/jesseduffield/lazydocker/pkg/runtime"
)

// ListContainers returns all containers (including stopped).
func (r *Runtime) ListContainers(ctx context.Context) ([]domain.ContainerInfo, error) {
	summaries, err := r.cli.ContainerList(ctx, dockercontainer.ListOptions{All: true})
	if err != nil {
		return nil, mapErr("list containers", err)
	}
	out := make([]domain.ContainerInfo, 0, len(summaries))
	for _, s := range summaries {
		out = append(out, containerSummaryToInfo(s))
	}
	return out, nil
}

// InspectContainer returns the full state of a single container.
func (r *Runtime) InspectContainer(ctx context.Context, id string) (domain.ContainerDetails, error) {
	resp, err := r.cli.ContainerInspect(ctx, id)
	if err != nil {
		return domain.ContainerDetails{}, mapErr("inspect container", err)
	}
	return containerInspectToDetails(resp), nil
}

// StartContainer starts a stopped container.
func (r *Runtime) StartContainer(ctx context.Context, id string) error {
	err := r.cli.ContainerStart(ctx, id, dockercontainer.StartOptions{})
	return mapErr("start container", err)
}

// StopContainer stops a running container, optionally with a grace period.
func (r *Runtime) StopContainer(ctx context.Context, id string, timeout *time.Duration) error {
	opts := dockercontainer.StopOptions{}
	if timeout != nil {
		seconds := int(timeout.Seconds())
		opts.Timeout = &seconds
	}
	return mapErr("stop container", r.cli.ContainerStop(ctx, id, opts))
}

// RestartContainer restarts a container, optionally with a grace period.
func (r *Runtime) RestartContainer(ctx context.Context, id string, timeout *time.Duration) error {
	opts := dockercontainer.StopOptions{}
	if timeout != nil {
		seconds := int(timeout.Seconds())
		opts.Timeout = &seconds
	}
	return mapErr("restart container", r.cli.ContainerRestart(ctx, id, opts))
}

// PauseContainer freezes the container's processes.
func (r *Runtime) PauseContainer(ctx context.Context, id string) error {
	return mapErr("pause container", r.cli.ContainerPause(ctx, id))
}

// UnpauseContainer resumes a paused container.
func (r *Runtime) UnpauseContainer(ctx context.Context, id string) error {
	return mapErr("unpause container", r.cli.ContainerUnpause(ctx, id))
}

// RemoveContainer deletes a container.
func (r *Runtime) RemoveContainer(ctx context.Context, id string, opts runtime.RemoveContainerOptions) error {
	dockerOpts := dockercontainer.RemoveOptions{
		Force:         opts.Force,
		RemoveVolumes: opts.RemoveVolumes,
		RemoveLinks:   opts.RemoveLinks,
	}
	return mapErr("remove container", r.cli.ContainerRemove(ctx, id, dockerOpts))
}

// ContainerLogs streams a container's logs. Caller must Close the reader.
func (r *Runtime) ContainerLogs(ctx context.Context, id string, opts runtime.LogOptions) (io.ReadCloser, error) {
	dockerOpts := dockercontainer.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     opts.Follow,
		Tail:       opts.Tail,
		Timestamps: opts.Timestamps,
	}
	if !opts.Since.IsZero() {
		dockerOpts.Since = opts.Since.Format(time.RFC3339Nano)
	}
	if !opts.Until.IsZero() {
		dockerOpts.Until = opts.Until.Format(time.RFC3339Nano)
	}
	rc, err := r.cli.ContainerLogs(ctx, id, dockerOpts)
	if err != nil {
		return nil, mapErr("container logs", err)
	}
	return rc, nil
}

// ContainerTop returns the process table of a running container.
func (r *Runtime) ContainerTop(ctx context.Context, id string) (domain.TopOutput, error) {
	resp, err := r.cli.ContainerTop(ctx, id, nil)
	if err != nil {
		return domain.TopOutput{}, mapErr("container top", err)
	}
	out := domain.TopOutput{
		Headers:   append([]string(nil), resp.Titles...),
		Processes: make([]domain.TopProcess, 0, len(resp.Processes)),
	}
	for _, p := range resp.Processes {
		out.Processes = append(out.Processes, domain.TopProcess{
			Fields: append([]string(nil), p...),
		})
	}
	return out, nil
}

// PruneContainers removes all stopped containers.
func (r *Runtime) PruneContainers(ctx context.Context) (domain.PruneReport, error) {
	report, err := r.cli.ContainersPrune(ctx, dockerFilters())
	if err != nil {
		return domain.PruneReport{}, mapErr("prune containers", err)
	}
	return domain.PruneReport{
		ItemsDeleted:   append([]string(nil), report.ContainersDeleted...),
		SpaceReclaimed: report.SpaceReclaimed,
	}, nil
}
