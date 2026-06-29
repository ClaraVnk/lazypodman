package commands

import (
	"strings"
	"testing"

	"github.com/ClaraVnk/lazypodman/pkg/config"
	"github.com/ClaraVnk/lazypodman/pkg/i18n"
	"github.com/ClaraVnk/lazypodman/pkg/utils"
	"github.com/stretchr/testify/assert"
)

// TestEngineName covers the human-facing backend label used in messages.
func TestEngineName(t *testing.T) {
	cases := map[string]string{
		"podman": "Podman",
		"":       "container engine",
		"weird":  "container engine",
	}
	for backend, want := range cases {
		assert.Equal(t, want, (&ContainerCommand{Backend: backend}).EngineName())
	}
}

// TestConnectionFailedIsBackendAware guards #13: the connection-failure
// message must carry the {{engine}} placeholder and, once rendered with the
// active backend, name it without leaking "docker" when running Podman.
func TestConnectionFailedIsBackendAware(t *testing.T) {
	tr := i18n.GetTranslationSets()["en"]
	assert.Contains(t, tr.ConnectionFailed, "{{engine}}")

	rendered := utils.ResolvePlaceholderString(tr.ConnectionFailed, map[string]string{"engine": "Podman"})
	assert.Contains(t, rendered, "Podman")
	assert.NotContains(t, strings.ToLower(rendered), "docker")
}

// TestSelectBackend covers the runtime selection precedence:
// LAZYPODMAN_RUNTIME env > config `runtime:` > "podman" default. selectBackend
// only resolves the raw string; buildRuntime is what rejects anything other
// than "podman" (lazypodman is Podman-only since ADR 0002 Phase 6).
func TestSelectBackend(t *testing.T) {
	cases := []struct {
		name      string
		env       string
		cfgValue  string
		nilConfig bool
		want      string
	}{
		{"default when nothing set", "", "", false, "podman"},
		{"config value is resolved", "", "podman", false, "podman"},
		{"env overrides config", "podman", "other", false, "podman"},
		{"env wins, trimmed and lowercased", "  Podman  ", "other", false, "podman"},
		{"nil UserConfig falls back to default", "", "", true, "podman"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(runtimeEnvKey, tc.env)
			cfg := &config.AppConfig{UserConfig: &config.UserConfig{Runtime: tc.cfgValue}}
			if tc.nilConfig {
				cfg = &config.AppConfig{}
			}
			assert.Equal(t, tc.want, selectBackend(cfg))
		})
	}
}

// TestIsProjectScoped covers the predicate that drives whether the
// project/services panels appear and whether the containers panel filters by
// project. The "outside compose dir + -p" case is the regression we fixed
// after PR #776 silently disabled it.
func TestIsProjectScoped(t *testing.T) {
	cases := []struct {
		name                   string
		inDockerComposeProject bool
		projectName            string
		want                   bool
	}{
		{"inside compose dir, no -p", true, "", true},
		{"inside compose dir, with -p", true, "myproject", true},
		{"outside compose dir, no -p", false, "", false},
		{"outside compose dir, with -p", false, "myproject", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := &ContainerCommand{
				InDockerComposeProject: tc.inDockerComposeProject,
				Config:                 &config.AppConfig{ProjectName: tc.projectName},
			}
			assert.Equal(t, tc.want, c.IsProjectScoped())
		})
	}
}
