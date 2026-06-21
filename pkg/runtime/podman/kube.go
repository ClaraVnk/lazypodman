package podman

import (
	"context"
	"io"

	"github.com/containers/podman/v5/pkg/bindings/generate"

	"github.com/jesseduffield/lazydocker/pkg/runtime"
)

// Compile-time check that the Podman runtime advertises the optional
// KubeGenerator capability.
var _ runtime.KubeGenerator = (*Runtime)(nil)

// GenerateKube exports the named containers/pods as Kubernetes YAML.
func (r *Runtime) GenerateKube(ctx context.Context, names []string) ([]byte, error) {
	conn, err := r.client()
	if err != nil {
		return nil, err
	}
	report, err := generate.Kube(conn, names, nil)
	if err != nil {
		return nil, mapErr("generate kube", err)
	}
	if report == nil || report.Reader == nil {
		return nil, nil
	}
	data, err := io.ReadAll(report.Reader)
	if err != nil {
		return nil, mapErr("generate kube", err)
	}
	return data, nil
}
