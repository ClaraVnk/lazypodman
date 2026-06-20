package commands

import (
	"context"

	"github.com/jesseduffield/lazydocker/pkg/domain"
	"github.com/jesseduffield/lazydocker/pkg/runtime"
	"github.com/sirupsen/logrus"
)

// Volume : a container engine volume known to lazypodman.
type Volume struct {
	Name          string
	Volume        domain.VolumeInfo
	OSCommand     *OSCommand
	Log           *logrus.Entry
	DockerCommand LimitedDockerCommand
	Runtime       runtime.ContainerRuntime
}

// RefreshVolumes returns the current list of volumes.
func (c *DockerCommand) RefreshVolumes() ([]*Volume, error) {
	volumes, err := c.Runtime.ListVolumes(context.Background())
	if err != nil {
		return nil, err
	}

	ownVolumes := make([]*Volume, len(volumes))
	for i, vol := range volumes {
		ownVolumes[i] = &Volume{
			Name:          vol.Name,
			Volume:        vol,
			OSCommand:     c.OSCommand,
			Log:           c.Log,
			DockerCommand: c,
			Runtime:       c.Runtime,
		}
	}
	return ownVolumes, nil
}

// PruneVolumes removes all unused volumes.
func (c *DockerCommand) PruneVolumes() error {
	_, err := c.Runtime.PruneVolumes(context.Background())
	return err
}

// Remove deletes the volume.
func (v *Volume) Remove(force bool) error {
	return v.Runtime.RemoveVolume(context.Background(), v.Name, force)
}
