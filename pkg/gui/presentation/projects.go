package presentation

import "github.com/ClaraVnk/lazypodman/pkg/commands"

func GetProjectDisplayStrings(project *commands.Project) []string {
	return []string{project.Name}
}
