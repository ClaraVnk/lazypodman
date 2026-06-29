package commands

import (
	"io"

	"github.com/ClaraVnk/lazypodman/pkg/config"
	"github.com/ClaraVnk/lazypodman/pkg/i18n"
	"github.com/sirupsen/logrus"
)

// This file exports dummy constructors for use by tests in other packages

// NewDummyOSCommand creates a new dummy OSCommand for testing
func NewDummyOSCommand() *OSCommand {
	return NewOSCommand(NewDummyLog(), NewDummyAppConfig())
}

// NewDummyAppConfig creates a new dummy AppConfig for testing
func NewDummyAppConfig() *config.AppConfig {
	appConfig := &config.AppConfig{
		Name:        "lazypodman",
		Version:     "unversioned",
		Commit:      "",
		BuildDate:   "",
		Debug:       false,
		BuildSource: "",
	}
	return appConfig
}

// NewDummyLog creates a new dummy Log for testing
func NewDummyLog() *logrus.Entry {
	log := logrus.New()
	log.Out = io.Discard
	return log.WithField("test", "test")
}

// NewDummyContainerCommand creates a new dummy ContainerCommand for testing
func NewDummyContainerCommand() *ContainerCommand {
	return NewDummyContainerCommandWithOSCommand(NewDummyOSCommand())
}

// NewDummyContainerCommandWithOSCommand creates a new dummy ContainerCommand for testing
func NewDummyContainerCommandWithOSCommand(osCommand *OSCommand) *ContainerCommand {
	newAppConfig := NewDummyAppConfig()
	return &ContainerCommand{
		Log:       NewDummyLog(),
		OSCommand: osCommand,
		Tr:        i18n.NewTranslationSet(NewDummyLog(), newAppConfig.UserConfig.Gui.Language),
		Config:    newAppConfig,
	}
}
