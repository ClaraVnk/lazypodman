package commands

import (
	"testing"

	"github.com/jesseduffield/lazydocker/pkg/config"
	"github.com/stretchr/testify/assert"
)

// TestSelectBackend covers the runtime selection precedence:
// LAZYPODMAN_RUNTIME env > config `runtime:` > "docker" default.
func TestSelectBackend(t *testing.T) {
	cases := []struct {
		name      string
		env       string
		cfgValue  string
		nilConfig bool
		want      string
	}{
		{"default when nothing set", "", "", false, "docker"},
		{"config selects podman", "", "podman", false, "podman"},
		{"env overrides config", "docker", "podman", false, "docker"},
		{"env wins, trimmed and lowercased", "  Podman  ", "docker", false, "podman"},
		{"nil UserConfig falls back to default", "", "", true, "docker"},
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
			c := &DockerCommand{
				InDockerComposeProject: tc.inDockerComposeProject,
				Config:                 &config.AppConfig{ProjectName: tc.projectName},
			}
			assert.Equal(t, tc.want, c.IsProjectScoped())
		})
	}
}
