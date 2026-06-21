package commands

import (
	"context"

	"github.com/jesseduffield/lazydocker/pkg/domain"
	"github.com/jesseduffield/lazydocker/pkg/runtime"
	"github.com/sirupsen/logrus"
)

// Pod is a Podman pod. Pods are a Podman-native concept exposed only when
// the active runtime implements the runtime.PodRuntime capability; on the
// Docker backend there are none and the GUI hides the panel.
type Pod struct {
	Name          string
	Pod           domain.PodInfo
	OSCommand     *OSCommand
	Log           *logrus.Entry
	DockerCommand LimitedDockerCommand
	Runtime       runtime.PodRuntime
}

// PodsSupported reports whether the active runtime exposes pods.
func (c *DockerCommand) PodsSupported() bool {
	_, ok := c.Runtime.(runtime.PodRuntime)
	return ok
}

// RefreshPods returns the current list of pods, or nil when the runtime
// does not support pods (e.g. the Docker backend).
func (c *DockerCommand) RefreshPods() ([]*Pod, error) {
	pr, ok := c.Runtime.(runtime.PodRuntime)
	if !ok {
		return nil, nil
	}
	infos, err := pr.ListPods(context.Background())
	if err != nil {
		return nil, err
	}
	ownPods := make([]*Pod, len(infos))
	for i, info := range infos {
		ownPods[i] = &Pod{
			Name:          info.Name,
			Pod:           info,
			OSCommand:     c.OSCommand,
			Log:           c.Log,
			DockerCommand: c,
			Runtime:       pr,
		}
	}
	return ownPods, nil
}

// PrunePods removes all stopped pods.
func (c *DockerCommand) PrunePods() error {
	pr, ok := c.Runtime.(runtime.PodRuntime)
	if !ok {
		return nil
	}
	_, err := pr.PrunePods(context.Background())
	return err
}

// Remove removes the pod and its containers.
func (p *Pod) Remove(force bool) error {
	return p.Runtime.RemovePod(context.Background(), p.Pod.ID, force)
}
