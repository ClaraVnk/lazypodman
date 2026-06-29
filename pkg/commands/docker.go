package commands

import (
	"context"
	"fmt"
	"io"
	ogLog "log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ClaraVnk/lazypodman/pkg/config"
	"github.com/ClaraVnk/lazypodman/pkg/i18n"
	"github.com/ClaraVnk/lazypodman/pkg/runtime"
	podmanruntime "github.com/ClaraVnk/lazypodman/pkg/runtime/podman"
	"github.com/ClaraVnk/lazypodman/pkg/utils"
	"github.com/imdario/mergo"
	"github.com/sasha-s/go-deadlock"
	"github.com/sirupsen/logrus"
)

// runtimeEnvKey overrides the configured backend ("docker" | "podman").
const runtimeEnvKey = "LAZYPODMAN_RUNTIME"

// selectBackend resolves which container runtime to use: the
// LAZYPODMAN_RUNTIME env var wins, then the config `runtime:` field, then
// the "podman" default. Docker stays available as an explicit fallback
// (runtime: docker). See docs/adr/0005-podman-native-backend.md.
func selectBackend(config *config.AppConfig) string {
	if v := strings.TrimSpace(os.Getenv(runtimeEnvKey)); v != "" {
		return strings.ToLower(v)
	}
	if config.UserConfig != nil {
		if v := strings.TrimSpace(config.UserConfig.Runtime); v != "" {
			return strings.ToLower(v)
		}
	}
	return "podman"
}

// EngineName returns the human-facing name of the active backend
// ("Podman", "Docker") for user-facing messages, falling back to a neutral
// label when the backend is unknown.
func (c *ContainerCommand) EngineName() string {
	switch c.Backend {
	case "podman":
		return "Podman"
	case "docker":
		return "Docker"
	default:
		return "container engine"
	}
}

// ContainerCommand is our main docker interface
type ContainerCommand struct {
	Log       *logrus.Entry
	OSCommand *OSCommand
	Tr        *i18n.TranslationSet
	Config    *config.AppConfig
	// Backend is the resolved container backend ("podman" | "docker"), used
	// for backend-aware user-facing messages. See EngineName.
	Backend string
	// Runtime is the sole abstraction lazypodman uses to talk to the
	// container engine. See docs/adr/0004-phase-1d-staged-rewire-strategy.md.
	Runtime                runtime.ContainerRuntime
	InDockerComposeProject bool
	// LocalProjectName is the compose project name for the directory where lazypodman was launched.
	LocalProjectName string
	ErrorChan        chan error
	ContainerMutex   deadlock.Mutex
	ServiceMutex     deadlock.Mutex

	Closers []io.Closer
}

var _ io.Closer = &ContainerCommand{}

// LimitedContainerCommand is a stripped-down ContainerCommand with just the methods the container/service/image might need
type LimitedContainerCommand interface {
	NewCommandObject(CommandObject) CommandObject
}

// CommandObject is what we pass to our template resolvers when we are running a custom command. We do not guarantee that all fields will be populated: just the ones that make sense for the current context
type CommandObject struct {
	DockerCompose string
	Service       *Service
	Container     *Container
	Image         *Image
	Volume        *Volume
	Network       *Network
	Project       *Project
}

// NewCommandObject takes a command object and returns a default command object with the passed command object merged in
func (c *ContainerCommand) NewCommandObject(obj CommandObject) CommandObject {
	defaultObj := CommandObject{DockerCompose: c.Config.UserConfig.CommandTemplates.DockerCompose}
	_ = mergo.Merge(&defaultObj, obj)

	// When operating on a specific project, include -p flag so that
	// docker compose targets the correct project.
	if obj.Service != nil && obj.Service.ProjectName != "" {
		defaultObj.DockerCompose = fmt.Sprintf("%s -p %s", defaultObj.DockerCompose, obj.Service.ProjectName)
	} else if obj.Project != nil && obj.Project.Name != "" {
		defaultObj.DockerCompose = fmt.Sprintf("%s -p %s", defaultObj.DockerCompose, obj.Project.Name)
	}

	return defaultObj
}

// NewContainerCommand it runs docker commands
func NewContainerCommand(log *logrus.Entry, osCommand *OSCommand, tr *i18n.TranslationSet, config *config.AppConfig, errorChan chan error) (*ContainerCommand, error) {
	backend := selectBackend(config)
	rt, closers, err := buildRuntime(backend, osCommand)
	if err != nil {
		return nil, err
	}

	dockerCommand := &ContainerCommand{
		Log:                    log,
		OSCommand:              osCommand,
		Tr:                     tr,
		Config:                 config,
		Backend:                backend,
		Runtime:                rt,
		ErrorChan:              errorChan,
		InDockerComposeProject: true,
		Closers:                closers,
	}

	dockerCommand.setDockerComposeCommand(config)

	err = osCommand.RunCommand(
		utils.ApplyTemplate(
			config.UserConfig.CommandTemplates.CheckDockerComposeConfig,
			dockerCommand.NewCommandObject(CommandObject{}),
		),
	)
	if err != nil {
		dockerCommand.InDockerComposeProject = false
		log.Warn(err.Error())
	}

	// When the user passes -p outside of a compose directory, treat it as the
	// local project so the project/services panels still appear and filtering
	// is applied. Inside a compose dir, LocalProjectName is derived from
	// container labels later in RefreshContainersAndServices.
	if !dockerCommand.InDockerComposeProject && config.ProjectName != "" {
		dockerCommand.LocalProjectName = config.ProjectName
	}

	return dockerCommand, nil
}

// buildRuntime constructs the selected container runtime and the closers
// that must run on shutdown. The Docker backend is gated behind -tags docker
// (see runtime_docker.go / runtime_nodocker.go); the Podman backend connects
// lazily to the engine socket.
func buildRuntime(backend string, osCommand *OSCommand) (runtime.ContainerRuntime, []io.Closer, error) {
	switch backend {
	case "docker":
		// The Docker backend is compiled in only with -tags docker; the
		// default Podman-only build returns an error here. See
		// runtime_docker.go / runtime_nodocker.go.
		return newDockerBackend(osCommand)

	case "podman":
		rt, err := podmanruntime.NewFromEnv()
		if err != nil {
			ogLog.Fatal(err)
		}
		return rt, []io.Closer{rt}, nil

	default:
		return nil, nil, fmt.Errorf("unknown runtime %q: set runtime to \"docker\" or \"podman\"", backend)
	}
}

// IsProjectScoped reports whether lazypodman should be scoped to a single
// compose project — either because we're inside a compose directory or
// because the user passed -p. When false, the project/services panels are
// hidden and all containers are shown in a flat list.
func (c *ContainerCommand) IsProjectScoped() bool {
	return c.InDockerComposeProject || c.Config.ProjectName != ""
}

func (c *ContainerCommand) setDockerComposeCommand(config *config.AppConfig) {
	if config.UserConfig.CommandTemplates.DockerCompose != "docker compose" {
		return
	}

	// it's possible that a user is still using docker-compose, so we'll check if 'docker comopose' is available, and if not, we'll fall back to 'docker-compose'
	err := c.OSCommand.RunCommand("docker compose version")
	if err != nil {
		config.UserConfig.CommandTemplates.DockerCompose = "docker-compose"
	}
}

func (c *ContainerCommand) Close() error {
	return utils.CloseMany(c.Closers)
}

func (c *ContainerCommand) CreateClientStatMonitor(container *Container) {
	container.MonitoringStats = true
	defer func() { container.MonitoringStats = false }()

	ctx := context.Background()
	stream, err := c.Runtime.ContainerStats(ctx, container.ID)
	if err != nil {
		// Not creating an error panel — if we've disconnected from the engine
		// we'll already have one shown by the event loop.
		c.Log.Error(err)
		return
	}

	for snapshot := range stream {
		stats := statsFromDomain(snapshot)
		recordedStats := &RecordedStats{
			ClientStats: stats,
			DerivedStats: DerivedStats{
				CPUPercentage:    stats.CalculateContainerCPUPercentage(),
				MemoryPercentage: stats.CalculateContainerMemoryUsage(),
			},
			RecordedAt: time.Now(),
		}
		container.appendStats(recordedStats, c.Config.UserConfig.Stats.MaxDuration)
	}
}

func (c *ContainerCommand) RefreshContainersAndServices(currentContainers []*Container) ([]*Container, []*Service, error) {
	c.ServiceMutex.Lock()
	defer c.ServiceMutex.Unlock()

	containers, err := c.GetContainers(currentContainers)
	if err != nil {
		return nil, nil, err
	}

	// Derive services from container labels (covers all projects)
	services := c.GetServicesFromContainers(containers)

	var composeServices []*Service
	if c.InDockerComposeProject {
		composeServices, err = c.GetServices()
		if err != nil {
			c.Log.Warn("Failed to get compose services: " + err.Error())
		}
	}

	// Determine the local project name before merging services, since
	// mergeServices needs it. We match compose service names against container
	// labels to handle cases where the project name differs from the directory
	// name (e.g. a `name:` directive in the compose file).
	if c.LocalProjectName == "" && c.InDockerComposeProject && composeServices != nil {
		for _, ctr := range containers {
			if ctr.ProjectName == "" || ctr.ServiceName == "" {
				continue
			}
			for _, svc := range composeServices {
				if ctr.ServiceName == svc.Name {
					c.LocalProjectName = ctr.ProjectName
					break
				}
			}
			if c.LocalProjectName != "" {
				break
			}
		}
		// Fall back to directory name
		if c.LocalProjectName == "" && c.Config.ProjectDir != "" {
			c.LocalProjectName = filepath.Base(c.Config.ProjectDir)
		}
	}

	// Merge compose services (which include stopped services) with
	// container-derived services from all projects
	if composeServices != nil {
		services = c.mergeServices(services, composeServices)
	}

	c.assignContainersToServices(containers, services)

	return containers, services, nil
}

// GetServicesFromContainers derives services from container labels for all projects
func (c *ContainerCommand) GetServicesFromContainers(containers []*Container) []*Service {
	// Use project+service as key to avoid duplicates
	type serviceKey struct {
		project string
		service string
	}
	seen := make(map[serviceKey]bool)
	services := make([]*Service, 0, len(containers))

	for _, ctr := range containers {
		if ctr.ServiceName == "" || ctr.OneOff {
			continue
		}
		key := serviceKey{project: ctr.ProjectName, service: ctr.ServiceName}
		if seen[key] {
			continue
		}
		seen[key] = true
		services = append(services, &Service{
			Name:             ctr.ServiceName,
			ID:               ctr.ProjectName + "-" + ctr.ServiceName,
			ProjectName:      ctr.ProjectName,
			OSCommand:        c.OSCommand,
			Log:              c.Log,
			ContainerCommand: c,
		})
	}

	return services
}

// mergeServices merges compose services (which may lack ProjectName) with
// container-derived services. Compose services take priority because they
// include services without running containers.
func (c *ContainerCommand) mergeServices(containerServices []*Service, composeServices []*Service) []*Service {
	// Set project name on compose services
	for _, svc := range composeServices {
		if svc.ProjectName == "" {
			svc.ProjectName = c.LocalProjectName
		}
	}

	// Build a set of compose service names for the local project
	composeServiceNames := make(map[string]bool)
	for _, svc := range composeServices {
		composeServiceNames[svc.Name] = true
	}

	// Start with compose services, then add container-derived services
	// that aren't already covered by compose (i.e. from other projects)
	result := make([]*Service, 0, len(composeServices)+len(containerServices))
	result = append(result, composeServices...)

	for _, svc := range containerServices {
		if svc.ProjectName == c.LocalProjectName && composeServiceNames[svc.Name] {
			continue // already covered by compose service
		}
		result = append(result, svc)
	}

	return result
}

// GetProjectNames returns all unique project names from containers
func (c *ContainerCommand) GetProjectNames(containers []*Container) []string {
	seen := make(map[string]bool)
	var names []string
	for _, ctr := range containers {
		if ctr.ProjectName != "" && !seen[ctr.ProjectName] {
			seen[ctr.ProjectName] = true
			names = append(names, ctr.ProjectName)
		}
	}
	sort.Strings(names)
	return names
}

func (c *ContainerCommand) assignContainersToServices(containers []*Container, services []*Service) {
L:
	for _, service := range services {
		for _, ctr := range containers {
			if !ctr.OneOff && ctr.ServiceName == service.Name && ctr.ProjectName == service.ProjectName {
				service.Container = ctr
				continue L
			}
		}
		service.Container = nil
	}
}

// GetContainers gets the docker containers
func (c *ContainerCommand) GetContainers(existingContainers []*Container) ([]*Container, error) {
	c.ContainerMutex.Lock()
	defer c.ContainerMutex.Unlock()

	containers, err := c.Runtime.ListContainers(context.Background())
	if err != nil {
		return nil, err
	}

	ownContainers := make([]*Container, len(containers))

	for i, ctr := range containers {
		var newContainer *Container

		// check if we already have data stored against the container
		for _, existingContainer := range existingContainers {
			if existingContainer.ID == ctr.ID {
				newContainer = existingContainer
				break
			}
		}

		// initialise the container if it's completely new
		if newContainer == nil {
			newContainer = &Container{
				ID:               ctr.ID,
				Runtime:          c.Runtime,
				OSCommand:        c.OSCommand,
				Log:              c.Log,
				ContainerCommand: c,
				Tr:               c.Tr,
			}
		}

		newContainer.Container = ctr
		// if the container is made with a name label we will use that
		if name, ok := ctr.Labels["name"]; ok {
			newContainer.Name = name
		} else if primary := ctr.PrimaryName(); primary != "" {
			newContainer.Name = primary
		} else {
			newContainer.Name = ctr.ID
		}
		newContainer.ServiceName = ctr.Labels["com.docker.compose.service"]
		newContainer.ProjectName = ctr.Labels["com.docker.compose.project"]
		newContainer.ContainerNumber = ctr.Labels["com.docker.compose.container"]
		newContainer.OneOff = ctr.Labels["com.docker.compose.oneoff"] == "True"

		ownContainers[i] = newContainer
	}

	c.SetContainerDetails(ownContainers)

	return ownContainers, nil
}

// GetServices gets services
func (c *ContainerCommand) GetServices() ([]*Service, error) {
	if !c.InDockerComposeProject {
		return nil, nil
	}

	composeCommand := c.Config.UserConfig.CommandTemplates.DockerCompose
	output, err := c.OSCommand.RunCommandWithOutput(fmt.Sprintf("%s config --services", composeCommand))
	if err != nil {
		return nil, err
	}

	// output looks like:
	// service1
	// service2

	lines := utils.SplitLines(output)
	services := make([]*Service, len(lines))
	for i, str := range lines {
		services[i] = &Service{
			Name:             str,
			ID:               c.LocalProjectName + "-" + str,
			ProjectName:      c.LocalProjectName,
			OSCommand:        c.OSCommand,
			Log:              c.Log,
			ContainerCommand: c,
		}
	}

	return services, nil
}

func (c *ContainerCommand) RefreshContainerDetails(containers []*Container) error {
	c.ContainerMutex.Lock()
	defer c.ContainerMutex.Unlock()

	c.SetContainerDetails(containers)

	return nil
}

// Attaches the details returned from docker inspect to each of the containers
// this contains a bit more info than what you get from the go-docker client
func (c *ContainerCommand) SetContainerDetails(containers []*Container) {
	wg := sync.WaitGroup{}
	for _, ctr := range containers {
		ctr := ctr
		wg.Add(1)
		go func() {
			details, err := c.Runtime.InspectContainer(context.Background(), ctr.ID)
			if err != nil {
				c.Log.Error(err)
			} else {
				ctr.Details = details
				ctr.DetailsFetched = true
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

// ViewAllLogs attaches to a subprocess viewing all the logs from docker-compose
func (c *ContainerCommand) ViewAllLogs(project *Project) (*exec.Cmd, error) {
	cmd := c.OSCommand.ExecutableFromString(
		utils.ApplyTemplate(
			c.OSCommand.Config.UserConfig.CommandTemplates.ViewAllLogs,
			c.NewCommandObject(CommandObject{Project: project}),
		),
	)

	c.OSCommand.PrepareForChildren(cmd)

	return cmd, nil
}

// DockerComposeConfig returns the result of 'docker-compose config'
func (c *ContainerCommand) DockerComposeConfig() string {
	return c.DockerComposeConfigForProject(nil)
}

// DockerComposeConfigForProject returns the result of 'docker-compose config' for a specific project
func (c *ContainerCommand) DockerComposeConfigForProject(project *Project) string {
	output, err := c.OSCommand.RunCommandWithOutput(
		utils.ApplyTemplate(
			c.OSCommand.Config.UserConfig.CommandTemplates.DockerComposeConfig,
			c.NewCommandObject(CommandObject{Project: project}),
		),
	)
	if err != nil {
		output = err.Error()
	}
	return output
}
