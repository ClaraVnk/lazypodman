//go:build docker && windows

package docker

const defaultDockerHost = "npipe:////./pipe/docker_engine"
