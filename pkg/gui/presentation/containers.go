package presentation

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/ClaraVnk/lazypodman/pkg/commands"
	"github.com/ClaraVnk/lazypodman/pkg/config"
	"github.com/ClaraVnk/lazypodman/pkg/domain"
	"github.com/ClaraVnk/lazypodman/pkg/utils"
	"github.com/fatih/color"
	"github.com/samber/lo"
)

func GetContainerDisplayStrings(guiConfig *config.GuiConfig, container *commands.Container) []string {
	return []string{
		getContainerDisplayStatus(guiConfig, container),
		getContainerDisplaySubstatus(guiConfig, container),
		container.Name,
		getDisplayCPUPerc(container),
		utils.ColoredString(displayPorts(container), color.FgYellow),
		utils.ColoredString(displayContainerImage(container), color.FgMagenta),
	}
}

func displayContainerImage(container *commands.Container) string {
	return strings.TrimPrefix(container.Container.Image, "sha256:")
}

func displayPorts(c *commands.Container) string {
	portStrings := lo.Map(c.Container.Ports, func(port domain.Port, _ int) string {
		if port.HostPort == 0 {
			return fmt.Sprintf("%d/%s", port.ContainerPort, port.Protocol)
		}

		// docker ps will show '0.0.0.0:80->80/tcp' but we'll show
		// '80->80/tcp' instead to save space (unless the IP is something other
		// than 0.0.0.0)
		ipString := ""
		if port.HostIP != "0.0.0.0" && port.HostIP != "" {
			ipString = port.HostIP + ":"
		}
		return fmt.Sprintf("%s%d->%d/%s", ipString, port.HostPort, port.ContainerPort, port.Protocol)
	})

	// sorting because the order of the ports is not deterministic
	// and we don't want to have them constantly swapping
	sort.Strings(portStrings)

	return strings.Join(portStrings, ", ")
}

// getContainerDisplayStatus returns the colored status of the container
func getContainerDisplayStatus(guiConfig *config.GuiConfig, c *commands.Container) string {
	shortStatusMap := map[domain.ContainerState]string{
		domain.ContainerStatePaused:     "P",
		domain.ContainerStateExited:     "X",
		domain.ContainerStateCreated:    "C",
		domain.ContainerStateRemoving:   "RM",
		domain.ContainerStateRestarting: "RS",
		domain.ContainerStateRunning:    "R",
		domain.ContainerStateDead:       "D",
	}

	iconStatusMap := map[domain.ContainerState]rune{
		domain.ContainerStatePaused:     '◫',
		domain.ContainerStateExited:     '⨯',
		domain.ContainerStateCreated:    '+',
		domain.ContainerStateRemoving:   '−',
		domain.ContainerStateRestarting: '⟳',
		domain.ContainerStateRunning:    '▶',
		domain.ContainerStateDead:       '!',
	}

	var containerState string
	switch guiConfig.ContainerStatusHealthStyle {
	case "short":
		containerState = shortStatusMap[c.Container.State]
	case "icon":
		containerState = string(iconStatusMap[c.Container.State])
	case "long":
		fallthrough
	default:
		containerState = string(c.Container.State)
	}

	return utils.ColoredString(containerState, getContainerColor(c))
}

// GetDisplayStatus returns the exit code if the container has exited, and the health status if the container is running (and has a health check)
func getContainerDisplaySubstatus(guiConfig *config.GuiConfig, c *commands.Container) string {
	if !c.DetailsLoaded() {
		return ""
	}

	switch c.Container.State {
	case domain.ContainerStateExited:
		return utils.ColoredString(
			fmt.Sprintf("(%s)", strconv.Itoa(c.Details.ExitCode)), getContainerColor(c),
		)
	case domain.ContainerStateRunning:
		return getHealthStatus(guiConfig, c)
	default:
		return ""
	}
}

func getHealthStatus(guiConfig *config.GuiConfig, c *commands.Container) string {
	if !c.DetailsLoaded() {
		return ""
	}

	healthStatusColorMap := map[domain.HealthStatus]color.Attribute{
		domain.HealthStatusHealthy:   color.FgGreen,
		domain.HealthStatusUnhealthy: color.FgRed,
		domain.HealthStatusStarting:  color.FgYellow,
	}

	if c.Details.Health == nil {
		return ""
	}

	shortHealthStatusMap := map[domain.HealthStatus]string{
		domain.HealthStatusHealthy:   "H",
		domain.HealthStatusUnhealthy: "U",
		domain.HealthStatusStarting:  "S",
	}

	iconHealthStatusMap := map[domain.HealthStatus]rune{
		domain.HealthStatusHealthy:   '✔',
		domain.HealthStatusUnhealthy: '?',
		domain.HealthStatusStarting:  '…',
	}

	var healthStatus string
	switch guiConfig.ContainerStatusHealthStyle {
	case "short":
		healthStatus = shortHealthStatusMap[c.Details.Health.Status]
	case "icon":
		healthStatus = string(iconHealthStatusMap[c.Details.Health.Status])
	case "long":
		fallthrough
	default:
		healthStatus = string(c.Details.Health.Status)
	}

	if healthStatusColor, ok := healthStatusColorMap[c.Details.Health.Status]; ok {
		return utils.ColoredString(fmt.Sprintf("(%s)", healthStatus), healthStatusColor)
	}
	return ""
}

// getDisplayCPUPerc colors the cpu percentage based on how extreme it is
func getDisplayCPUPerc(c *commands.Container) string {
	stats, ok := c.GetLastStats()
	if !ok {
		return ""
	}

	percentage := stats.DerivedStats.CPUPercentage
	formattedPercentage := fmt.Sprintf("%.2f%%", stats.DerivedStats.CPUPercentage)

	var clr color.Attribute
	if percentage > 90 {
		clr = color.FgRed
	} else if percentage > 50 {
		clr = color.FgYellow
	} else {
		clr = color.FgWhite
	}

	return utils.ColoredString(formattedPercentage, clr)
}

// getContainerColor Container color
func getContainerColor(c *commands.Container) color.Attribute {
	switch c.Container.State {
	case domain.ContainerStateExited:
		// This means the colour may be briefly yellow and then switch to red upon starting
		// Not sure what a better alternative is.
		if !c.DetailsLoaded() || c.Details.ExitCode == 0 {
			return color.FgYellow
		}
		return color.FgRed
	case domain.ContainerStateCreated:
		return color.FgCyan
	case domain.ContainerStateRunning:
		return color.FgGreen
	case domain.ContainerStatePaused:
		return color.FgYellow
	case domain.ContainerStateDead:
		return color.FgRed
	case domain.ContainerStateRestarting:
		return color.FgBlue
	case domain.ContainerStateRemoving:
		return color.FgMagenta
	default:
		return color.FgWhite
	}
}
