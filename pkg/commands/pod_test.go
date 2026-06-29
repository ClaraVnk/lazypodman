package commands

import (
	"testing"

	podmanruntime "github.com/ClaraVnk/lazypodman/pkg/runtime/podman"
)

// TestPodsCapabilityGating checks that pods are exposed only when the active
// runtime implements runtime.PodRuntime. The Podman backend advertises the
// capability; a runtime without it (nil stands in) must report no support and
// short-circuit RefreshPods/PrunePods to nil without touching a socket.
// Construction is lazy, so neither path needs a daemon.
func TestPodsCapabilityGating(t *testing.T) {
	podmanRT, err := podmanruntime.NewFromEnv()
	if err != nil {
		t.Fatalf("podman NewFromEnv: %v", err)
	}
	podman := &ContainerCommand{Runtime: podmanRT}
	if !podman.PodsSupported() {
		t.Error("Podman backend should support pods")
	}

	// A runtime that implements neither PodRuntime nor KubeGenerator must
	// degrade gracefully rather than panic.
	noPods := &ContainerCommand{Runtime: nil}
	if noPods.PodsSupported() {
		t.Error("a runtime without PodRuntime should not support pods")
	}
	pods, err := noPods.RefreshPods()
	if err != nil {
		t.Errorf("RefreshPods err = %v, want nil", err)
	}
	if pods != nil {
		t.Errorf("RefreshPods = %v, want nil (no pod support)", pods)
	}
	if err := noPods.PrunePods(); err != nil {
		t.Errorf("PrunePods err = %v, want nil (no-op)", err)
	}

	// GenerateKube is gated by the KubeGenerator capability.
	if _, err := noPods.GenerateKube([]string{"x"}); err == nil {
		t.Error("GenerateKube should error when KubeGenerator is unsupported")
	}
}
