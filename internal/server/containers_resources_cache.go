package server

import (
	"time"

	"contiwatch/internal/dockerwatcher"
)

type containersResourcesCacheEntry struct {
	Resource  dockerwatcher.ContainerResource
	FetchedAt time.Time
}

func (s *Server) getContainersResourcesCache(scope string, ids []string, maxAge time.Duration, refreshAfter time.Duration) (resources []dockerwatcher.ContainerResource, needsRefresh bool) {
	if scope == "" || len(ids) == 0 {
		return nil, true
	}
	now := time.Now()
	s.containersResourcesCacheMu.RLock()
	defer s.containersResourcesCacheMu.RUnlock()
	byScope := s.containersResourcesCache[scope]
	if len(byScope) == 0 {
		return nil, true
	}
	out := make([]dockerwatcher.ContainerResource, 0, len(ids))
	for _, id := range ids {
		entry, ok := byScope[id]
		if !ok {
			needsRefresh = true
			continue
		}
		age := now.Sub(entry.FetchedAt)
		if maxAge > 0 && age > maxAge {
			needsRefresh = true
			continue
		}
		if refreshAfter > 0 && age > refreshAfter {
			needsRefresh = true
		}
		out = append(out, entry.Resource)
	}
	if len(out) == 0 {
		return nil, true
	}
	return out, needsRefresh
}

func (s *Server) setContainersResourcesCache(scope string, resources []dockerwatcher.ContainerResource) {
	if scope == "" || len(resources) == 0 {
		return
	}
	now := time.Now()
	s.containersResourcesCacheMu.Lock()
	defer s.containersResourcesCacheMu.Unlock()
	if s.containersResourcesCache == nil {
		s.containersResourcesCache = map[string]map[string]containersResourcesCacheEntry{}
	}
	byScope, ok := s.containersResourcesCache[scope]
	if !ok || byScope == nil {
		byScope = map[string]containersResourcesCacheEntry{}
		s.containersResourcesCache[scope] = byScope
	}
	for _, res := range resources {
		if res.ID == "" {
			continue
		}
		byScope[res.ID] = containersResourcesCacheEntry{Resource: res, FetchedAt: now}
	}
}

func (s *Server) shouldRefreshContainersResources(scope string, minInterval time.Duration) bool {
	if scope == "" {
		return false
	}
	now := time.Now()
	s.containersResourcesCacheMu.Lock()
	defer s.containersResourcesCacheMu.Unlock()
	if s.containersResourcesRefresh == nil {
		s.containersResourcesRefresh = map[string]time.Time{}
	}
	last := s.containersResourcesRefresh[scope]
	if !last.IsZero() && minInterval > 0 && now.Sub(last) < minInterval {
		return false
	}
	s.containersResourcesRefresh[scope] = now
	return true
}
