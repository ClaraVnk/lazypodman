package gui

import (
	"os"

	"github.com/ClaraVnk/lazypodman/pkg/commands"
	"github.com/ClaraVnk/lazypodman/pkg/gui/panels"
	"github.com/ClaraVnk/lazypodman/pkg/gui/presentation"
	"github.com/ClaraVnk/lazypodman/pkg/tasks"
	"github.com/ClaraVnk/lazypodman/pkg/utils"
	"github.com/jesseduffield/gocui"
)

func (gui *Gui) getQuadletsPanel() *panels.SideListPanel[*commands.Quadlet] {
	return &panels.SideListPanel[*commands.Quadlet]{
		ContextState: &panels.ContextState[*commands.Quadlet]{
			GetMainTabs: func() []panels.MainTab[*commands.Quadlet] {
				return []panels.MainTab[*commands.Quadlet]{
					{
						Key:    "source",
						Title:  gui.Tr.QuadletSourceTitle,
						Render: gui.renderQuadletSource,
					},
				}
			},
			GetItemContextCacheKey: func(quadlet *commands.Quadlet) string {
				return "quadlets-" + quadlet.Name
			},
		},
		ListPanel: panels.ListPanel[*commands.Quadlet]{
			List: panels.NewFilteredList[*commands.Quadlet](),
			View: gui.Views.Quadlets,
		},
		NoItemsMessage: gui.Tr.NoQuadlets,
		Gui:            gui.intoInterface(),
		Sort: func(a *commands.Quadlet, b *commands.Quadlet) bool {
			return a.Name < b.Name
		},
		GetTableCells: presentation.GetQuadletDisplayStrings,
		// Quadlets are managed via systemd and only exist on a Podman
		// backend; hide the panel otherwise.
		Hide: func() bool {
			return !gui.ContainerCommand.QuadletsSupported()
		},
	}
}

// renderQuadletSource shows the quadlet's source unit file in the main panel.
func (gui *Gui) renderQuadletSource(quadlet *commands.Quadlet) tasks.TaskFunc {
	return gui.NewSimpleRenderStringTask(func() string {
		return gui.quadletSourceStr(quadlet)
	})
}

func (gui *Gui) quadletSourceStr(quadlet *commands.Quadlet) string {
	padding := 15
	output := ""
	output += utils.WithPadding("Name: ", padding) + quadlet.Name + "\n"
	output += utils.WithPadding("Unit: ", padding) + quadlet.Quadlet.UnitName + "\n"
	output += utils.WithPadding("Type: ", padding) + string(quadlet.Quadlet.Type) + "\n"
	output += utils.WithPadding("State: ", padding) + quadlet.Quadlet.ActiveState + "\n"
	output += utils.WithPadding("Source: ", padding) + quadlet.Quadlet.SourcePath + "\n\n"

	contents, err := os.ReadFile(quadlet.Quadlet.SourcePath)
	if err != nil {
		return output + err.Error()
	}
	return output + string(contents)
}

func (gui *Gui) reloadQuadlets() error {
	if err := gui.refreshStateQuadlets(); err != nil {
		return err
	}

	return gui.Panels.Quadlets.RerenderList()
}

func (gui *Gui) refreshStateQuadlets() error {
	quadlets, err := gui.ContainerCommand.RefreshQuadlets()
	if err != nil {
		return err
	}

	gui.Panels.Quadlets.SetItems(quadlets)

	return nil
}

func (gui *Gui) handleQuadletStart(g *gocui.Gui, v *gocui.View) error {
	quadlet, err := gui.Panels.Quadlets.GetSelectedItem()
	if err != nil {
		return nil
	}

	return gui.WithWaitingStatus(gui.Tr.StartingStatus, func() error {
		if err := quadlet.Start(); err != nil {
			return gui.createErrorPanel(err.Error())
		}
		return gui.reloadQuadlets()
	})
}

func (gui *Gui) handleQuadletStop(g *gocui.Gui, v *gocui.View) error {
	quadlet, err := gui.Panels.Quadlets.GetSelectedItem()
	if err != nil {
		return nil
	}

	return gui.WithWaitingStatus(gui.Tr.StoppingStatus, func() error {
		if err := quadlet.Stop(); err != nil {
			return gui.createErrorPanel(err.Error())
		}
		return gui.reloadQuadlets()
	})
}

func (gui *Gui) handleQuadletRestart(g *gocui.Gui, v *gocui.View) error {
	quadlet, err := gui.Panels.Quadlets.GetSelectedItem()
	if err != nil {
		return nil
	}

	return gui.WithWaitingStatus(gui.Tr.RestartingStatus, func() error {
		if err := quadlet.Restart(); err != nil {
			return gui.createErrorPanel(err.Error())
		}
		return gui.reloadQuadlets()
	})
}
