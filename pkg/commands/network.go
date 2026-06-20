package commands

import (
	"context"

	"github.com/jesseduffield/lazydocker/pkg/domain"
	"github.com/jesseduffield/lazydocker/pkg/runtime"
	"github.com/sirupsen/logrus"
)

// Network : a container engine network known to lazypodman.
type Network struct {
	Name          string
	Network       domain.NetworkInfo
	OSCommand     *OSCommand
	Log           *logrus.Entry
	DockerCommand LimitedDockerCommand
	Runtime       runtime.ContainerRuntime
}

// RefreshNetworks returns the current list of networks.
func (c *DockerCommand) RefreshNetworks() ([]*Network, error) {
	networks, err := c.Runtime.ListNetworks(context.Background())
	if err != nil {
		return nil, err
	}

	ownNetworks := make([]*Network, len(networks))
	for i, nw := range networks {
		ownNetworks[i] = &Network{
			Name:          nw.Name,
			Network:       nw,
			OSCommand:     c.OSCommand,
			Log:           c.Log,
			DockerCommand: c,
			Runtime:       c.Runtime,
		}
	}
	return ownNetworks, nil
}

// PruneNetworks removes all unused networks.
func (c *DockerCommand) PruneNetworks() error {
	_, err := c.Runtime.PruneNetworks(context.Background())
	return err
}

// Remove deletes the network.
func (n *Network) Remove() error {
	return n.Runtime.RemoveNetwork(context.Background(), n.Network.ID)
}
