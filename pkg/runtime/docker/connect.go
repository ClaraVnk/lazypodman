//go:build docker

package docker

import (
	"fmt"
	"os"

	cliconfig "github.com/docker/cli/cli/config"
	ddocker "github.com/docker/cli/cli/context/docker"
	ctxstore "github.com/docker/cli/cli/context/store"
	dockerclient "github.com/docker/docker/client"
)

const dockerHostEnvKey = "DOCKER_HOST"

// NewFromEnv builds a *Runtime against the Docker engine the user's
// environment points at. Resolution order matches the docker CLI:
//
//  1. DOCKER_HOST environment variable, when set;
//  2. host of the current docker context (DOCKER_CONTEXT or
//     ~/.docker/config.json);
//  3. the operating-system default socket (defaultDockerHost).
//
// The returned runtime has API version negotiation enabled and respects
// the TLS client config from the environment.
//
// SSH tunneling for ssh:// hosts is NOT handled here — callers needing
// that should resolve the host themselves (via ResolveDockerHost) and
// pass it to New(client.NewClientWithOpts(...)). The SSH tunnel helper
// lives in pkg/commands/ssh and stays there until Phase 1d wires
// pkg/commands to construct runtimes via this package.
func NewFromEnv() (*Runtime, error) {
	host, err := ResolveDockerHost()
	if err != nil {
		return nil, fmt.Errorf("resolve docker host: %w", err)
	}
	return NewWithHost(host)
}

// NewWithHost builds a *Runtime against an explicit Docker host string,
// bypassing env-based resolution. It is the entry point for callers that
// resolve the host themselves — for example pkg/commands, which must set
// up an SSH tunnel and mutate DOCKER_HOST before the client is built.
// API version negotiation is enabled and the TLS client config comes from
// the environment.
func NewWithHost(host string) (*Runtime, error) {
	cli, err := dockerclient.NewClientWithOpts(
		dockerclient.WithTLSClientConfigFromEnv(),
		dockerclient.WithAPIVersionNegotiation(),
		dockerclient.WithHost(host),
	)
	if err != nil {
		return nil, fmt.Errorf("docker client: %w", err)
	}
	return New(cli), nil
}

// ResolveDockerHost reproduces the docker CLI's host-resolution logic:
//   - $DOCKER_HOST if set;
//   - host of the docker context named by $DOCKER_CONTEXT or stored in
//     ~/.docker/config.json;
//   - defaultDockerHost otherwise.
//
// Returned with empty error even when falling back to the default — the
// only error case is a malformed docker config file or context store.
func ResolveDockerHost() (string, error) {
	if h := os.Getenv(dockerHostEnvKey); h != "" {
		return h, nil
	}

	currentContext := os.Getenv("DOCKER_CONTEXT")
	if currentContext == "" {
		cf, err := cliconfig.Load(cliconfig.Dir())
		if err != nil {
			return "", err
		}
		currentContext = cf.CurrentContext
	}

	if currentContext == "" || currentContext == "default" {
		return defaultDockerHost, nil
	}

	storeConfig := ctxstore.NewConfig(
		func() any { return &ddocker.EndpointMeta{} },
		ctxstore.EndpointTypeGetter(ddocker.DockerEndpoint, func() any { return &ddocker.EndpointMeta{} }),
	)
	st := ctxstore.New(cliconfig.ContextStoreDir(), storeConfig)
	md, err := st.GetMetadata(currentContext)
	if err != nil {
		return "", err
	}
	endpoint, ok := md.Endpoints[ddocker.DockerEndpoint]
	if !ok {
		return defaultDockerHost, nil
	}
	meta, ok := endpoint.(ddocker.EndpointMeta)
	if !ok {
		return "", fmt.Errorf("expected docker.EndpointMeta, got %T", endpoint)
	}
	if meta.Host != "" {
		return meta.Host, nil
	}
	return defaultDockerHost, nil
}
