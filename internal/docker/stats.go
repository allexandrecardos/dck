package docker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type StatsSnapshot struct {
	CPUPercent float64
	MemUsage   uint64
	MemLimit   uint64
}

func StatsOnce(ctx context.Context, cli *client.Client, id string) (StatsSnapshot, error) {
	resp, err := cli.ContainerStats(ctx, id, false)
	if err != nil {
		return StatsSnapshot{}, err
	}
	defer resp.Body.Close()

	var stats types.StatsJSON
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return StatsSnapshot{}, fmt.Errorf("decode stats: %w", err)
	}

	return StatsSnapshot{
		CPUPercent: cpuPercent(stats),
		MemUsage:   stats.MemoryStats.Usage,
		MemLimit:   stats.MemoryStats.Limit,
	}, nil
}

func cpuPercent(stats types.StatsJSON) float64 {
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage)
	if cpuDelta <= 0 || systemDelta <= 0 {
		return 0
	}

	onlineCPUs := float64(stats.CPUStats.OnlineCPUs)
	if onlineCPUs == 0 {
		onlineCPUs = float64(len(stats.CPUStats.CPUUsage.PercpuUsage))
		if onlineCPUs == 0 {
			onlineCPUs = 1
		}
	}

	return (cpuDelta / systemDelta) * onlineCPUs * 100.0
}
