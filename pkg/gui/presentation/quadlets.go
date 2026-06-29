package presentation

import "github.com/ClaraVnk/lazypodman/pkg/commands"

func GetQuadletDisplayStrings(quadlet *commands.Quadlet) []string {
	return []string{quadlet.Quadlet.ActiveState, string(quadlet.Quadlet.Type), quadlet.Name}
}
