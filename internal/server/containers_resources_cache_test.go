package server

import (
	"testing"
	"time"

	"contiwatch/internal/dockerwatcher"
)

func TestContainersResourcesCacheReportsMissingEntries(t *testing.T) {
	srv := &Server{containersResourcesCache: map[string]map[string]containersResourcesCacheEntry{}}
	srv.setContainersResourcesCache("local:test", []dockerwatcher.ContainerResource{{ID: "one", CPUPercent: 1.5}})

	resources, missing, needsRefresh := srv.getContainersResourcesCache(
		"local:test",
		[]string{"one", "two"},
		30*time.Minute,
		6*time.Second,
	)

	if len(resources) != 1 || resources[0].ID != "one" {
		t.Fatalf("expected cached resource one, got %#v", resources)
	}
	if len(missing) != 1 || missing[0] != "two" {
		t.Fatalf("expected missing resource two, got %#v", missing)
	}
	if !needsRefresh {
		t.Fatal("expected partial cache to require refresh")
	}
}

func TestContainersResourcesCacheReturnsCompleteStaleSnapshot(t *testing.T) {
	srv := &Server{containersResourcesCache: map[string]map[string]containersResourcesCacheEntry{
		"local:test": {
			"one": {Resource: dockerwatcher.ContainerResource{ID: "one"}, FetchedAt: time.Now().Add(-10 * time.Second)},
		},
	}}

	resources, missing, needsRefresh := srv.getContainersResourcesCache(
		"local:test",
		[]string{"one"},
		30*time.Minute,
		6*time.Second,
	)

	if len(resources) != 1 || len(missing) != 0 {
		t.Fatalf("expected complete stale snapshot, resources=%#v missing=%#v", resources, missing)
	}
	if !needsRefresh {
		t.Fatal("expected stale snapshot to request background refresh")
	}
}
