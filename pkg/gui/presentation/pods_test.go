package presentation

import (
	"testing"

	"github.com/ClaraVnk/lazypodman/pkg/commands"
	"github.com/ClaraVnk/lazypodman/pkg/domain"
)

func TestGetPodDisplayStrings(t *testing.T) {
	pod := &commands.Pod{
		Name: "web",
		Pod:  domain.PodInfo{Status: domain.PodStatusRunning},
	}
	got := GetPodDisplayStrings(pod)
	want := []string{"running", "web"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("GetPodDisplayStrings = %v, want %v", got, want)
	}
}
