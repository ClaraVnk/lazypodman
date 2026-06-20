package domain

import "time"

// CPUStats is the CPU usage of a container at a point in time.
type CPUStats struct {
	TotalUsage  uint64
	SystemUsage uint64
	OnlineCPUs  uint32
	PerCPUUsage []uint64
}

// MemoryStats is the memory usage of a container at a point in time.
type MemoryStats struct {
	Usage    uint64 // bytes
	MaxUsage uint64 // bytes
	Limit    uint64 // bytes
}

// BlkioEntry is one block I/O metric.
type BlkioEntry struct {
	Major uint64
	Minor uint64
	Op    string
	Value uint64
}

// BlkioStats is the block I/O of a container at a point in time.
type BlkioStats struct {
	ReadBytes  uint64
	WriteBytes uint64
	Entries    []BlkioEntry // optional, when the runtime reports a breakdown
}

// NetworkStats is the per-interface network usage of a container.
type NetworkStats struct {
	RxBytes   uint64
	RxPackets uint64
	RxErrors  uint64
	RxDropped uint64
	TxBytes   uint64
	TxPackets uint64
	TxErrors  uint64
	TxDropped uint64
}

// Stats is one snapshot of a container's resource usage.
type Stats struct {
	Time     time.Time
	PreCPU   CPUStats
	CPU      CPUStats
	Memory   MemoryStats
	BlkIO    BlkioStats
	Networks map[string]NetworkStats // keyed by interface name
}

// AttachStream is the full-duplex stream returned when attaching to a
// container. The caller writes to Stdin (if interactive) and reads from
// Stdout/Stderr. Close releases the underlying connection.
type AttachStream interface {
	Stdin() interface{ Write(p []byte) (int, error) }
	Stdout() interface{ Read(p []byte) (int, error) }
	Stderr() interface{ Read(p []byte) (int, error) }
	Close() error
}
