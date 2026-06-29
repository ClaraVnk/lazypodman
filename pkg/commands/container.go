package commands

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/ClaraVnk/lazypodman/pkg/domain"
	"github.com/ClaraVnk/lazypodman/pkg/i18n"
	"github.com/ClaraVnk/lazypodman/pkg/runtime"
	"github.com/ClaraVnk/lazypodman/pkg/utils"
	"github.com/go-errors/errors"
	"github.com/sasha-s/go-deadlock"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

// Container is one container managed by lazypodman.
type Container struct {
	Name            string
	ServiceName     string
	ContainerNumber string // might make this an int in the future if need be

	// OneOff tells us if the container is just a job container or is actually
	// bound to the service.
	OneOff      bool
	ProjectName string
	ID          string
	Container   domain.ContainerInfo
	OSCommand   *OSCommand
	Log         *logrus.Entry
	StatHistory []*RecordedStats
	Details     domain.ContainerDetails
	// DetailsFetched is true once Inspect has populated Details at least
	// once. Replaces the upstream ContainerJSONBase != nil heuristic.
	DetailsFetched   bool
	MonitoringStats  bool
	ContainerCommand LimitedContainerCommand
	Runtime          runtime.ContainerRuntime
	Tr               *i18n.TranslationSet

	StatsMutex deadlock.Mutex
}

// Remove removes the container.
func (c *Container) Remove(options runtime.RemoveContainerOptions) error {
	c.Log.Warn(fmt.Sprintf("removing container %s", c.Name))
	err := c.Runtime.RemoveContainer(context.Background(), c.ID, options)
	if err != nil {
		if strings.Contains(err.Error(), "Stop the container before attempting removal or force remove") {
			return ComplexError{
				Code:    MustStopContainer,
				Message: err.Error(),
				frame:   xerrors.Caller(1),
			}
		}
		return err
	}
	return nil
}

// Start starts the container.
func (c *Container) Start() error {
	c.Log.Warn(fmt.Sprintf("starting container %s", c.Name))
	return c.Runtime.StartContainer(context.Background(), c.ID)
}

// Stop stops the container.
func (c *Container) Stop() error {
	c.Log.Warn(fmt.Sprintf("stopping container %s", c.Name))
	return c.Runtime.StopContainer(context.Background(), c.ID, nil)
}

// Pause pauses the container.
func (c *Container) Pause() error {
	c.Log.Warn(fmt.Sprintf("pausing container %s", c.Name))
	return c.Runtime.PauseContainer(context.Background(), c.ID)
}

// Unpause unpauses the container.
func (c *Container) Unpause() error {
	c.Log.Warn(fmt.Sprintf("unpausing container %s", c.Name))
	return c.Runtime.UnpauseContainer(context.Background(), c.ID)
}

// Restart restarts the container.
func (c *Container) Restart() error {
	c.Log.Warn(fmt.Sprintf("restarting container %s", c.Name))
	return c.Runtime.RestartContainer(context.Background(), c.ID, nil)
}

// Attach attaches to the container by spawning a `docker attach` process.
// Phase 1d keeps the interactive attach as a CLI exec, see ADR 0004.
func (c *Container) Attach() (*exec.Cmd, error) {
	if !c.DetailsLoaded() {
		return nil, errors.New(c.Tr.WaitingForContainerInfo)
	}
	if !c.Details.Config.OpenStdin {
		return nil, errors.New(c.Tr.UnattachableContainerError)
	}
	if c.Container.State == domain.ContainerStateExited {
		return nil, errors.New(c.Tr.CannotAttachStoppedContainerError)
	}

	c.Log.Warn(fmt.Sprintf("attaching to container %s", c.Name))
	cmd := c.OSCommand.NewCmd("docker", "attach", "--sig-proxy=false", c.ID)
	return cmd, nil
}

// Top returns the process table of the container. Errors if the container
// is not running.
func (c *Container) Top(ctx context.Context) (domain.TopOutput, error) {
	details, err := c.Inspect()
	if err != nil {
		return domain.TopOutput{}, err
	}
	if details.State != domain.ContainerStateRunning {
		return domain.TopOutput{}, errors.New("container is not running")
	}
	return c.Runtime.ContainerTop(ctx, c.ID)
}

// PruneContainers removes all stopped containers.
func (c *ContainerCommand) PruneContainers() error {
	_, err := c.Runtime.PruneContainers(context.Background())
	return err
}

// Inspect returns the details of the container.
func (c *Container) Inspect() (domain.ContainerDetails, error) {
	return c.Runtime.InspectContainer(context.Background(), c.ID)
}

// RenderTop returns the formatted process table of the container.
func (c *Container) RenderTop(ctx context.Context) (string, error) {
	result, err := c.Top(ctx)
	if err != nil {
		return "", err
	}
	rows := make([][]string, 0, len(result.Processes)+1)
	rows = append(rows, result.Headers)
	for _, p := range result.Processes {
		rows = append(rows, p.Fields)
	}
	return utils.RenderTable(rows)
}

// DetailsLoaded reports whether Inspect has populated Details at least
// once. Sometimes it takes a moment for a freshly-started container to
// have its details available.
func (c *Container) DetailsLoaded() bool {
	return c.DetailsFetched
}
