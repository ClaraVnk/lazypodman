package podman

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ClaraVnk/lazypodman/pkg/domain"
)

func TestListQuadlets(t *testing.T) {
	tmp := t.TempDir()
	sysd := filepath.Join(tmp, "containers", "systemd")
	if err := os.MkdirAll(sysd, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{
		"web.container": "[Container]\nImage=x\n",
		"db.pod":        "[Pod]\n",
		"ignore.txt":    "not a quadlet",
	} {
		if err := os.WriteFile(filepath.Join(sysd, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("XDG_CONFIG_HOME", tmp)

	var calls [][]string
	r := &Runtime{runCommand: func(ctx context.Context, name string, args ...string) ([]byte, error) {
		calls = append(calls, append([]string{name}, args...))
		return []byte("ActiveState=active\n"), nil
	}}

	qs, err := r.ListQuadlets(context.Background())
	if err != nil {
		t.Fatalf("ListQuadlets: %v", err)
	}
	if len(qs) != 2 {
		t.Fatalf("got %d quadlets, want 2 (ignore.txt skipped): %+v", len(qs), qs)
	}

	byUnit := map[string]domain.Quadlet{}
	for _, q := range qs {
		byUnit[q.UnitName] = q
	}
	if q, ok := byUnit["web.service"]; !ok || q.Type != domain.QuadletContainer || !q.Active {
		t.Errorf("web.container should map to active web.service (container), got %+v", q)
	}
	if q, ok := byUnit["db-pod.service"]; !ok || q.Type != domain.QuadletPod {
		t.Errorf("db.pod should map to db-pod.service (pod), got %+v", q)
	}
	if len(calls) == 0 || calls[0][0] != "systemctl" || calls[0][1] != "--user" {
		t.Errorf("expected `systemctl --user ...`, got %v", calls)
	}
}

func TestQuadletLifecycleArgs(t *testing.T) {
	var calls [][]string
	r := &Runtime{runCommand: func(ctx context.Context, name string, args ...string) ([]byte, error) {
		calls = append(calls, append([]string{name}, args...))
		return nil, nil
	}}
	ctx := context.Background()
	if err := r.StartQuadlet(ctx, "web.service"); err != nil {
		t.Fatal(err)
	}
	if err := r.StopQuadlet(ctx, "web.service"); err != nil {
		t.Fatal(err)
	}
	if err := r.RestartQuadlet(ctx, "web.service"); err != nil {
		t.Fatal(err)
	}
	for i, verb := range []string{"start", "stop", "restart"} {
		if calls[i][0] != "systemctl" || calls[i][1] != "--user" || calls[i][2] != verb || calls[i][3] != "web.service" {
			t.Errorf("call %d = %v, want systemctl --user %s web.service", i, calls[i], verb)
		}
	}
}
