package dockerwatcher

import "testing"

func TestContainerIDsFromRuntimeMetadataFindsCurrentDockerContainer(t *testing.T) {
	const id = "23cfe06c9c2a79a3be3bd7bd3c4cd16dce1013d999d89548331b8686d98a941b"
	mountInfo := "420 399 0:78 /docker/containers/" + id + "/hostname /etc/hostname rw,relatime - ext4 /dev/sda rw"
	cgroup := "0::/system.slice/docker-" + id + ".scope"

	ids := containerIDsFromRuntimeMetadata(mountInfo, cgroup)

	if len(ids) != 1 || ids[0] != id {
		t.Fatalf("expected current container ID, got %v", ids)
	}
}

func TestMatchesContainerIDUsesRuntimeIDWhenHostnameIsStale(t *testing.T) {
	const id = "23cfe06c9c2a79a3be3bd7bd3c4cd16dce1013d999d89548331b8686d98a941b"

	if !matchesContainerID(id, []string{"542fc63dcd86", id}) {
		t.Fatal("expected runtime metadata ID to match despite stale hostname")
	}
}

func TestDockerGeneratedHostnameDetection(t *testing.T) {
	if !isDockerGeneratedHostname("542fc63dcd86") {
		t.Fatal("expected Docker-generated hostname to be detected")
	}
	if isDockerGeneratedHostname("contiwatch-agent") {
		t.Fatal("expected explicit hostname to be preserved")
	}
	if isDockerGeneratedHostname("ABCDEF123456") {
		t.Fatal("expected non-default uppercase hostname to be preserved")
	}
}
