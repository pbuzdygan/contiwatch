package dockerwatcher

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
)

type DockerStorageSummary struct {
	Count            int   `json:"count"`
	Active           int   `json:"active"`
	SizeBytes        int64 `json:"size_bytes"`
	ReclaimableBytes int64 `json:"reclaimable_bytes"`
}

type DockerHealthSnapshot struct {
	EngineVersion       string               `json:"engine_version"`
	APIVersion          string               `json:"api_version"`
	OperatingSystem     string               `json:"operating_system"`
	Architecture        string               `json:"architecture"`
	KernelVersion       string               `json:"kernel_version"`
	StorageDriver       string               `json:"storage_driver"`
	LoggingDriver       string               `json:"logging_driver"`
	DockerRootDir       string               `json:"docker_root_dir"`
	CPUs                int                  `json:"cpus"`
	MemoryBytes         int64                `json:"memory_bytes"`
	ContainersRunning   int                  `json:"containers_running"`
	ContainersStopped   int                  `json:"containers_stopped"`
	ContainersPaused    int                  `json:"containers_paused"`
	ContainersUnhealthy int                  `json:"containers_unhealthy"`
	Images              DockerStorageSummary `json:"images"`
	Containers          DockerStorageSummary `json:"containers"`
	Volumes             DockerStorageSummary `json:"volumes"`
	BuildCache          DockerStorageSummary `json:"build_cache"`
	Warnings            []string             `json:"warnings,omitempty"`
}

func (w *Watcher) DockerHealth(ctx context.Context) (DockerHealthSnapshot, error) {
	info, err := w.client.Info(ctx)
	if err != nil {
		return DockerHealthSnapshot{}, err
	}
	version, err := w.client.ServerVersion(ctx)
	if err != nil {
		return DockerHealthSnapshot{}, err
	}

	snapshot := DockerHealthSnapshot{
		EngineVersion:     info.ServerVersion,
		APIVersion:        version.APIVersion,
		OperatingSystem:   info.OperatingSystem,
		Architecture:      info.Architecture,
		KernelVersion:     info.KernelVersion,
		StorageDriver:     info.Driver,
		LoggingDriver:     info.LoggingDriver,
		DockerRootDir:     info.DockerRootDir,
		CPUs:              info.NCPU,
		MemoryBytes:       info.MemTotal,
		ContainersRunning: info.ContainersRunning,
		ContainersStopped: info.ContainersStopped,
		ContainersPaused:  info.ContainersPaused,
		Images:            DockerStorageSummary{Count: info.Images},
		Warnings:          append([]string(nil), info.Warnings...),
	}

	containers, listErr := w.client.ContainerList(ctx, container.ListOptions{All: true})
	if listErr != nil {
		snapshot.Warnings = append(snapshot.Warnings, "container health status unavailable")
	} else {
		for _, item := range containers {
			if strings.Contains(strings.ToLower(item.Status), "(unhealthy)") {
				snapshot.ContainersUnhealthy++
			}
		}
	}

	usage, usageErr := w.client.DiskUsage(ctx, types.DiskUsageOptions{})
	if usageErr != nil {
		snapshot.Warnings = append(snapshot.Warnings, "Docker storage usage unavailable")
		return snapshot, nil
	}
	applyDockerDiskUsage(&snapshot, usage)
	return snapshot, nil
}

func applyDockerDiskUsage(snapshot *DockerHealthSnapshot, usage types.DiskUsage) {
	if snapshot == nil {
		return
	}
	snapshot.Images.Count = len(usage.Images)
	snapshot.Images.SizeBytes = usage.LayersSize
	for _, item := range usage.Images {
		if item == nil {
			continue
		}
		if item.Containers > 0 {
			snapshot.Images.Active++
			continue
		}
		uniqueSize := item.Size - item.SharedSize
		if uniqueSize > 0 {
			snapshot.Images.ReclaimableBytes += uniqueSize
		}
	}

	snapshot.Containers.Count = len(usage.Containers)
	for _, item := range usage.Containers {
		if item == nil {
			continue
		}
		snapshot.Containers.SizeBytes += item.SizeRw
		if item.State == container.StateRunning || item.State == container.StatePaused || item.State == container.StateRestarting {
			snapshot.Containers.Active++
		} else {
			snapshot.Containers.ReclaimableBytes += item.SizeRw
		}
	}

	snapshot.Volumes.Count = len(usage.Volumes)
	for _, item := range usage.Volumes {
		if item == nil || item.UsageData == nil || item.UsageData.Size < 0 {
			continue
		}
		snapshot.Volumes.SizeBytes += item.UsageData.Size
		if item.UsageData.RefCount > 0 {
			snapshot.Volumes.Active++
		} else {
			snapshot.Volumes.ReclaimableBytes += item.UsageData.Size
		}
	}

	snapshot.BuildCache.Count = len(usage.BuildCache)
	for _, item := range usage.BuildCache {
		if item == nil {
			continue
		}
		snapshot.BuildCache.SizeBytes += item.Size
		if item.InUse {
			snapshot.BuildCache.Active++
		} else {
			snapshot.BuildCache.ReclaimableBytes += item.Size
		}
	}
}
