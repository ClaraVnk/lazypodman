package docker

import (
	"strings"
	"time"

	dockercontainer "github.com/docker/docker/api/types/container"
	dockerevents "github.com/docker/docker/api/types/events"
	dockerimage "github.com/docker/docker/api/types/image"
	dockernetwork "github.com/docker/docker/api/types/network"
	dockervolume "github.com/docker/docker/api/types/volume"

	"github.com/ClaraVnk/lazypodman/pkg/domain"
)

// containerSummaryToInfo converts a Docker SDK container.Summary to
// domain.ContainerInfo.
func containerSummaryToInfo(s dockercontainer.Summary) domain.ContainerInfo {
	return domain.ContainerInfo{
		ID:         s.ID,
		Names:      append([]string(nil), s.Names...),
		Image:      s.Image,
		ImageID:    s.ImageID,
		Command:    s.Command,
		Created:    time.Unix(s.Created, 0),
		State:      mapContainerState(s.State),
		Status:     s.Status,
		Ports:      mapPorts(s.Ports),
		Labels:     copyStringMap(s.Labels),
		SizeRw:     s.SizeRw,
		SizeRootFs: s.SizeRootFs,
	}
}

// containerInspectToDetails converts a Docker SDK container.InspectResponse
// to domain.ContainerDetails. The summary half is filled in from the
// inspect response itself (since list-time fields are not always present
// on the inspect payload, we reconstruct what we can).
func containerInspectToDetails(r dockercontainer.InspectResponse) domain.ContainerDetails {
	d := domain.ContainerDetails{
		ContainerInfo: domain.ContainerInfo{
			ID:     r.ID,
			Names:  []string{r.Name},
			Image:  r.Config.Image,
			Status: r.State.Status,
			State:  mapContainerState(r.State.Status),
			Labels: copyStringMap(r.Config.Labels),
		},
		Path:         r.Path,
		Args:         append([]string(nil), r.Args...),
		Config:       containerConfigToDomain(r.Config),
		Mounts:       mountsToDomain(r.Mounts),
		Health:       healthToDomain(r.State.Health),
		RestartCount: r.RestartCount,
		Platform:     r.Platform,
		ExitCode:     r.State.ExitCode,
	}

	if created, err := time.Parse(time.RFC3339Nano, r.Created); err == nil {
		d.Created = created
	}
	if started, err := time.Parse(time.RFC3339Nano, r.State.StartedAt); err == nil {
		d.StartedAt = started
	}
	if finished, err := time.Parse(time.RFC3339Nano, r.State.FinishedAt); err == nil {
		d.FinishedAt = finished
	}

	if r.NetworkSettings != nil {
		d.NetworkSettings = networkSettingsToDomain(r.NetworkSettings)
	}
	return d
}

// mapContainerState maps the docker state string to the domain enum.
func mapContainerState(s string) domain.ContainerState {
	switch strings.ToLower(s) {
	case "created":
		return domain.ContainerStateCreated
	case "running":
		return domain.ContainerStateRunning
	case "paused":
		return domain.ContainerStatePaused
	case "restarting":
		return domain.ContainerStateRestarting
	case "removing":
		return domain.ContainerStateRemoving
	case "exited":
		return domain.ContainerStateExited
	case "dead":
		return domain.ContainerStateDead
	default:
		return domain.ContainerStateUnknown
	}
}

// mapPorts converts docker SDK ports to domain ports.
func mapPorts(in []dockercontainer.Port) []domain.Port {
	if len(in) == 0 {
		return nil
	}
	out := make([]domain.Port, 0, len(in))
	for _, p := range in {
		out = append(out, domain.Port{
			HostIP:        p.IP,
			HostPort:      p.PublicPort,
			ContainerPort: p.PrivatePort,
			Protocol:      mapPortProtocol(p.Type),
		})
	}
	return out
}

func mapPortProtocol(s string) domain.PortProtocol {
	switch strings.ToLower(s) {
	case "udp":
		return domain.PortProtocolUDP
	case "sctp":
		return domain.PortProtocolSCTP
	default:
		return domain.PortProtocolTCP
	}
}

func containerConfigToDomain(c *dockercontainer.Config) domain.ContainerConfig {
	if c == nil {
		return domain.ContainerConfig{}
	}
	return domain.ContainerConfig{
		Image:      c.Image,
		Cmd:        []string(c.Cmd),
		Entrypoint: []string(c.Entrypoint),
		Env:        append([]string(nil), c.Env...),
		Labels:     copyStringMap(c.Labels),
		WorkingDir: c.WorkingDir,
		User:       c.User,
		Tty:        c.Tty,
		OpenStdin:  c.OpenStdin,
	}
}

func mountsToDomain(in []dockercontainer.MountPoint) []domain.Mount {
	if len(in) == 0 {
		return nil
	}
	out := make([]domain.Mount, 0, len(in))
	for _, m := range in {
		out = append(out, domain.Mount{
			Type:        mapMountType(string(m.Type)),
			Name:        m.Name,
			Source:      m.Source,
			Destination: m.Destination,
			Driver:      m.Driver,
			Mode:        m.Mode,
			ReadWrite:   m.RW,
		})
	}
	return out
}

func mapMountType(s string) domain.MountType {
	switch strings.ToLower(s) {
	case "bind":
		return domain.MountTypeBind
	case "volume":
		return domain.MountTypeVolume
	case "tmpfs":
		return domain.MountTypeTmpfs
	default:
		return domain.MountType(s)
	}
}

func healthToDomain(h *dockercontainer.Health) *domain.Health {
	if h == nil {
		return nil
	}
	out := &domain.Health{
		Status:        mapHealthStatus(h.Status),
		FailingStreak: h.FailingStreak,
	}
	if len(h.Log) > 0 {
		out.Log = make([]domain.HealthLogEntry, 0, len(h.Log))
		for _, entry := range h.Log {
			if entry == nil {
				continue
			}
			out.Log = append(out.Log, domain.HealthLogEntry{
				Start:    entry.Start,
				End:      entry.End,
				ExitCode: entry.ExitCode,
				Output:   entry.Output,
			})
		}
	}
	return out
}

func mapHealthStatus(s string) domain.HealthStatus {
	switch strings.ToLower(s) {
	case "starting":
		return domain.HealthStatusStarting
	case "healthy":
		return domain.HealthStatusHealthy
	case "unhealthy":
		return domain.HealthStatusUnhealthy
	default:
		return domain.HealthStatusNone
	}
}

func networkSettingsToDomain(ns *dockercontainer.NetworkSettings) domain.NetworkSettings {
	if ns == nil {
		return domain.NetworkSettings{}
	}
	out := domain.NetworkSettings{}
	if len(ns.Networks) > 0 {
		out.Endpoints = make(map[string]domain.NetworkEndpoint, len(ns.Networks))
		for name, ep := range ns.Networks {
			if ep == nil {
				continue
			}
			out.Endpoints[name] = domain.NetworkEndpoint{
				NetworkID:   ep.NetworkID,
				NetworkName: name,
				IPAddress:   ep.IPAddress,
				IPv6Address: ep.GlobalIPv6Address,
				MACAddress:  ep.MacAddress,
				Aliases:     append([]string(nil), ep.Aliases...),
			}
		}
	}
	if len(ns.Ports) > 0 {
		out.PortBindings = make(map[string][]domain.PortBinding, len(ns.Ports))
		for port, bindings := range ns.Ports {
			key := string(port)
			if len(bindings) == 0 {
				out.PortBindings[key] = nil
				continue
			}
			converted := make([]domain.PortBinding, 0, len(bindings))
			for _, b := range bindings {
				converted = append(converted, domain.PortBinding{
					HostIP:   b.HostIP,
					HostPort: b.HostPort,
				})
			}
			out.PortBindings[key] = converted
		}
	}
	return out
}

// imageSummaryToInfo converts a Docker SDK image.Summary to
// domain.ImageInfo.
func imageSummaryToInfo(s dockerimage.Summary) domain.ImageInfo {
	return domain.ImageInfo{
		ID:          s.ID,
		ParentID:    s.ParentID,
		RepoTags:    append([]string(nil), s.RepoTags...),
		RepoDigests: append([]string(nil), s.RepoDigests...),
		Created:     time.Unix(s.Created, 0),
		Size:        s.Size,
		SharedSize:  s.SharedSize,
		VirtualSize: s.Size, // VirtualSize is omitted in API v1.44+, mirror Size for backwards compat
		Labels:      copyStringMap(s.Labels),
		Containers:  s.Containers,
	}
}

// imageHistoryToDomain converts a Docker SDK image.HistoryResponseItem to
// domain.ImageHistoryItem.
func imageHistoryToDomain(h dockerimage.HistoryResponseItem) domain.ImageHistoryItem {
	return domain.ImageHistoryItem{
		ID:        h.ID,
		Created:   time.Unix(h.Created, 0),
		CreatedBy: h.CreatedBy,
		Size:      h.Size,
		Comment:   h.Comment,
		Tags:      append([]string(nil), h.Tags...),
	}
}

// networkInspectToDomain converts a Docker SDK network.Inspect to
// domain.NetworkInfo.
func networkInspectToDomain(n dockernetwork.Inspect) domain.NetworkInfo {
	return domain.NetworkInfo{
		ID:         n.ID,
		Name:       n.Name,
		Driver:     n.Driver,
		Scope:      mapNetworkScope(n.Scope),
		Created:    n.Created,
		IPAM:       ipamToDomain(n.IPAM),
		Internal:   n.Internal,
		Attachable: n.Attachable,
		Ingress:    n.Ingress,
		EnableIPv6: n.EnableIPv6,
		Labels:     copyStringMap(n.Labels),
		Options:    copyStringMap(n.Options),
		Containers: networkContainersToDomain(n.Containers),
	}
}

func mapNetworkScope(s string) domain.NetworkScope {
	switch strings.ToLower(s) {
	case "global":
		return domain.NetworkScopeGlobal
	case "swarm":
		return domain.NetworkScopeSwarm
	default:
		return domain.NetworkScopeLocal
	}
}

func ipamToDomain(in dockernetwork.IPAM) domain.IPAM {
	out := domain.IPAM{
		Driver:  in.Driver,
		Options: copyStringMap(in.Options),
	}
	if len(in.Config) > 0 {
		out.Config = make([]domain.IPAMConfig, 0, len(in.Config))
		for _, c := range in.Config {
			out.Config = append(out.Config, domain.IPAMConfig{
				Subnet:  c.Subnet,
				Gateway: c.Gateway,
				IPRange: c.IPRange,
			})
		}
	}
	return out
}

func networkContainersToDomain(in map[string]dockernetwork.EndpointResource) map[string]domain.NetworkContainerAttachment {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]domain.NetworkContainerAttachment, len(in))
	for id, ep := range in {
		out[id] = domain.NetworkContainerAttachment{
			ContainerID: id,
			EndpointID:  ep.EndpointID,
			Name:        ep.Name,
			IPv4Address: ep.IPv4Address,
			IPv6Address: ep.IPv6Address,
			MACAddress:  ep.MacAddress,
		}
	}
	return out
}

// volumeToDomain converts a Docker SDK *volume.Volume to domain.VolumeInfo.
func volumeToDomain(v *dockervolume.Volume) domain.VolumeInfo {
	if v == nil {
		return domain.VolumeInfo{}
	}
	out := domain.VolumeInfo{
		Name:       v.Name,
		Driver:     v.Driver,
		Mountpoint: v.Mountpoint,
		Scope:      mapVolumeScope(v.Scope),
		Labels:     copyStringMap(v.Labels),
		Options:    copyStringMap(v.Options),
	}
	if created, err := time.Parse(time.RFC3339Nano, v.CreatedAt); err == nil {
		out.CreatedAt = created
	}
	if len(v.Status) > 0 {
		out.Status = make(map[string]any, len(v.Status))
		for k, val := range v.Status {
			out.Status[k] = val
		}
	}
	if v.UsageData != nil {
		out.UsageData = &domain.VolumeUsage{
			Size:     v.UsageData.Size,
			RefCount: v.UsageData.RefCount,
		}
	}
	return out
}

func mapVolumeScope(s string) domain.VolumeScope {
	if strings.ToLower(s) == "global" {
		return domain.VolumeScopeGlobal
	}
	return domain.VolumeScopeLocal
}

// eventToDomain converts a Docker SDK events.Message to domain.Event.
func eventToDomain(m dockerevents.Message) domain.Event {
	return domain.Event{
		Type:    mapEventType(string(m.Type)),
		Action:  string(m.Action),
		ActorID: m.Actor.ID,
		Actor:   m.Actor.Attributes["name"],
		Scope:   m.Scope,
		Time:    time.Unix(m.Time, m.TimeNano%int64(time.Second)),
		Attrs:   copyStringMap(m.Actor.Attributes),
	}
}

func mapEventType(s string) domain.EventType {
	switch strings.ToLower(s) {
	case "container":
		return domain.EventTypeContainer
	case "image":
		return domain.EventTypeImage
	case "network":
		return domain.EventTypeNetwork
	case "volume":
		return domain.EventTypeVolume
	default:
		return domain.EventTypeSystem
	}
}

func copyStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
