package commands

import (
	"testing"

	dockerruntime "github.com/jesseduffield/lazydocker/pkg/runtime/docker"
	podmanruntime "github.com/jesseduffield/lazydocker/pkg/runtime/podman"
)

// TestPodsCapabilityGating checks that pods are exposed only when the
// active runtime implements runtime.PodRuntime: the Docker backend must
// report no support and RefreshPods must short-circuit to nil without
// touching a socket, while the Podman backend advertises the capability.
// Both runtimes connect lazily, so neither construction needs a daemon.
func TestPodsCapabilityGating(t *testing.T) {
	dockerRT, err := dockerruntime.NewFromEnv()
	if err != nil {
		t.Fatalf("docker NewFromEnv: %v", err)
	}
	podmanRT, err := podmanruntime.NewFromEnv()
	if err != nil {
		t.Fatalf("podman NewFromEnv: %v", err)
	}

	docker := &DockerCommand{Runtime: dockerRT}
	if docker.PodsSupported() {
		t.Error("Docker backend should not support pods")
	}
	pods, err := docker.RefreshPods()
	if err != nil {
		t.Errorf("Docker RefreshPods err = %v, want nil", err)
	}
	if pods != nil {
		t.Errorf("Docker RefreshPods = %v, want nil (no pod support)", pods)
	}
	if err := docker.PrunePods(); err != nil {
		t.Errorf("Docker PrunePods err = %v, want nil (no-op)", err)
	}

	podman := &DockerCommand{Runtime: podmanRT}
	if !podman.PodsSupported() {
		t.Error("Podman backend should support pods")
	}

	// GenerateKube is gated by the KubeGenerator capability: Docker errors
	// instead of producing output.
	if _, err := docker.GenerateKube([]string{"x"}); err == nil {
		t.Error("Docker GenerateKube should error (capability unsupported)")
	}
}
