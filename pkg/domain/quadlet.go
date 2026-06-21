package domain

// QuadletType is the kind of object a quadlet source file declares.
type QuadletType string

const (
	QuadletContainer QuadletType = "container"
	QuadletPod       QuadletType = "pod"
	QuadletNetwork   QuadletType = "network"
	QuadletVolume    QuadletType = "volume"
	QuadletKube      QuadletType = "kube"
	QuadletImage     QuadletType = "image"
	QuadletBuild     QuadletType = "build"
)

// Quadlet is a Podman quadlet: a source unit file under the user's systemd
// directory and the systemd service it generates. Quadlets are a
// Podman-native concept managed via systemd; see the QuadletManager
// capability in pkg/runtime and docs/adr/0007-quadlets.md.
type Quadlet struct {
	Name        string      // base name, e.g. "web" for web.container
	UnitName    string      // generated systemd unit, e.g. "web.service"
	Type        QuadletType // container | pod | network | volume | kube | ...
	SourcePath  string      // path to the source file
	ActiveState string      // systemd ActiveState: active | inactive | failed
	Active      bool        // ActiveState == "active"
}
