package podman

import (
	"strconv"
	"strings"
	"time"

	"github.com/containers/podman/v5/libpod/define"
	handlersTypes "github.com/containers/podman/v5/pkg/api/handlers/types"
	entitiesTypes "github.com/containers/podman/v5/pkg/domain/entities/types"
	netTypes "go.podman.io/common/libnetwork/types"

	"github.com/ClaraVnk/lazypodman/pkg/domain"
)

// inspectContainerToDomain converts a Podman container inspect payload to
// the full detail view the inspect panel renders.
func inspectContainerToDomain(d *define.InspectContainerData) domain.ContainerDetails {
	info := domain.ContainerInfo{
		ID:         d.ID,
		Names:      []string{d.Name},
		ImageID:    d.Image,
		Created:    d.Created,
		SizeRootFs: d.SizeRootFs,
	}
	if d.SizeRw != nil {
		info.SizeRw = *d.SizeRw
	}

	details := domain.ContainerDetails{
		Path:         d.Path,
		Args:         d.Args,
		RestartCount: int(d.RestartCount),
	}

	if c := d.Config; c != nil {
		info.Image = c.Image
		info.Labels = c.Labels
		info.Command = strings.Join(c.Cmd, " ")
		details.Config = domain.ContainerConfig{
			Image:      c.Image,
			Cmd:        c.Cmd,
			Entrypoint: c.Entrypoint,
			Env:        c.Env,
			Labels:     c.Labels,
			WorkingDir: c.WorkingDir,
			User:       c.User,
			Tty:        c.Tty,
			OpenStdin:  c.OpenStdin,
		}
	}
	if s := d.State; s != nil {
		info.State = mapContainerState(s.Status)
		info.Status = s.Status
		details.StartedAt = s.StartedAt
		details.FinishedAt = s.FinishedAt
		details.ExitCode = int(s.ExitCode)
		details.Health = healthToDomain(s.Health)
	}
	if ns := d.NetworkSettings; ns != nil {
		details.NetworkSettings = inspectNetworkToDomain(ns)
		info.Ports = portsFromInspect(ns.Ports)
	}
	details.Mounts = inspectMountsToDomain(d.Mounts)

	details.ContainerInfo = info
	return details
}

func healthToDomain(h *define.HealthCheckResults) *domain.Health {
	if h == nil {
		return nil
	}
	out := &domain.Health{
		Status:        mapHealthStatus(h.Status),
		FailingStreak: h.FailingStreak,
	}
	for _, l := range h.Log {
		entry := domain.HealthLogEntry{ExitCode: l.ExitCode, Output: l.Output}
		if t, err := time.Parse(time.RFC3339Nano, l.Start); err == nil {
			entry.Start = t
		}
		if t, err := time.Parse(time.RFC3339Nano, l.End); err == nil {
			entry.End = t
		}
		out.Log = append(out.Log, entry)
	}
	return out
}

func mapHealthStatus(s string) domain.HealthStatus {
	switch strings.ToLower(s) {
	case "healthy":
		return domain.HealthStatusHealthy
	case "unhealthy":
		return domain.HealthStatusUnhealthy
	case "starting":
		return domain.HealthStatusStarting
	default:
		return domain.HealthStatusNone
	}
}

func inspectNetworkToDomain(ns *define.InspectNetworkSettings) domain.NetworkSettings {
	out := domain.NetworkSettings{}
	if len(ns.Networks) > 0 {
		out.Endpoints = make(map[string]domain.NetworkEndpoint, len(ns.Networks))
		for name, n := range ns.Networks {
			if n == nil {
				continue
			}
			out.Endpoints[name] = domain.NetworkEndpoint{
				NetworkID:   n.NetworkID,
				NetworkName: name,
				IPAddress:   n.IPAddress,
				IPv6Address: n.GlobalIPv6Address,
				MACAddress:  n.MacAddress,
				Aliases:     n.Aliases,
			}
		}
	}
	if len(ns.Ports) > 0 {
		out.PortBindings = make(map[string][]domain.PortBinding, len(ns.Ports))
		for key, hostPorts := range ns.Ports {
			bindings := make([]domain.PortBinding, 0, len(hostPorts))
			for _, hp := range hostPorts {
				bindings = append(bindings, domain.PortBinding{HostIP: hp.HostIP, HostPort: hp.HostPort})
			}
			out.PortBindings[key] = bindings
		}
	}
	return out
}

func inspectMountsToDomain(in []define.InspectMount) []domain.Mount {
	if len(in) == 0 {
		return nil
	}
	out := make([]domain.Mount, 0, len(in))
	for _, m := range in {
		out = append(out, domain.Mount{
			Type:        domain.MountType(m.Type),
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

// portsFromInspect parses the inspect "containerPort/proto" -> host bindings
// map into the flat port list the summary view renders.
func portsFromInspect(ports map[string][]define.InspectHostPort) []domain.Port {
	if len(ports) == 0 {
		return nil
	}
	var out []domain.Port
	for key, hostPorts := range ports {
		containerPort, proto := splitPortKey(key)
		if len(hostPorts) == 0 {
			out = append(out, domain.Port{ContainerPort: containerPort, Protocol: proto})
			continue
		}
		for _, hp := range hostPorts {
			p := domain.Port{HostIP: hp.HostIP, ContainerPort: containerPort, Protocol: proto}
			if hostPort, err := strconv.ParseUint(hp.HostPort, 10, 16); err == nil {
				p.HostPort = uint16(hostPort)
			}
			out = append(out, p)
		}
	}
	return out
}

func splitPortKey(key string) (uint16, domain.PortProtocol) {
	proto := domain.PortProtocolTCP
	numStr := key
	if i := strings.IndexByte(key, '/'); i >= 0 {
		numStr = key[:i]
		proto = domain.PortProtocol(strings.ToLower(key[i+1:]))
	}
	n, _ := strconv.ParseUint(numStr, 10, 16)
	return uint16(n), proto
}

// imageSummaryToDomain converts a Podman ImageSummary to the summary view
// the GUI list panel renders. Podman reports Created as a Unix timestamp.
func imageSummaryToDomain(s *entitiesTypes.ImageSummary) domain.ImageInfo {
	return domain.ImageInfo{
		ID:          s.ID,
		ParentID:    s.ParentId,
		RepoTags:    s.RepoTags,
		RepoDigests: s.RepoDigests,
		Created:     time.Unix(s.Created, 0),
		Size:        s.Size,
		SharedSize:  int64(s.SharedSize),
		VirtualSize: s.VirtualSize,
		Labels:      s.Labels,
		Containers:  int64(s.Containers),
	}
}

// networkToDomain converts a Podman (libnetwork) network to the GUI view.
// Podman networks are always local-scoped; the swarm-only Attachable and
// Ingress flags do not apply, and attached containers are only available
// via Inspect (not List), so Containers is left empty here.
func networkToDomain(n netTypes.Network) domain.NetworkInfo {
	return domain.NetworkInfo{
		ID:         n.ID,
		Name:       n.Name,
		Driver:     n.Driver,
		Scope:      domain.NetworkScopeLocal,
		Created:    n.Created,
		IPAM:       subnetsToIPAM(n.Subnets, n.IPAMOptions),
		Internal:   n.Internal,
		EnableIPv6: n.IPv6Enabled,
		Labels:     n.Labels,
		Options:    n.Options,
	}
}

func subnetsToIPAM(subnets []netTypes.Subnet, options map[string]string) domain.IPAM {
	ipam := domain.IPAM{
		Driver:  options["driver"],
		Options: options,
	}
	for _, s := range subnets {
		cfg := domain.IPAMConfig{Subnet: s.Subnet.String()}
		if len(s.Gateway) > 0 {
			cfg.Gateway = s.Gateway.String()
		}
		ipam.Config = append(ipam.Config, cfg)
	}
	return ipam
}

// volumeReportToDomain converts a Podman volume list entry to the GUI view.
// Podman does not report usage data in the list, so UsageData stays nil.
func volumeReportToDomain(v *entitiesTypes.VolumeListReport) domain.VolumeInfo {
	return domain.VolumeInfo{
		Name:       v.Name,
		Driver:     v.Driver,
		Mountpoint: v.Mountpoint,
		Scope:      mapVolumeScope(v.Scope),
		CreatedAt:  v.CreatedAt,
		Labels:     v.Labels,
		Options:    v.Options,
		Status:     v.Status,
	}
}

func mapVolumeScope(s string) domain.VolumeScope {
	if strings.EqualFold(s, "global") {
		return domain.VolumeScopeGlobal
	}
	return domain.VolumeScopeLocal
}

// historyToDomain converts one Podman image-history entry to domain.
func historyToDomain(h *handlersTypes.HistoryResponse) domain.ImageHistoryItem {
	return domain.ImageHistoryItem{
		ID:        h.ID,
		Created:   time.Unix(h.Created, 0),
		CreatedBy: h.CreatedBy,
		Size:      h.Size,
		Comment:   h.Comment,
		Tags:      h.Tags,
	}
}

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
