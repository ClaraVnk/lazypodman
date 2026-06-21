package podman

import (
	"strings"

	entitiesTypes "github.com/containers/podman/v5/pkg/domain/entities/types"
	netTypes "go.podman.io/common/libnetwork/types"

	"github.com/jesseduffield/lazydocker/pkg/domain"
)

// listContainerToDomain converts a Podman ListContainer to the summary
// view the GUI renders.
func listContainerToDomain(c entitiesTypes.ListContainer) domain.ContainerInfo {
	info := domain.ContainerInfo{
		ID:      c.ID,
		Names:   c.Names,
		Image:   c.Image,
		ImageID: c.ImageID,
		Command: strings.Join(c.Command, " "),
		Created: c.Created,
		State:   mapContainerState(c.State),
		Status:  c.Status,
		Ports:   portMappingsToDomain(c.Ports),
		Labels:  c.Labels,
	}
	if c.Size != nil {
		info.SizeRw = c.Size.RwSize
		info.SizeRootFs = c.Size.RootFsSize
	}
	return info
}

// mapContainerState normalizes Podman's container state vocabulary onto
// the domain states.
func mapContainerState(s string) domain.ContainerState {
	switch strings.ToLower(s) {
	case "created", "configured", "initialized":
		return domain.ContainerStateCreated
	case "running":
		return domain.ContainerStateRunning
	case "paused":
		return domain.ContainerStatePaused
	case "restarting":
		return domain.ContainerStateRestarting
	case "removing", "stopping":
		return domain.ContainerStateRemoving
	case "exited", "stopped":
		return domain.ContainerStateExited
	case "dead":
		return domain.ContainerStateDead
	default:
		return domain.ContainerStateUnknown
	}
}

// portMappingsToDomain flattens Podman's port mappings (which carry a
// Range and a possibly comma-separated protocol) into one domain.Port per
// concrete host:container port pair.
func portMappingsToDomain(in []netTypes.PortMapping) []domain.Port {
	if len(in) == 0 {
		return nil
	}
	var out []domain.Port
	for _, pm := range in {
		count := pm.Range
		if count == 0 {
			count = 1
		}
		protocols := []string{string(domain.PortProtocolTCP)}
		if pm.Protocol != "" {
			protocols = strings.Split(pm.Protocol, ",")
		}
		for _, proto := range protocols {
			for i := uint16(0); i < count; i++ {
				out = append(out, domain.Port{
					HostIP:        pm.HostIP,
					HostPort:      pm.HostPort + i,
					ContainerPort: pm.ContainerPort + i,
					Protocol:      domain.PortProtocol(strings.ToLower(strings.TrimSpace(proto))),
				})
			}
		}
	}
	return out
}
