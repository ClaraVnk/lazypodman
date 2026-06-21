package gui

import (
	"github.com/fatih/color"
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazydocker/pkg/commands"
	"github.com/jesseduffield/lazydocker/pkg/config"
	"github.com/jesseduffield/lazydocker/pkg/gui/panels"
	"github.com/jesseduffield/lazydocker/pkg/gui/presentation"
	"github.com/jesseduffield/lazydocker/pkg/gui/types"
	"github.com/jesseduffield/lazydocker/pkg/tasks"
	"github.com/jesseduffield/lazydocker/pkg/utils"
	"github.com/samber/lo"
)

func (gui *Gui) getPodsPanel() *panels.SideListPanel[*commands.Pod] {
	return &panels.SideListPanel[*commands.Pod]{
		ContextState: &panels.ContextState[*commands.Pod]{
			GetMainTabs: func() []panels.MainTab[*commands.Pod] {
				return []panels.MainTab[*commands.Pod]{
					{
						Key:    "config",
						Title:  gui.Tr.ConfigTitle,
						Render: gui.renderPodConfig,
					},
				}
			},
			GetItemContextCacheKey: func(pod *commands.Pod) string {
				return "pods-" + pod.Name
			},
		},
		ListPanel: panels.ListPanel[*commands.Pod]{
			List: panels.NewFilteredList[*commands.Pod](),
			View: gui.Views.Pods,
		},
		NoItemsMessage: gui.Tr.NoPods,
		Gui:            gui.intoInterface(),
		Sort: func(a *commands.Pod, b *commands.Pod) bool {
			return a.Name < b.Name
		},
		GetTableCells: presentation.GetPodDisplayStrings,
		// Pods only exist on a Podman backend; hide the panel otherwise.
		Hide: func() bool {
			return !gui.DockerCommand.PodsSupported()
		},
	}
}

func (gui *Gui) renderPodConfig(pod *commands.Pod) tasks.TaskFunc {
	return gui.NewSimpleRenderStringTask(func() string { return gui.podConfigStr(pod) })
}

func (gui *Gui) podConfigStr(pod *commands.Pod) string {
	padding := 15
	output := ""
	output += utils.WithPadding("ID: ", padding) + pod.Pod.ID + "\n"
	output += utils.WithPadding("Name: ", padding) + pod.Name + "\n"
	output += utils.WithPadding("Status: ", padding) + string(pod.Pod.Status) + "\n"
	output += utils.WithPadding("Created: ", padding) + pod.Pod.Created.Format("2006-01-02 15:04:05") + "\n"
	if pod.Pod.Namespace != "" {
		output += utils.WithPadding("Namespace: ", padding) + pod.Pod.Namespace + "\n"
	}

	output += utils.WithPadding("Containers: ", padding)
	if len(pod.Pod.Containers) > 0 {
		output += "\n"
		for _, c := range pod.Pod.Containers {
			output += utils.FormatMapItem(padding, c.Name, c.Status)
		}
	} else {
		output += "none\n"
	}

	output += "\n"
	output += utils.WithPadding("Networks: ", padding)
	if len(pod.Pod.Networks) > 0 {
		output += utils.FormatMap(padding, lo.SliceToMap(pod.Pod.Networks, func(n string) (string, string) {
			return n, ""
		}))
	} else {
		output += "none"
	}
	output += "\n"
	output += utils.WithPadding("Labels: ", padding) + utils.FormatMap(padding, pod.Pod.Labels)

	return output
}

func (gui *Gui) reloadPods() error {
	if err := gui.refreshStatePods(); err != nil {
		return err
	}

	return gui.Panels.Pods.RerenderList()
}

func (gui *Gui) refreshStatePods() error {
	pods, err := gui.DockerCommand.RefreshPods()
	if err != nil {
		return err
	}

	gui.Panels.Pods.SetItems(pods)

	return nil
}

func (gui *Gui) handlePodsRemoveMenu(g *gocui.Gui, v *gocui.View) error {
	pod, err := gui.Panels.Pods.GetSelectedItem()
	if err != nil {
		return nil
	}

	type removePodOption struct {
		description string
		command     string
		force       bool
	}

	options := []*removePodOption{
		{
			description: gui.Tr.Remove,
			command:     utils.WithShortSha("podman pod rm " + pod.Name),
			force:       false,
		},
		{
			description: gui.Tr.ForceRemove,
			command:     utils.WithShortSha("podman pod rm --force " + pod.Name),
			force:       true,
		},
	}

	menuItems := lo.Map(options, func(option *removePodOption, _ int) *types.MenuItem {
		return &types.MenuItem{
			LabelColumns: []string{option.description, color.New(color.FgRed).Sprint(option.command)},
			OnPress: func() error {
				return gui.WithWaitingStatus(gui.Tr.RemovingStatus, func() error {
					if err := pod.Remove(option.force); err != nil {
						return gui.createErrorPanel(err.Error())
					}
					return nil
				})
			},
		}
	})

	return gui.Menu(CreateMenuOptions{
		Title: "",
		Items: menuItems,
	})
}

func (gui *Gui) handlePrunePods() error {
	return gui.createConfirmationPanel(gui.Tr.Confirm, gui.Tr.ConfirmPrunePods, func(g *gocui.Gui, v *gocui.View) error {
		return gui.WithWaitingStatus(gui.Tr.PruningStatus, func() error {
			err := gui.DockerCommand.PrunePods()
			if err != nil {
				return gui.createErrorPanel(err.Error())
			}
			return nil
		})
	}, nil)
}

func (gui *Gui) handlePodsBulkCommand(g *gocui.Gui, v *gocui.View) error {
	baseBulkCommands := []config.CustomCommand{
		{
			Name:             gui.Tr.PrunePods,
			InternalFunction: gui.handlePrunePods,
		},
	}

	commandObject := gui.DockerCommand.NewCommandObject(commands.CommandObject{})

	return gui.createBulkCommandMenu(baseBulkCommands, commandObject)
}
