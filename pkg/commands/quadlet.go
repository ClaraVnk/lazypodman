package commands

import (
	"context"

	"github.com/ClaraVnk/lazypodman/pkg/domain"
	"github.com/ClaraVnk/lazypodman/pkg/runtime"
	"github.com/sirupsen/logrus"
)

// Quadlet is a Podman quadlet (a systemd unit generated from a source file).
// Quadlets are a Podman-native concept exposed only when the active runtime
// implements the runtime.QuadletManager capability; on the Docker backend
// there are none and the GUI hides the panel.
type Quadlet struct {
	Name          string
	Quadlet       domain.Quadlet
	OSCommand     *OSCommand
	Log           *logrus.Entry
	DockerCommand LimitedDockerCommand
	Runtime       runtime.QuadletManager
}

// QuadletsSupported reports whether the active runtime manages quadlets.
func (c *DockerCommand) QuadletsSupported() bool {
	_, ok := c.Runtime.(runtime.QuadletManager)
	return ok
}

// RefreshQuadlets returns the current quadlets, or nil when the runtime does
// not manage them (e.g. the Docker backend).
func (c *DockerCommand) RefreshQuadlets() ([]*Quadlet, error) {
	qm, ok := c.Runtime.(runtime.QuadletManager)
	if !ok {
		return nil, nil
	}
	infos, err := qm.ListQuadlets(context.Background())
	if err != nil {
		return nil, err
	}
	quadlets := make([]*Quadlet, len(infos))
	for i, info := range infos {
		quadlets[i] = &Quadlet{
			Name:          info.Name,
			Quadlet:       info,
			OSCommand:     c.OSCommand,
			Log:           c.Log,
			DockerCommand: c,
			Runtime:       qm,
		}
	}
	return quadlets, nil
}

// Start starts the quadlet's generated systemd unit.
func (q *Quadlet) Start() error {
	return q.Runtime.StartQuadlet(context.Background(), q.Quadlet.UnitName)
}

// Stop stops the quadlet's generated systemd unit.
func (q *Quadlet) Stop() error {
	return q.Runtime.StopQuadlet(context.Background(), q.Quadlet.UnitName)
}

// Restart restarts the quadlet's generated systemd unit.
func (q *Quadlet) Restart() error {
	return q.Runtime.RestartQuadlet(context.Background(), q.Quadlet.UnitName)
}
