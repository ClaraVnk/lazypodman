package presentation

import "github.com/ClaraVnk/lazypodman/pkg/commands"

func GetPodDisplayStrings(pod *commands.Pod) []string {
	return []string{string(pod.Pod.Status), pod.Name}
}
