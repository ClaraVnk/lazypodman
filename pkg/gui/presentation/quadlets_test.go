package presentation

import (
	"testing"

	"github.com/ClaraVnk/lazypodman/pkg/commands"
	"github.com/ClaraVnk/lazypodman/pkg/domain"
)

func TestGetQuadletDisplayStrings(t *testing.T) {
	quadlet := &commands.Quadlet{
		Name:    "web",
		Quadlet: domain.Quadlet{Type: domain.QuadletContainer, ActiveState: "active"},
	}
	got := GetQuadletDisplayStrings(quadlet)
	want := []string{"active", "container", "web"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Errorf("GetQuadletDisplayStrings = %v, want %v", got, want)
	}
}
