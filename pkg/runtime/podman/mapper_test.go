package podman

import (
	"net"
	"testing"
	"time"

	"github.com/containers/podman/v5/libpod/define"
	handlersTypes "github.com/containers/podman/v5/pkg/api/handlers/types"
	entitiesTypes "github.com/containers/podman/v5/pkg/domain/entities/types"
	psDefine "github.com/containers/podman/v5/pkg/ps/define"
	netTypes "go.podman.io/common/libnetwork/types"

	"github.com/ClaraVnk/lazypodman/pkg/domain"
)

func TestMapContainerState(t *testing.T) {
	cases := map[string]domain.ContainerState{
		"created":     domain.ContainerStateCreated,
		"configured":  domain.ContainerStateCreated, // podman-specific
		"initialized": domain.ContainerStateCreated, // podman-specific
		"RUNNING":     domain.ContainerStateRunning, // case-insensitive
		"paused":      domain.ContainerStatePaused,
		"restarting":  domain.ContainerStateRestarting,
		"removing":    domain.ContainerStateRemoving,
		"stopping":    domain.ContainerStateRemoving, // podman-specific
		"exited":      domain.ContainerStateExited,
		"stopped":     domain.ContainerStateExited, // podman-specific
		"dead":        domain.ContainerStateDead,
		"":            domain.ContainerStateUnknown,
		"gibberish":   domain.ContainerStateUnknown,
	}
	for in, want := range cases {
		if got := mapContainerState(in); got != want {
			t.Errorf("mapContainerState(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMapHealthStatus(t *testing.T) {
	cases := map[string]domain.HealthStatus{
		"healthy":   domain.HealthStatusHealthy,
		"UNHEALTHY": domain.HealthStatusUnhealthy, // case-insensitive
		"starting":  domain.HealthStatusStarting,
		"":          domain.HealthStatusNone,
		"weird":     domain.HealthStatusNone,
	}
	for in, want := range cases {
		if got := mapHealthStatus(in); got != want {
			t.Errorf("mapHealthStatus(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMapVolumeScope(t *testing.T) {
	cases := map[string]domain.VolumeScope{
		"global": domain.VolumeScopeGlobal,
		"GLOBAL": domain.VolumeScopeGlobal, // case-insensitive
		"local":  domain.VolumeScopeLocal,
		"":       domain.VolumeScopeLocal, // default
	}
	for in, want := range cases {
		if got := mapVolumeScope(in); got != want {
			t.Errorf("mapVolumeScope(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestListContainerToDomain(t *testing.T) {
	created := time.Unix(1700000000, 0)
	c := entitiesTypes.ListContainer{
		ID:      "abc123",
		Names:   []string{"web", "web-alias"},
		Image:   "nginx:latest",
		ImageID: "sha256:def",
		Command: []string{"nginx", "-g", "daemon off;"},
		Created: created,
		State:   "running",
		Status:  "Up 5 minutes",
		Ports: []netTypes.PortMapping{
			{HostIP: "0.0.0.0", HostPort: 8080, ContainerPort: 80, Protocol: "tcp"},
		},
		Labels: map[string]string{"app": "web"},
		Size:   &psDefine.ContainerSize{RwSize: 1024, RootFsSize: 2048},
	}

	got := listContainerToDomain(c)

	if got.ID != "abc123" {
		t.Errorf("ID = %q, want abc123", got.ID)
	}
	if got.State != domain.ContainerStateRunning {
		t.Errorf("State = %q, want running", got.State)
	}
	if got.Command != "nginx -g daemon off;" {
		t.Errorf("Command = %q, want the joined command", got.Command)
	}
	if got.PrimaryName() != "web" {
		t.Errorf("PrimaryName = %q, want web", got.PrimaryName())
	}
	if got.SizeRw != 1024 || got.SizeRootFs != 2048 {
		t.Errorf("sizes = (%d, %d), want (1024, 2048)", got.SizeRw, got.SizeRootFs)
	}
	if !got.Created.Equal(created) {
		t.Errorf("Created = %v, want %v", got.Created, created)
	}
	if len(got.Ports) != 1 || got.Ports[0].HostPort != 8080 {
		t.Errorf("Ports = %+v, want one 8080 mapping", got.Ports)
	}
	if got.Labels["app"] != "web" {
		t.Errorf("Labels lost: %v", got.Labels)
	}

	// Size is optional: a nil pointer must not panic and leaves zero sizes.
	noSize := listContainerToDomain(entitiesTypes.ListContainer{ID: "x", State: "exited"})
	if noSize.SizeRw != 0 || noSize.SizeRootFs != 0 {
		t.Errorf("nil Size should leave zero sizes, got (%d, %d)", noSize.SizeRw, noSize.SizeRootFs)
	}
}

func TestPortMappingsToDomain(t *testing.T) {
	t.Run("nil is nil", func(t *testing.T) {
		if got := portMappingsToDomain(nil); got != nil {
			t.Errorf("got %+v, want nil", got)
		}
	})

	t.Run("range expands to consecutive ports", func(t *testing.T) {
		got := portMappingsToDomain([]netTypes.PortMapping{
			{HostIP: "127.0.0.1", HostPort: 8080, ContainerPort: 80, Range: 3, Protocol: "tcp"},
		})
		if len(got) != 3 {
			t.Fatalf("len = %d, want 3 (range expansion)", len(got))
		}
		for i, p := range got {
			wantHost := uint16(8080 + i)
			wantCtr := uint16(80 + i)
			if p.HostPort != wantHost || p.ContainerPort != wantCtr {
				t.Errorf("port[%d] = %d:%d, want %d:%d", i, p.HostPort, p.ContainerPort, wantHost, wantCtr)
			}
			if p.Protocol != domain.PortProtocolTCP {
				t.Errorf("port[%d].Protocol = %q, want tcp", i, p.Protocol)
			}
		}
	})

	t.Run("range 0 means a single port", func(t *testing.T) {
		got := portMappingsToDomain([]netTypes.PortMapping{
			{HostPort: 5000, ContainerPort: 5000, Range: 0, Protocol: "udp"},
		})
		if len(got) != 1 || got[0].Protocol != domain.PortProtocolUDP {
			t.Fatalf("got %+v, want one udp port", got)
		}
	})

	t.Run("comma-separated protocols fan out", func(t *testing.T) {
		got := portMappingsToDomain([]netTypes.PortMapping{
			{HostPort: 53, ContainerPort: 53, Protocol: "tcp,udp"},
		})
		if len(got) != 2 {
			t.Fatalf("len = %d, want 2 (one per protocol)", len(got))
		}
		protos := map[domain.PortProtocol]bool{got[0].Protocol: true, got[1].Protocol: true}
		if !protos[domain.PortProtocolTCP] || !protos[domain.PortProtocolUDP] {
			t.Errorf("protocols = %v, want both tcp and udp", protos)
		}
	})

	t.Run("empty protocol defaults to tcp", func(t *testing.T) {
		got := portMappingsToDomain([]netTypes.PortMapping{
			{HostPort: 1, ContainerPort: 1},
		})
		if len(got) != 1 || got[0].Protocol != domain.PortProtocolTCP {
			t.Errorf("got %+v, want a single tcp port", got)
		}
	})
}

func TestSplitPortKey(t *testing.T) {
	cases := []struct {
		key       string
		wantPort  uint16
		wantProto domain.PortProtocol
	}{
		{"80/tcp", 80, domain.PortProtocolTCP},
		{"53/UDP", 53, domain.PortProtocolUDP}, // proto lowercased
		{"443", 443, domain.PortProtocolTCP},   // no proto defaults to tcp
		{"bad/tcp", 0, domain.PortProtocolTCP}, // unparseable port -> 0
	}
	for _, tc := range cases {
		gotPort, gotProto := splitPortKey(tc.key)
		if gotPort != tc.wantPort || gotProto != tc.wantProto {
			t.Errorf("splitPortKey(%q) = (%d, %q), want (%d, %q)", tc.key, gotPort, gotProto, tc.wantPort, tc.wantProto)
		}
	}
}

func TestPortsFromInspect(t *testing.T) {
	t.Run("nil is nil", func(t *testing.T) {
		if got := portsFromInspect(nil); got != nil {
			t.Errorf("got %+v, want nil", got)
		}
	})

	t.Run("exposed port without host binding", func(t *testing.T) {
		got := portsFromInspect(map[string][]define.InspectHostPort{
			"9090/tcp": nil,
		})
		if len(got) != 1 {
			t.Fatalf("len = %d, want 1", len(got))
		}
		if got[0].ContainerPort != 9090 || got[0].HostPort != 0 {
			t.Errorf("got %+v, want exposed 9090 with no host port", got[0])
		}
	})

	t.Run("published port parses host binding", func(t *testing.T) {
		got := portsFromInspect(map[string][]define.InspectHostPort{
			"80/tcp": {{HostIP: "0.0.0.0", HostPort: "8080"}},
		})
		if len(got) != 1 {
			t.Fatalf("len = %d, want 1", len(got))
		}
		if got[0].HostPort != 8080 || got[0].ContainerPort != 80 || got[0].HostIP != "0.0.0.0" {
			t.Errorf("got %+v, want 0.0.0.0:8080->80", got[0])
		}
	})
}

func TestImageSummaryToDomain(t *testing.T) {
	s := &entitiesTypes.ImageSummary{
		ID:          "img1",
		ParentId:    "parent",
		RepoTags:    []string{"nginx:1.25"},
		RepoDigests: []string{"nginx@sha256:abc"},
		Created:     1700000000,
		Size:        100,
		SharedSize:  50,
		VirtualSize: 120,
		Labels:      map[string]string{"k": "v"},
		Containers:  2,
	}
	got := imageSummaryToDomain(s)
	if got.ID != "img1" || got.ParentID != "parent" {
		t.Errorf("IDs wrong: %+v", got)
	}
	if got.Created.Unix() != 1700000000 {
		t.Errorf("Created = %v, want unix 1700000000", got.Created.Unix())
	}
	if got.SharedSize != 50 || got.Containers != 2 {
		t.Errorf("SharedSize/Containers = (%d, %d), want (50, 2)", got.SharedSize, got.Containers)
	}
	if got.RepoTags[0] != "nginx:1.25" {
		t.Errorf("RepoTags = %v", got.RepoTags)
	}
}

func TestVolumeReportToDomain(t *testing.T) {
	created := time.Unix(1700000000, 0)
	v := &entitiesTypes.VolumeListReport{
		VolumeConfigResponse: entitiesTypes.VolumeConfigResponse{
			InspectVolumeData: define.InspectVolumeData{
				Name:       "vol1",
				Driver:     "local",
				Mountpoint: "/var/lib/containers/storage/volumes/vol1/_data",
				Scope:      "local",
				CreatedAt:  created,
				Labels:     map[string]string{"k": "v"},
				Options:    map[string]string{"o": "p"},
				Status:     map[string]any{"foo": "bar"},
			},
		},
	}
	got := volumeReportToDomain(v)
	if got.Name != "vol1" || got.Scope != domain.VolumeScopeLocal {
		t.Errorf("basic fields wrong: %+v", got)
	}
	if !got.CreatedAt.Equal(created) {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, created)
	}
	if got.Status["foo"] != "bar" {
		t.Errorf("Status = %v", got.Status)
	}
	// Podman does not report usage data in the list.
	if got.UsageData != nil {
		t.Errorf("UsageData = %+v, want nil for podman list", got.UsageData)
	}
}

func TestHistoryToDomain(t *testing.T) {
	h := &handlersTypes.HistoryResponse{
		ID:        "layer1",
		Created:   1700000000,
		CreatedBy: "RUN apk add curl",
		Size:      512,
		Comment:   "buildkit",
		Tags:      []string{"app:1.0"},
	}
	got := historyToDomain(h)
	if got.ID != "layer1" || got.CreatedBy != "RUN apk add curl" {
		t.Errorf("basic fields wrong: %+v", got)
	}
	if got.Created.Unix() != 1700000000 {
		t.Errorf("Created = %v, want unix 1700000000", got.Created.Unix())
	}
	if got.Size != 512 || got.Tags[0] != "app:1.0" {
		t.Errorf("Size/Tags wrong: %+v", got)
	}
}

func TestInspectMountsToDomain(t *testing.T) {
	if got := inspectMountsToDomain(nil); got != nil {
		t.Errorf("nil mounts: got %+v, want nil", got)
	}

	got := inspectMountsToDomain([]define.InspectMount{
		{Type: "bind", Source: "/src", Destination: "/dst", Mode: "Z", RW: true},
		{Type: "volume", Name: "data", Source: "data", Destination: "/var/lib", RW: false},
	})
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].Type != domain.MountType("bind") || !got[0].ReadWrite || got[0].Mode != "Z" {
		t.Errorf("mount[0] = %+v", got[0])
	}
	if got[1].Type != domain.MountType("volume") || got[1].Name != "data" || got[1].ReadWrite {
		t.Errorf("mount[1] = %+v", got[1])
	}
}

func TestHealthToDomain(t *testing.T) {
	if got := healthToDomain(nil); got != nil {
		t.Errorf("nil health: got %+v, want nil", got)
	}

	h := &define.HealthCheckResults{
		Status:        "unhealthy",
		FailingStreak: 2,
		Log: []define.HealthCheckLog{
			{
				Start:    "2026-06-20T10:00:00.000000000Z",
				End:      "2026-06-20T10:00:01.000000000Z",
				ExitCode: 1,
				Output:   "probe failed",
			},
		},
	}
	got := healthToDomain(h)
	if got == nil || got.Status != domain.HealthStatusUnhealthy || got.FailingStreak != 2 {
		t.Fatalf("health header not mapped: %+v", got)
	}
	if len(got.Log) != 1 {
		t.Fatalf("len(Log) = %d, want 1", len(got.Log))
	}
	if got.Log[0].ExitCode != 1 || got.Log[0].Output != "probe failed" {
		t.Errorf("log entry = %+v", got.Log[0])
	}
	if !got.Log[0].Start.Equal(time.Date(2026, 6, 20, 10, 0, 0, 0, time.UTC)) {
		t.Errorf("Start = %v, want parsed RFC3339Nano", got.Log[0].Start)
	}
}

func TestNetworkToDomain(t *testing.T) {
	created := time.Unix(1700000000, 0)
	sub, err := netTypes.ParseCIDR("10.89.0.0/24")
	if err != nil {
		t.Fatalf("ParseCIDR: %v", err)
	}
	n := netTypes.Network{
		ID:      "net1",
		Name:    "mynet",
		Driver:  "bridge",
		Created: created,
		Subnets: []netTypes.Subnet{
			{Subnet: sub, Gateway: net.ParseIP("10.89.0.1")},
		},
		IPAMOptions: map[string]string{"driver": "host-local"},
		Internal:    true,
		IPv6Enabled: true,
		Labels:      map[string]string{"k": "v"},
		Options:     map[string]string{"o": "p"},
	}
	got := networkToDomain(n)
	if got.ID != "net1" || got.Name != "mynet" {
		t.Errorf("basic fields wrong: %+v", got)
	}
	// Podman networks are always local-scoped.
	if got.Scope != domain.NetworkScopeLocal {
		t.Errorf("Scope = %q, want local", got.Scope)
	}
	if !got.Internal || !got.EnableIPv6 {
		t.Errorf("Internal/EnableIPv6 = (%v, %v), want (true, true)", got.Internal, got.EnableIPv6)
	}
	if got.IPAM.Driver != "host-local" {
		t.Errorf("IPAM.Driver = %q, want host-local", got.IPAM.Driver)
	}
	if len(got.IPAM.Config) != 1 {
		t.Fatalf("len(IPAM.Config) = %d, want 1", len(got.IPAM.Config))
	}
	if got.IPAM.Config[0].Subnet != "10.89.0.0/24" || got.IPAM.Config[0].Gateway != "10.89.0.1" {
		t.Errorf("IPAM.Config[0] = %+v, want 10.89.0.0/24 gw 10.89.0.1", got.IPAM.Config[0])
	}
}

func TestInspectContainerToDomain(t *testing.T) {
	created := time.Unix(1700000000, 0)
	started := time.Date(2026, 6, 20, 10, 1, 0, 0, time.UTC)
	finished := time.Date(2026, 6, 20, 10, 5, 0, 0, time.UTC)
	sizeRw := int64(1024)

	d := &define.InspectContainerData{
		ID:           "id",
		Name:         "test",
		Image:        "sha256:img",
		Created:      created,
		Path:         "/bin/sh",
		Args:         []string{"-c", "echo hi"},
		RestartCount: 3,
		SizeRw:       &sizeRw,
		SizeRootFs:   2048,
		Config: &define.InspectContainerConfig{
			Image:      "alpine",
			Cmd:        []string{"sh"},
			Entrypoint: []string{"/entry"},
			Env:        []string{"FOO=bar"},
			Labels:     map[string]string{"app": "web"},
			WorkingDir: "/work",
			User:       "root",
			Tty:        true,
			OpenStdin:  true,
		},
		State: &define.InspectContainerState{
			Status:     "exited",
			ExitCode:   0,
			StartedAt:  started,
			FinishedAt: finished,
			Health: &define.HealthCheckResults{
				Status:        "unhealthy",
				FailingStreak: 1,
			},
		},
		Mounts: []define.InspectMount{
			{Type: "bind", Source: "/src", Destination: "/dst", RW: true},
		},
		NetworkSettings: &define.InspectNetworkSettings{
			Networks: map[string]*define.InspectAdditionalNetwork{
				"podman": {
					NetworkID: "net-id",
					InspectBasicNetworkConfig: define.InspectBasicNetworkConfig{
						IPAddress:  "10.88.0.2",
						MacAddress: "aa:bb",
					},
				},
				// A nil endpoint must be skipped, not panic.
				"broken": nil,
			},
			Ports: map[string][]define.InspectHostPort{
				"80/tcp": {{HostIP: "0.0.0.0", HostPort: "8080"}},
			},
		},
	}

	got := inspectContainerToDomain(d)

	if got.ID != "id" || got.RestartCount != 3 {
		t.Errorf("ID/RestartCount wrong: %+v", got)
	}
	if got.Names[0] != "test" {
		t.Errorf("Names = %v, want [test]", got.Names)
	}
	if got.SizeRw != 1024 || got.SizeRootFs != 2048 {
		t.Errorf("sizes = (%d, %d), want (1024, 2048)", got.SizeRw, got.SizeRootFs)
	}
	if got.State != domain.ContainerStateExited {
		t.Errorf("State = %q, want exited", got.State)
	}
	// Config-derived summary fields.
	if got.Image != "alpine" || got.Command != "sh" || got.Labels["app"] != "web" {
		t.Errorf("summary from Config wrong: image=%q cmd=%q labels=%v", got.Image, got.Command, got.Labels)
	}
	if got.Config.WorkingDir != "/work" || got.Config.User != "root" || !got.Config.Tty || !got.Config.OpenStdin {
		t.Errorf("Config not mapped: %+v", got.Config)
	}
	if len(got.Config.Entrypoint) != 1 || got.Config.Entrypoint[0] != "/entry" {
		t.Errorf("Entrypoint = %v", got.Config.Entrypoint)
	}
	if got.Health == nil || got.Health.Status != domain.HealthStatusUnhealthy {
		t.Errorf("Health not mapped: %+v", got.Health)
	}
	if !got.StartedAt.Equal(started) || !got.FinishedAt.Equal(finished) {
		t.Errorf("times = (%v, %v), want (%v, %v)", got.StartedAt, got.FinishedAt, started, finished)
	}
	if len(got.Mounts) != 1 || got.Mounts[0].Type != domain.MountType("bind") {
		t.Errorf("Mounts = %+v", got.Mounts)
	}
	if ep, ok := got.NetworkSettings.Endpoints["podman"]; !ok || ep.IPAddress != "10.88.0.2" {
		t.Errorf("endpoint not mapped: %+v", got.NetworkSettings.Endpoints)
	}
	// Ports come from the inspect NetworkSettings.Ports map.
	if len(got.Ports) != 1 || got.Ports[0].HostPort != 8080 || got.Ports[0].ContainerPort != 80 {
		t.Errorf("Ports = %+v, want 8080->80", got.Ports)
	}
}
