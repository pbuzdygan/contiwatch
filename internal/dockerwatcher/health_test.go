package dockerwatcher

import (
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/volume"
)

func TestApplyDockerDiskUsageSummarizesReclaimableStorage(t *testing.T) {
	snapshot := DockerHealthSnapshot{}
	applyDockerDiskUsage(&snapshot, types.DiskUsage{
		LayersSize: 1000,
		Images: []*image.Summary{
			{Containers: 1, Size: 700, SharedSize: 100},
			{Containers: 0, Size: 500, SharedSize: 100},
		},
		Containers: []*container.Summary{
			{State: container.StateRunning, SizeRw: 100},
			{State: container.StateExited, SizeRw: 200},
		},
		Volumes: []*volume.Volume{
			{UsageData: &volume.UsageData{RefCount: 1, Size: 300}},
			{UsageData: &volume.UsageData{RefCount: 0, Size: 400}},
		},
		BuildCache: []*build.CacheRecord{
			{InUse: true, Size: 500},
			{InUse: false, Size: 600},
		},
	})

	if snapshot.Images.SizeBytes != 1000 || snapshot.Images.ReclaimableBytes != 400 || snapshot.Images.Active != 1 {
		t.Fatalf("unexpected images summary: %#v", snapshot.Images)
	}
	if snapshot.Containers.SizeBytes != 300 || snapshot.Containers.ReclaimableBytes != 200 || snapshot.Containers.Active != 1 {
		t.Fatalf("unexpected containers summary: %#v", snapshot.Containers)
	}
	if snapshot.Volumes.SizeBytes != 700 || snapshot.Volumes.ReclaimableBytes != 400 || snapshot.Volumes.Active != 1 {
		t.Fatalf("unexpected volumes summary: %#v", snapshot.Volumes)
	}
	if snapshot.BuildCache.SizeBytes != 1100 || snapshot.BuildCache.ReclaimableBytes != 600 || snapshot.BuildCache.Active != 1 {
		t.Fatalf("unexpected build cache summary: %#v", snapshot.BuildCache)
	}
}
