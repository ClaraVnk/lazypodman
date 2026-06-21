package runtime

import "context"

// KubeGenerator is an optional capability implemented by backends that can
// export containers or pods as Kubernetes YAML (Podman's
// `podman generate kube`). The Docker backend does not implement it.
// Callers discover the capability with a type assertion:
//
//	if kg, ok := rt.(runtime.KubeGenerator); ok {
//		yaml, err := kg.GenerateKube(ctx, []string{name})
//	}
//
// See docs/adr/0006-podman-pods-and-capabilities.md.
type KubeGenerator interface {
	// GenerateKube returns the Kubernetes YAML for the named containers
	// and/or pods.
	GenerateKube(ctx context.Context, names []string) ([]byte, error)
}
