package presentation

import "github.com/ClaraVnk/lazypodman/pkg/commands"

func GetNetworkDisplayStrings(network *commands.Network) []string {
	return []string{network.Network.Driver, network.Name}
}
