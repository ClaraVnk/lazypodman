package podman

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ClaraVnk/lazypodman/pkg/domain"
	"github.com/ClaraVnk/lazypodman/pkg/runtime"
)

// Compile-time check that the Podman runtime advertises the optional
// QuadletManager capability.
var _ runtime.QuadletManager = (*Runtime)(nil)

// systemctl runs `systemctl --user <args>` via the injectable runner
// (os/exec by default).
func (r *Runtime) systemctl(ctx context.Context, args ...string) ([]byte, error) {
	run := r.runCommand
	if run == nil {
		run = func(ctx context.Context, name string, a ...string) ([]byte, error) {
			return exec.CommandContext(ctx, name, a...).Output()
		}
	}
	return run(ctx, "systemctl", append([]string{"--user"}, args...)...)
}

// quadletDir is the rootless quadlet source directory, matching the
// default rootless Podman target.
func quadletDir() string {
	if xdg := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); xdg != "" {
		return filepath.Join(xdg, "containers", "systemd")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "containers", "systemd")
}

// quadletUnitName derives the generated systemd unit name and type from a
// quadlet source filename, following Podman's generator conventions.
func quadletUnitName(filename string) (string, domain.QuadletType, bool) {
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)
	switch ext {
	case ".container":
		return base + ".service", domain.QuadletContainer, true
	case ".kube":
		return base + ".service", domain.QuadletKube, true
	case ".pod":
		return base + "-pod.service", domain.QuadletPod, true
	case ".network":
		return base + "-network.service", domain.QuadletNetwork, true
	case ".volume":
		return base + "-volume.service", domain.QuadletVolume, true
	case ".image":
		return base + "-image.service", domain.QuadletImage, true
	case ".build":
		return base + "-build.service", domain.QuadletBuild, true
	default:
		return "", "", false
	}
}

// ListQuadlets enumerates the quadlet source files and reports each
// generated unit's active state.
func (r *Runtime) ListQuadlets(ctx context.Context) ([]domain.Quadlet, error) {
	dir := quadletDir()
	if dir == "" {
		return nil, nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, mapErr("list quadlets", err)
	}

	var out []domain.Quadlet
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		unit, qType, ok := quadletUnitName(e.Name())
		if !ok {
			continue
		}
		state := r.quadletActiveState(ctx, unit)
		out = append(out, domain.Quadlet{
			Name:        strings.TrimSuffix(e.Name(), filepath.Ext(e.Name())),
			UnitName:    unit,
			Type:        qType,
			SourcePath:  filepath.Join(dir, e.Name()),
			ActiveState: state,
			Active:      state == "active",
		})
	}
	return out, nil
}

// quadletActiveState queries the unit's ActiveState, returning "unknown"
// when systemctl cannot report it.
func (r *Runtime) quadletActiveState(ctx context.Context, unit string) string {
	out, err := r.systemctl(ctx, "show", unit, "--property=ActiveState")
	if err != nil {
		return "unknown"
	}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if v, ok := strings.CutPrefix(strings.TrimSpace(line), "ActiveState="); ok {
			return v
		}
	}
	return "unknown"
}

func (r *Runtime) StartQuadlet(ctx context.Context, unit string) error {
	_, err := r.systemctl(ctx, "start", unit)
	return mapErr("start quadlet", err)
}

func (r *Runtime) StopQuadlet(ctx context.Context, unit string) error {
	_, err := r.systemctl(ctx, "stop", unit)
	return mapErr("stop quadlet", err)
}

func (r *Runtime) RestartQuadlet(ctx context.Context, unit string) error {
	_, err := r.systemctl(ctx, "restart", unit)
	return mapErr("restart quadlet", err)
}
