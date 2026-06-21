package domain

import "time"

// PodStatus is the lifecycle state of a Podman pod.
type PodStatus string

const (
	PodStatusCreated  PodStatus = "created"
	PodStatusRunning  PodStatus = "running"
	PodStatusStopped  PodStatus = "stopped"
	PodStatusExited   PodStatus = "exited"
	PodStatusPaused   PodStatus = "paused"
	PodStatusDegraded PodStatus = "degraded"
	PodStatusDead     PodStatus = "dead"
	PodStatusUnknown  PodStatus = "unknown"
)

// PodContainer is a container that belongs to a pod.
type PodContainer struct {
	ID     string
	Name   string
	Status string
}

// PodInfo is the summary view of a Podman pod as rendered by the GUI.
// Pods are a Podman-native concept; see the PodRuntime capability in
// pkg/runtime and docs/adr/0006-podman-pods-and-capabilities.md.
type PodInfo struct {
	ID         string
	Name       string
	Namespace  string
	Status     PodStatus
	Created    time.Time
	InfraID    string
	Labels     map[string]string
	Networks   []string
	Containers []PodContainer
}
