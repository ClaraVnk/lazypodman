package domain

import "time"

// ContainerState is the lifecycle state of a container, normalized across
// runtimes. Values match the OCI / Docker / Podman vocabulary.
type ContainerState string

const (
	ContainerStateCreated    ContainerState = "created"
	ContainerStateRunning    ContainerState = "running"
	ContainerStatePaused     ContainerState = "paused"
	ContainerStateRestarting ContainerState = "restarting"
	ContainerStateRemoving   ContainerState = "removing"
	ContainerStateExited     ContainerState = "exited"
	ContainerStateDead       ContainerState = "dead"
	ContainerStateUnknown    ContainerState = "unknown"
)

// PortProtocol is the L4 protocol of a published port.
type PortProtocol string

const (
	PortProtocolTCP  PortProtocol = "tcp"
	PortProtocolUDP  PortProtocol = "udp"
	PortProtocolSCTP PortProtocol = "sctp"
)

// Port is one published port mapping between the host and the container.
type Port struct {
	HostIP        string
	HostPort      uint16
	ContainerPort uint16
	Protocol      PortProtocol
}

// MountType identifies the kind of mount attached to a container.
type MountType string

const (
	MountTypeBind   MountType = "bind"
	MountTypeVolume MountType = "volume"
	MountTypeTmpfs  MountType = "tmpfs"
)

// Mount is a filesystem mount attached to a container.
type Mount struct {
	Type        MountType
	Name        string // volume name, empty for bind/tmpfs
	Source      string // host path for bind, volume name for volume
	Destination string // path inside the container
	Driver      string
	Mode        string
	ReadWrite   bool
}

// HealthStatus is the health-check verdict of a container.
type HealthStatus string

const (
	HealthStatusStarting  HealthStatus = "starting"
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusNone      HealthStatus = "none"
)

// Health is the current health-check result of a container.
type Health struct {
	Status        HealthStatus
	FailingStreak int
	Log           []HealthLogEntry
}

// HealthLogEntry is one health-check execution result.
type HealthLogEntry struct {
	Start    time.Time
	End      time.Time
	ExitCode int
	Output   string
}

// NetworkEndpoint is a container's attachment to a network.
type NetworkEndpoint struct {
	NetworkID   string
	NetworkName string
	IPAddress   string
	IPv6Address string
	MACAddress  string
	Aliases     []string
}

// NetworkSettings groups every network endpoint of a container.
type NetworkSettings struct {
	Endpoints map[string]NetworkEndpoint // keyed by network name
}

// ContainerConfig is the configuration a container was created with that
// the GUI surfaces (env, cmd, entrypoint, labels, etc.).
type ContainerConfig struct {
	Image      string
	Cmd        []string
	Entrypoint []string
	Env        []string
	Labels     map[string]string
	WorkingDir string
	User       string
	Tty        bool
	OpenStdin  bool
}

// ContainerInfo is the summary view of a container — what the list panel
// renders. Equivalent to docker's container.Summary, podman's
// ListContainer.
type ContainerInfo struct {
	ID         string
	Names      []string
	Image      string
	ImageID    string
	Command    string
	Created    time.Time
	State      ContainerState
	Status     string // free-form human label ("Up 3 hours", "Exited (0) 2 minutes ago"...)
	Ports      []Port
	Labels     map[string]string
	SizeRw     int64
	SizeRootFs int64
}

// PrimaryName returns the first name of a container without the leading
// slash that some runtimes add, or an empty string if no name is set.
func (c ContainerInfo) PrimaryName() string {
	if len(c.Names) == 0 {
		return ""
	}
	name := c.Names[0]
	if len(name) > 0 && name[0] == '/' {
		name = name[1:]
	}
	return name
}

// ContainerDetails is the full view of a container — what the inspect
// panel renders. Equivalent to docker's container.InspectResponse.
type ContainerDetails struct {
	ContainerInfo

	Path            string
	Args            []string
	Config          ContainerConfig
	NetworkSettings NetworkSettings
	Mounts          []Mount
	Health          *Health // nil when no healthcheck is defined
	RestartCount    int
	Platform        string
	StartedAt       time.Time
	FinishedAt      time.Time
	ExitCode        int
}

// TopProcess is one line of the container's process table.
type TopProcess struct {
	Fields []string
}

// TopOutput is the full process table of a running container.
type TopOutput struct {
	Headers   []string
	Processes []TopProcess
}

// PruneReport summarizes what a prune operation removed.
type PruneReport struct {
	ItemsDeleted   []string
	SpaceReclaimed uint64
}
