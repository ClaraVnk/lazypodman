//go:build docker

package docker

import (
	"context"
	"encoding/json"
	"errors"
	"io"

	dockercontainer "github.com/docker/docker/api/types/container"

	"github.com/ClaraVnk/lazypodman/pkg/domain"
	"github.com/ClaraVnk/lazypodman/pkg/runtime"
)

// ContainerStats opens a streaming stats subscription on a container. The
// returned channel is closed when ctx is cancelled or the stream errors.
func (r *Runtime) ContainerStats(ctx context.Context, id string) (<-chan domain.Stats, error) {
	resp, err := r.cli.ContainerStats(ctx, id, true)
	if err != nil {
		return nil, mapErr("container stats", err)
	}
	out := make(chan domain.Stats)
	go func() {
		defer close(out)
		defer resp.Body.Close()
		dec := json.NewDecoder(resp.Body)
		for {
			var s dockercontainer.StatsResponse
			if err := dec.Decode(&s); err != nil {
				if !errors.Is(err, io.EOF) && ctx.Err() == nil {
					// Stream broke unexpectedly. Nothing to signal to
					// the caller besides closing the channel — Stats
					// streams are best-effort and reopened on the next
					// refresh tick by the GUI layer.
					return
				}
				return
			}
			select {
			case out <- statsToDomain(s):
			case <-ctx.Done():
				return
			}
		}
	}()
	return out, nil
}

func statsToDomain(s dockercontainer.StatsResponse) domain.Stats {
	out := domain.Stats{
		Time: s.Read,
		PreCPU: domain.CPUStats{
			TotalUsage:  s.PreCPUStats.CPUUsage.TotalUsage,
			SystemUsage: s.PreCPUStats.SystemUsage,
			OnlineCPUs:  s.PreCPUStats.OnlineCPUs,
			PerCPUUsage: append([]uint64(nil), s.PreCPUStats.CPUUsage.PercpuUsage...),
		},
		CPU: domain.CPUStats{
			TotalUsage:  s.CPUStats.CPUUsage.TotalUsage,
			SystemUsage: s.CPUStats.SystemUsage,
			OnlineCPUs:  s.CPUStats.OnlineCPUs,
			PerCPUUsage: append([]uint64(nil), s.CPUStats.CPUUsage.PercpuUsage...),
		},
		Memory: domain.MemoryStats{
			Usage:    s.MemoryStats.Usage,
			MaxUsage: s.MemoryStats.MaxUsage,
			Limit:    s.MemoryStats.Limit,
		},
		BlkIO: blkioToDomain(s.BlkioStats),
	}
	if len(s.Networks) > 0 {
		out.Networks = make(map[string]domain.NetworkStats, len(s.Networks))
		for name, n := range s.Networks {
			out.Networks[name] = domain.NetworkStats{
				RxBytes:   n.RxBytes,
				RxPackets: n.RxPackets,
				RxErrors:  n.RxErrors,
				RxDropped: n.RxDropped,
				TxBytes:   n.TxBytes,
				TxPackets: n.TxPackets,
				TxErrors:  n.TxErrors,
				TxDropped: n.TxDropped,
			}
		}
	}
	return out
}

func blkioToDomain(b dockercontainer.BlkioStats) domain.BlkioStats {
	out := domain.BlkioStats{}
	for _, e := range b.IoServiceBytesRecursive {
		switch e.Op {
		case "Read", "read":
			out.ReadBytes += e.Value
		case "Write", "write":
			out.WriteBytes += e.Value
		}
		out.Entries = append(out.Entries, domain.BlkioEntry{
			Major: e.Major,
			Minor: e.Minor,
			Op:    e.Op,
			Value: e.Value,
		})
	}
	return out
}

// AttachContainer is intentionally unsupported in the runtime abstraction.
// Per ADR 0004 the interactive attach stays a CLI exec (`docker attach` via
// exec.Cmd) in pkg/commands; both backends return ErrUnsupported here.
func (r *Runtime) AttachContainer(ctx context.Context, id string, opts runtime.AttachOptions) (domain.AttachStream, error) {
	return nil, runtime.ErrUnsupported
}
