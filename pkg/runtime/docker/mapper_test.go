package docker

import (
	"testing"
	"time"

	dockercontainer "github.com/docker/docker/api/types/container"
	dockerimage "github.com/docker/docker/api/types/image"
	dockernetwork "github.com/docker/docker/api/types/network"
	dockervolume "github.com/docker/docker/api/types/volume"

	"github.com/jesseduffield/lazydocker/pkg/domain"
)

func TestContainerSummaryToInfo(t *testing.T) {
	s := dockercontainer.Summary{
		ID:      "abc123",
		Names:   []string{"/foo", "/bar"},
		Image:   "nginx:latest",
		ImageID: "sha256:def",
		Command: "nginx -g daemon off;",
		Created: 1700000000,
		State:   "running",
		Status:  "Up 5 minutes",
		Ports: []dockercontainer.Port{
			{IP: "0.0.0.0", PublicPort: 8080, PrivatePort: 80, Type: "tcp"},
			{PrivatePort: 9090, Type: "udp"},
		},
		Labels:     map[string]string{"app": "web"},
		SizeRw:     1024,
		SizeRootFs: 2048,
	}

	got := containerSummaryToInfo(s)

	if got.ID != "abc123" {
		t.Errorf("ID = %q, want abc123", got.ID)
	}
	if got.State != domain.ContainerStateRunning {
		t.Errorf("State = %q, want running", got.State)
	}
	if got.PrimaryName() != "foo" {
		t.Errorf("PrimaryName = %q, want foo", got.PrimaryName())
	}
	if len(got.Ports) != 2 {
		t.Fatalf("len(Ports) = %d, want 2", len(got.Ports))
	}
	if got.Ports[0].HostPort != 8080 || got.Ports[0].Protocol != domain.PortProtocolTCP {
		t.Errorf("port[0] = %+v, want TCP 8080:80", got.Ports[0])
	}
	if got.Ports[1].Protocol != domain.PortProtocolUDP {
		t.Errorf("port[1].Protocol = %q, want udp", got.Ports[1].Protocol)
	}
	if got.Created.Unix() != 1700000000 {
		t.Errorf("Created = %v, want 1700000000", got.Created.Unix())
	}
	if got.Labels["app"] != "web" {
		t.Errorf("Labels lost: %v", got.Labels)
	}
}

func TestMapContainerState(t *testing.T) {
	cases := map[string]domain.ContainerState{
		"created":    domain.ContainerStateCreated,
		"RUNNING":    domain.ContainerStateRunning, // case-insensitive
		"paused":     domain.ContainerStatePaused,
		"restarting": domain.ContainerStateRestarting,
		"removing":   domain.ContainerStateRemoving,
		"exited":     domain.ContainerStateExited,
		"dead":       domain.ContainerStateDead,
		"":           domain.ContainerStateUnknown,
		"gibberish":  domain.ContainerStateUnknown,
	}
	for in, want := range cases {
		if got := mapContainerState(in); got != want {
			t.Errorf("mapContainerState(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestContainerInspectToDetails(t *testing.T) {
	created := "2026-06-20T10:00:00.000000000Z"
	started := "2026-06-20T10:01:00.000000000Z"
	finished := "2026-06-20T10:05:00.000000000Z"

	resp := dockercontainer.InspectResponse{
		ContainerJSONBase: &dockercontainer.ContainerJSONBase{
			ID:           "id",
			Name:         "/test",
			Created:      created,
			Path:         "/bin/sh",
			Args:         []string{"-c", "echo hi"},
			RestartCount: 3,
			Platform:     "linux/amd64",
			State: &dockercontainer.State{
				Status:     "exited",
				ExitCode:   0,
				StartedAt:  started,
				FinishedAt: finished,
				Health: &dockercontainer.Health{
					Status:        "unhealthy",
					FailingStreak: 2,
					Log: []*dockercontainer.HealthcheckResult{
						{Start: time.Unix(1, 0), End: time.Unix(2, 0), ExitCode: 1, Output: "bad"},
					},
				},
			},
			HostConfig: &dockercontainer.HostConfig{},
		},
		Mounts: []dockercontainer.MountPoint{
			{Type: "bind", Source: "/src", Destination: "/dst", RW: true},
			{Type: "volume", Name: "data", Source: "data", Destination: "/var/lib"},
		},
		Config: &dockercontainer.Config{
			Image: "alpine",
			Cmd:   []string{"sh"},
			Env:   []string{"FOO=bar"},
		},
		NetworkSettings: &dockercontainer.NetworkSettings{
			Networks: map[string]*dockernetwork.EndpointSettings{
				"bridge": {NetworkID: "net-id", IPAddress: "172.17.0.2", MacAddress: "aa:bb"},
			},
		},
	}

	got := containerInspectToDetails(resp)

	if got.ID != "id" {
		t.Errorf("ID = %q", got.ID)
	}
	if got.RestartCount != 3 {
		t.Errorf("RestartCount = %d, want 3", got.RestartCount)
	}
	if got.Health == nil || got.Health.Status != domain.HealthStatusUnhealthy {
		t.Errorf("Health not mapped: %+v", got.Health)
	}
	if len(got.Health.Log) != 1 || got.Health.Log[0].ExitCode != 1 {
		t.Errorf("Health log not mapped: %+v", got.Health.Log)
	}
	if len(got.Mounts) != 2 {
		t.Fatalf("len(Mounts) = %d, want 2", len(got.Mounts))
	}
	if got.Mounts[0].Type != domain.MountTypeBind {
		t.Errorf("Mounts[0].Type = %q", got.Mounts[0].Type)
	}
	if got.Mounts[1].Type != domain.MountTypeVolume || got.Mounts[1].Name != "data" {
		t.Errorf("Mounts[1] = %+v", got.Mounts[1])
	}
	if len(got.NetworkSettings.Endpoints) != 1 {
		t.Fatalf("Endpoints map = %+v", got.NetworkSettings.Endpoints)
	}
	if ep, ok := got.NetworkSettings.Endpoints["bridge"]; !ok || ep.IPAddress != "172.17.0.2" {
		t.Errorf("bridge endpoint not mapped: %+v", got.NetworkSettings.Endpoints)
	}
	if !got.StartedAt.Equal(time.Date(2026, 6, 20, 10, 1, 0, 0, time.UTC)) {
		t.Errorf("StartedAt = %v", got.StartedAt)
	}
	if !got.FinishedAt.Equal(time.Date(2026, 6, 20, 10, 5, 0, 0, time.UTC)) {
		t.Errorf("FinishedAt = %v", got.FinishedAt)
	}
}

func TestImageSummaryToInfo(t *testing.T) {
	s := dockerimage.Summary{
		ID:          "img1",
		ParentID:    "parent",
		RepoTags:    []string{"nginx:1.25"},
		RepoDigests: []string{"nginx@sha256:abc"},
		Created:     1700000000,
		Size:        100,
		SharedSize:  50,
		VirtualSize: 120,
		Labels:      map[string]string{"k": "v"},
		Containers:  -1,
	}
	got := imageSummaryToInfo(s)
	if got.ID != "img1" || got.ParentID != "parent" {
		t.Errorf("IDs wrong: %+v", got)
	}
	if got.Containers != -1 {
		t.Errorf("Containers = %d, want -1", got.Containers)
	}
	if got.RepoTags[0] != "nginx:1.25" {
		t.Errorf("RepoTags = %v", got.RepoTags)
	}
}

func TestVolumeToDomain(t *testing.T) {
	v := &dockervolume.Volume{
		Name:       "vol1",
		Driver:     "local",
		Mountpoint: "/var/lib/docker/volumes/vol1/_data",
		Scope:      "local",
		CreatedAt:  "2026-06-20T10:00:00Z",
		Labels:     map[string]string{"k": "v"},
		Status:     map[string]any{"foo": "bar"},
		UsageData: &dockervolume.UsageData{
			Size:     1024,
			RefCount: 2,
		},
	}
	got := volumeToDomain(v)
	if got.Name != "vol1" || got.Scope != domain.VolumeScopeLocal {
		t.Errorf("basic fields wrong: %+v", got)
	}
	if got.UsageData == nil || got.UsageData.Size != 1024 {
		t.Errorf("UsageData = %+v", got.UsageData)
	}
	if got.Status["foo"] != "bar" {
		t.Errorf("Status = %v", got.Status)
	}

	// nil safety: must not panic.
	_ = volumeToDomain(nil)
}

func TestNetworkInspectToDomain(t *testing.T) {
	n := dockernetwork.Inspect{
		ID:         "net1",
		Name:       "bridge",
		Driver:     "bridge",
		Scope:      "local",
		Internal:   false,
		Attachable: true,
		EnableIPv6: false,
		IPAM: dockernetwork.IPAM{
			Driver:  "default",
			Options: map[string]string{"x": "y"},
			Config: []dockernetwork.IPAMConfig{
				{Subnet: "172.17.0.0/16", Gateway: "172.17.0.1"},
			},
		},
		Containers: map[string]dockernetwork.EndpointResource{
			"abc": {Name: "web", IPv4Address: "172.17.0.2/16"},
		},
	}
	got := networkInspectToDomain(n)
	if got.ID != "net1" || got.Scope != domain.NetworkScopeLocal {
		t.Errorf("basic fields wrong: %+v", got)
	}
	if len(got.IPAM.Config) != 1 || got.IPAM.Config[0].Subnet != "172.17.0.0/16" {
		t.Errorf("IPAM not mapped: %+v", got.IPAM)
	}
	if got.Containers["abc"].Name != "web" {
		t.Errorf("Containers = %+v", got.Containers)
	}
}
