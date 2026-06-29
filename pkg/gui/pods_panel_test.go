package gui

import (
	"strings"
	"testing"
	"time"

	"github.com/ClaraVnk/lazypodman/pkg/commands"
	"github.com/ClaraVnk/lazypodman/pkg/domain"
)

// TestPodConfigStr exercises the pod detail rendering (the new, untested-
// by-the-TUI code path) and asserts the key fields appear without panic,
// including the empty-collections branches.
func TestPodConfigStr(t *testing.T) {
	gui := &Gui{}

	full := &commands.Pod{
		Name: "web-pod",
		Pod: domain.PodInfo{
			ID:        "abc123",
			Name:      "web-pod",
			Status:    domain.PodStatusRunning,
			Created:   time.Date(2026, 6, 21, 10, 0, 0, 0, time.UTC),
			Namespace: "ns1",
			Networks:  []string{"podman"},
			Labels:    map[string]string{"app": "web"},
			Containers: []domain.PodContainer{
				{ID: "c1", Name: "infra", Status: "running"},
			},
		},
	}
	out := gui.podConfigStr(full)
	for _, want := range []string{"abc123", "web-pod", "running", "infra", "podman", "app"} {
		if !strings.Contains(out, want) {
			t.Errorf("podConfigStr missing %q in:\n%s", want, out)
		}
	}

	// Empty collections must render the "none" branches without panicking.
	empty := &commands.Pod{Name: "bare", Pod: domain.PodInfo{ID: "x", Status: domain.PodStatusCreated}}
	emptyOut := gui.podConfigStr(empty)
	if !strings.Contains(emptyOut, "none") {
		t.Errorf("podConfigStr(empty) should render 'none' for empty containers/networks:\n%s", emptyOut)
	}
}
