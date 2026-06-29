package gui

import (
	"sort"
	"testing"

	"github.com/ClaraVnk/lazypodman/pkg/commands"
	"github.com/ClaraVnk/lazypodman/pkg/domain"
	"github.com/stretchr/testify/assert"
)

func sampleContainers() []*commands.Container {
	return []*commands.Container{
		{
			ID:   "1",
			Name: "1",
			Container: domain.ContainerInfo{
				State: domain.ContainerStateExited,
			},
		},
		{
			ID:   "2",
			Name: "2",
			Container: domain.ContainerInfo{
				State: domain.ContainerStateRunning,
			},
		},
		{
			ID:   "3",
			Name: "3",
			Container: domain.ContainerInfo{
				State: domain.ContainerStateRunning,
			},
		},
		{
			ID:   "4",
			Name: "4",
			Container: domain.ContainerInfo{
				State: domain.ContainerStateCreated,
			},
		},
	}
}

func expectedPerStatusContainers() []*commands.Container {
	return []*commands.Container{
		{
			ID:   "2",
			Name: "2",
			Container: domain.ContainerInfo{
				State: domain.ContainerStateRunning,
			},
		},
		{
			ID:   "3",
			Name: "3",
			Container: domain.ContainerInfo{
				State: domain.ContainerStateRunning,
			},
		},
		{
			ID:   "1",
			Name: "1",
			Container: domain.ContainerInfo{
				State: domain.ContainerStateExited,
			},
		},
		{
			ID:   "4",
			Name: "4",
			Container: domain.ContainerInfo{
				State: domain.ContainerStateCreated,
			},
		},
	}
}

func expectedLegacySortedContainers() []*commands.Container {
	return []*commands.Container{
		{
			ID:   "1",
			Name: "1",
			Container: domain.ContainerInfo{
				State: domain.ContainerStateExited,
			},
		},
		{
			ID:   "2",
			Name: "2",
			Container: domain.ContainerInfo{
				State: domain.ContainerStateRunning,
			},
		},
		{
			ID:   "3",
			Name: "3",
			Container: domain.ContainerInfo{
				State: domain.ContainerStateRunning,
			},
		},
		{
			ID:   "4",
			Name: "4",
			Container: domain.ContainerInfo{
				State: domain.ContainerStateCreated,
			},
		},
	}
}

func assertEqualContainers(t *testing.T, left *commands.Container, right *commands.Container) {
	t.Helper()
	assert.Equal(t, left.Container.State, right.Container.State)
	assert.Equal(t, left.Container.ID, right.Container.ID)
	assert.Equal(t, left.Name, right.Name)
}

func TestSortContainers(t *testing.T) {
	actual := sampleContainers()

	expected := expectedPerStatusContainers()

	sort.Slice(actual, func(i, j int) bool {
		return sortContainers(actual[i], actual[j], false)
	})

	assert.Equal(t, len(actual), len(expected))

	for i := 0; i < len(actual); i++ {
		assertEqualContainers(t, expected[i], actual[i])
	}
}

func TestLegacySortedContainers(t *testing.T) {
	actual := sampleContainers()

	expected := expectedLegacySortedContainers()

	sort.Slice(actual, func(i, j int) bool {
		return sortContainers(actual[i], actual[j], true)
	})

	assert.Equal(t, len(actual), len(expected))

	for i := 0; i < len(actual); i++ {
		assertEqualContainers(t, expected[i], actual[i])
	}
}
