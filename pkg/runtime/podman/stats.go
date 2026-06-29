package podman

import (
	"context"

	"github.com/containers/podman/v5/libpod/define"
	"github.com/containers/podman/v5/pkg/bindings/containers"

	"github.com/ClaraVnk/lazypodman/pkg/domain"
)

// ContainerStats opens a streaming stats subscription on a container.
// Podman reports cumulative CPU counters per sample but no previous
// sample, so we carry the prior report forward as PreCPU, letting the GUI
// compute the percentage with the same delta formula it uses for Docker.
func (r *Runtime) ContainerStats(ctx context.Context, id string) (<-chan domain.Stats, error) {
	conn, err := r.client()
	if err != nil {
		return nil, err
	}
	// Derive the stream context from the connection (so it keeps the
	// bindings client) but make it cancellable, so when the caller's ctx
	// is done we tear the bindings stream down instead of leaking its
	// goroutine and HTTP connection. Mirrors ContainerLogs.
	streamCtx, cancel := context.WithCancel(conn)
	reports, err := containers.Stats(streamCtx, []string{id}, new(containers.StatsOptions).WithStream(true))
	if err != nil {
		cancel()
		return nil, mapErr("container stats", err)
	}
	out := make(chan domain.Stats)
	go func() {
		defer close(out)
		defer cancel()
		var prev *define.ContainerStats
		for {
			select {
			case <-ctx.Done():
				return
			case rep, ok := <-reports:
				if !ok {
					return
				}
				if rep.Error != nil || len(rep.Stats) == 0 {
					continue
				}
				cur := rep.Stats[0]
				select {
				case out <- podmanStatsToDomain(cur, prev):
				case <-ctx.Done():
					return
				}
				snapshot := cur
				prev = &snapshot
			}
		}
	}()
	return out, nil
}

func podmanStatsToDomain(c define.ContainerStats, prev *define.ContainerStats) domain.Stats {
	s := domain.Stats{
		CPU: domain.CPUStats{
			TotalUsage:  c.CPUNano,
			SystemUsage: c.SystemNano,
			OnlineCPUs:  uint32(len(c.PerCPU)),
			PerCPUUsage: append([]uint64(nil), c.PerCPU...),
		},
		Memory: domain.MemoryStats{
			Usage: c.MemUsage,
			Limit: c.MemLimit,
		},
		BlkIO: domain.BlkioStats{
			ReadBytes:  c.BlockInput,
			WriteBytes: c.BlockOutput,
		},
	}
	if prev != nil {
		s.PreCPU = domain.CPUStats{
			TotalUsage:  prev.CPUNano,
			SystemUsage: prev.SystemNano,
			OnlineCPUs:  uint32(len(prev.PerCPU)),
			PerCPUUsage: append([]uint64(nil), prev.PerCPU...),
		}
	}
	if len(c.Network) > 0 {
		s.Networks = make(map[string]domain.NetworkStats, len(c.Network))
		for name, n := range c.Network {
			s.Networks[name] = domain.NetworkStats{
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
	return s
}
