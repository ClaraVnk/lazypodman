package podman

import (
	"context"
	"strings"
	"time"

	"github.com/containers/podman/v5/pkg/bindings/containers"

	"github.com/jesseduffield/lazydocker/pkg/domain"
	"github.com/jesseduffield/lazydocker/pkg/runtime"
)

// ListContainers returns every container (running and stopped).
func (r *Runtime) ListContainers(ctx context.Context) ([]domain.ContainerInfo, error) {
	list, err := containers.List(r.conn, new(containers.ListOptions).WithAll(true))
	if err != nil {
		return nil, mapErr("list containers", err)
	}
	out := make([]domain.ContainerInfo, 0, len(list))
	for i := range list {
		out = append(out, listContainerToDomain(list[i]))
	}
	return out, nil
}

func (r *Runtime) StartContainer(ctx context.Context, id string) error {
	return mapErr("start container", containers.Start(r.conn, id, nil))
}

func (r *Runtime) StopContainer(ctx context.Context, id string, timeout *time.Duration) error {
	opts := new(containers.StopOptions)
	if timeout != nil {
		opts = opts.WithTimeout(uint(timeout.Seconds()))
	}
	return mapErr("stop container", containers.Stop(r.conn, id, opts))
}

func (r *Runtime) RestartContainer(ctx context.Context, id string, timeout *time.Duration) error {
	opts := new(containers.RestartOptions)
	if timeout != nil {
		opts = opts.WithTimeout(int(timeout.Seconds()))
	}
	return mapErr("restart container", containers.Restart(r.conn, id, opts))
}

func (r *Runtime) PauseContainer(ctx context.Context, id string) error {
	return mapErr("pause container", containers.Pause(r.conn, id, nil))
}

func (r *Runtime) UnpauseContainer(ctx context.Context, id string) error {
	return mapErr("unpause container", containers.Unpause(r.conn, id, nil))
}

func (r *Runtime) RemoveContainer(ctx context.Context, id string, opts runtime.RemoveContainerOptions) error {
	o := new(containers.RemoveOptions).WithForce(opts.Force).WithVolumes(opts.RemoveVolumes)
	_, err := containers.Remove(r.conn, id, o)
	return mapErr("remove container", err)
}

// ContainerTop returns the container's process table. Podman returns the
// table as lines of tab-separated columns, the first line being headers.
func (r *Runtime) ContainerTop(ctx context.Context, id string) (domain.TopOutput, error) {
	lines, err := containers.Top(r.conn, id, nil)
	if err != nil {
		return domain.TopOutput{}, mapErr("container top", err)
	}
	return topToDomain(lines), nil
}

func (r *Runtime) PruneContainers(ctx context.Context) (domain.PruneReport, error) {
	reps, err := containers.Prune(r.conn, nil)
	if err != nil {
		return domain.PruneReport{}, mapErr("prune containers", err)
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

func topToDomain(lines []string) domain.TopOutput {
	if len(lines) == 0 {
		return domain.TopOutput{}
	}
	out := domain.TopOutput{Headers: strings.Split(lines[0], "\t")}
	for _, line := range lines[1:] {
		out.Processes = append(out.Processes, domain.TopProcess{Fields: strings.Split(line, "\t")})
	}
	return out
}
