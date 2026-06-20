package domain

import "time"

// VolumeScope is the visibility of a volume.
type VolumeScope string

const (
	VolumeScopeLocal  VolumeScope = "local"
	VolumeScopeGlobal VolumeScope = "global"
)

// VolumeUsage is the on-disk usage of a volume when the runtime reports
// it. Some runtimes return -1 for fields they do not compute.
type VolumeUsage struct {
	Size     int64
	RefCount int64
}

// VolumeInfo is the view of a volume as rendered by the GUI. Equivalent
// to docker's volume.Volume.
type VolumeInfo struct {
	Name       string
	Driver     string
	Mountpoint string
	Scope      VolumeScope
	CreatedAt  time.Time
	Labels     map[string]string
	Options    map[string]string
	Status     map[string]any
	UsageData  *VolumeUsage // nil when the runtime does not report usage
}
