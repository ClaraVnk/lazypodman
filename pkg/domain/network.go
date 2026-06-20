package domain

import "time"

// NetworkScope is the visibility of a network.
type NetworkScope string

const (
	NetworkScopeLocal  NetworkScope = "local"
	NetworkScopeGlobal NetworkScope = "global"
	NetworkScopeSwarm  NetworkScope = "swarm"
)

// IPAMConfig is one IP allocation rule attached to a network.
type IPAMConfig struct {
	Subnet  string
	Gateway string
	IPRange string
}

// IPAM is the IP address management configuration of a network.
type IPAM struct {
	Driver  string
	Options map[string]string
	Config  []IPAMConfig
}

// NetworkContainerAttachment is a container attached to a network.
type NetworkContainerAttachment struct {
	ContainerID string
	EndpointID  string
	Name        string
	IPv4Address string
	IPv6Address string
	MACAddress  string
}

// NetworkInfo is the view of a network as rendered by the GUI. Equivalent
// to docker's network.Inspect.
type NetworkInfo struct {
	ID         string
	Name       string
	Driver     string
	Scope      NetworkScope
	Created    time.Time
	IPAM       IPAM
	Internal   bool
	Attachable bool
	Ingress    bool
	EnableIPv6 bool
	Labels     map[string]string
	Options    map[string]string
	Containers map[string]NetworkContainerAttachment
}
