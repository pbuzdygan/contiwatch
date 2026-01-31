package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	releaseCheckInterval = 6 * time.Hour
	releaseCheckTimeout  = 10 * time.Second
	defaultReleaseRepo   = "pbuzdygan/contiwatch"
)

type releaseMeta struct {
	Version    string `json:"version"`
	Repo       string `json:"repo"`
	Channel    string `json:"channel"`
	ReleaseTag string `json:"release_tag,omitempty"`
}

type releaseInfo struct {
	Version string `json:"version,omitempty"`
	Tag     string `json:"tag,omitempty"`
	URL     string `json:"url,omitempty"`
}

type releaseCheckState struct {
	Repo            string
	Channel         string
	Latest          releaseInfo
	UpdateAvailable bool
	CheckedAt       time.Time
	Error           string
	ETag            string
}

type releaseCheckResponse struct {
	Meta            releaseMeta `json:"meta"`
	Latest          releaseInfo `json:"latest,omitempty"`
	UpdateAvailable bool        `json:"update_available"`
	CheckedAt       time.Time   `json:"checked_at,omitempty"`
	Error           string      `json:"error,omitempty"`
}

type githubRelease struct {
	TagName         string `json:"tag_name"`
	Name            string `json:"name"`
	TargetCommitish string `json:"target_commitish"`
	HTMLURL         string `json:"html_url"`
	Prerelease      bool   `json:"prerelease"`
	Draft           bool   `json:"draft"`
}

func (s *Server) handleMeta(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, s.releaseMeta())
}

func (s *Server) handleRelease(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	meta := s.releaseMeta()
	if !releaseCheckEnabled(meta) {
		writeJSON(w, http.StatusOK, releaseCheckResponse{
			Meta:            meta,
			UpdateAvailable: false,
			Error:           "release check disabled",
		})
		return
	}
	state := s.getReleaseCheckState()
	if state.CheckedAt.IsZero() || time.Since(state.CheckedAt) > releaseCheckInterval {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), releaseCheckTimeout)
			defer cancel()
			s.refreshReleaseCheck(ctx)
		}()
	}
	resp := releaseCheckResponse{
		Meta:            meta,
		Latest:          state.Latest,
		UpdateAvailable: state.UpdateAvailable,
		CheckedAt:       state.CheckedAt,
		Error:           state.Error,
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) startReleaseCheckMonitor() {
	if s.agentMode {
		return
	}
	meta := s.releaseMeta()
	if !releaseCheckEnabled(meta) {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.releaseCheckStop = cancel
	go s.releaseCheckLoop(ctx)
}

func (s *Server) releaseCheckLoop(ctx context.Context) {
	s.refreshReleaseCheck(ctx)
	ticker := time.NewTicker(releaseCheckInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.refreshReleaseCheck(ctx)
		}
	}
}

func (s *Server) refreshReleaseCheck(ctx context.Context) {
	meta := s.releaseMeta()
	if !releaseCheckEnabled(meta) {
		return
	}
	prev := s.getReleaseCheckState()
	releases, etag, notModified, err := fetchGitHubReleases(ctx, meta.Repo, prev.ETag)
	if err != nil {
		s.setReleaseCheckState(releaseCheckState{
			Repo:            meta.Repo,
			Channel:         meta.Channel,
			Latest:          prev.Latest,
			UpdateAvailable: prev.UpdateAvailable,
			CheckedAt:       time.Now(),
			Error:           err.Error(),
			ETag:            prev.ETag,
		})
		return
	}
	if notModified {
		prev.CheckedAt = time.Now()
		prev.Error = ""
		prev.Repo = meta.Repo
		prev.Channel = meta.Channel
		s.setReleaseCheckState(prev)
		return
	}
	latest := selectReleaseForChannel(releases, meta.Channel)
	latestInfo := releaseInfo{}
	if latest != nil {
		latestInfo.Tag = latest.TagName
		latestInfo.Version = releaseVersion(*latest)
		latestInfo.URL = latest.HTMLURL
	}
	updateAvailable := false
	if latestInfo.Version != "" && meta.Version != "" {
		updateAvailable = compareVersions(meta.Version, latestInfo.Version) < 0
	}
	s.setReleaseCheckState(releaseCheckState{
		Repo:            meta.Repo,
		Channel:         meta.Channel,
		Latest:          latestInfo,
		UpdateAvailable: updateAvailable,
		CheckedAt:       time.Now(),
		Error:           "",
		ETag:            etag,
	})
}

func (s *Server) getReleaseCheckState() releaseCheckState {
	s.releaseCheckMu.RLock()
	defer s.releaseCheckMu.RUnlock()
	return s.releaseCheck
}

func (s *Server) setReleaseCheckState(state releaseCheckState) {
	s.releaseCheckMu.Lock()
	s.releaseCheck = state
	s.releaseCheckMu.Unlock()
}

func releaseCheckEnabled(meta releaseMeta) bool {
	if meta.Repo == "" {
		return false
	}
	if envBool("CONTIWATCH_RELEASE_CHECK", true) == false {
		return false
	}
	if envBool("APP_RELEASE_CHECK", true) == false {
		return false
	}
	return true
}

func (s *Server) releaseMeta() releaseMeta {
	version := strings.TrimSpace(s.version)
	repo := strings.TrimSpace(os.Getenv("CONTIWATCH_REPO"))
	if repo == "" {
		repo = strings.TrimSpace(os.Getenv("APP_REPO"))
	}
	if repo == "" {
		repo = defaultReleaseRepo
	}
	channel := strings.TrimSpace(os.Getenv("CONTIWATCH_CHANNEL"))
	if channel == "" {
		channel = strings.TrimSpace(os.Getenv("APP_CHANNEL"))
	}
	if channel == "" {
		if strings.HasPrefix(strings.ToLower(version), "dev") || strings.HasPrefix(strings.ToLower(version), "vdev") {
			channel = "dev"
		} else {
			channel = "main"
		}
	}
	releaseTag := releaseTagFromVersion(version)
	return releaseMeta{
		Version:    version,
		Repo:       repo,
		Channel:    channel,
		ReleaseTag: releaseTag,
	}
}

func releaseTagFromVersion(version string) string {
	v := strings.TrimSpace(version)
	if v == "" {
		return ""
	}
	lower := strings.ToLower(v)
	if strings.HasPrefix(lower, "vdev") {
		return strings.TrimPrefix(v, "v")
	}
	if strings.HasPrefix(lower, "dev") {
		return v
	}
	if strings.HasPrefix(lower, "v") {
		return v
	}
	return "v" + v
}

func fetchGitHubReleases(ctx context.Context, repo, etag string) ([]githubRelease, string, bool, error) {
	if strings.TrimSpace(repo) == "" {
		return nil, "", false, errors.New("repo not configured")
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases?per_page=30", repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", false, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "contiwatch")
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}
	if token := strings.TrimSpace(os.Getenv("CONTIWATCH_GITHUB_TOKEN")); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	} else if token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	client := &http.Client{Timeout: releaseCheckTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotModified {
		return nil, resp.Header.Get("ETag"), true, nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", false, fmt.Errorf("github releases status %d", resp.StatusCode)
	}
	var releases []githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, "", false, err
	}
	return releases, resp.Header.Get("ETag"), false, nil
}

func isDevRelease(rel githubRelease) bool {
	tag := strings.ToLower(strings.TrimSpace(rel.TagName))
	name := strings.ToLower(strings.TrimSpace(rel.Name))
	branch := strings.ToLower(strings.TrimSpace(rel.TargetCommitish))
	return strings.HasPrefix(tag, "dev") || strings.HasPrefix(name, "dev") || branch == "dev"
}

func selectReleaseForChannel(releases []githubRelease, channel string) *githubRelease {
	if len(releases) == 0 {
		return nil
	}
	normalized := strings.ToLower(strings.TrimSpace(channel))
	filtered := make([]githubRelease, 0, len(releases))
	for _, rel := range releases {
		if rel.Draft {
			continue
		}
		isDev := isDevRelease(rel)
		if normalized == "dev" {
			if isDev {
				filtered = append(filtered, rel)
			}
		} else {
			if !isDev {
				filtered = append(filtered, rel)
			}
		}
	}
	if len(filtered) == 0 {
		filtered = releases
	}
	sort.Slice(filtered, func(i, j int) bool {
		return compareVersions(releaseVersion(filtered[j]), releaseVersion(filtered[i])) < 0
	})
	latest := filtered[0]
	return &latest
}

func releaseVersion(rel githubRelease) string {
	if rel.TagName != "" {
		return rel.TagName
	}
	if rel.Name != "" {
		return rel.Name
	}
	return ""
}

func compareVersions(a, b string) int {
	left := normalizeVersion(a)
	right := normalizeVersion(b)
	if left == "" && right == "" {
		return 0
	}
	if left == "" {
		return -1
	}
	if right == "" {
		return 1
	}
	leftDev := isDevVersion(left)
	rightDev := isDevVersion(right)
	if leftDev || rightDev {
		if leftDev && !rightDev {
			return -1
		}
		if !leftDev && rightDev {
			return 1
		}
		return compareDevVersions(left, right)
	}
	leftParts, leftOk := parseSemverParts(left)
	rightParts, rightOk := parseSemverParts(right)
	if leftOk && rightOk {
		return compareSemver(leftParts, rightParts)
	}
	return strings.Compare(left, right)
}

func normalizeVersion(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(v), "v") {
		return strings.TrimPrefix(v, "v")
	}
	return v
}

func isDevVersion(v string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(v)), "dev")
}

func compareDevVersions(a, b string) int {
	left := devSuffix(a)
	right := devSuffix(b)
	leftParts, leftOk := parseSemverParts(left)
	rightParts, rightOk := parseSemverParts(right)
	if leftOk && rightOk {
		return compareSemver(leftParts, rightParts)
	}
	if li, ok := parseInt(left); ok {
		if ri, ok := parseInt(right); ok {
			switch {
			case li > ri:
				return 1
			case li < ri:
				return -1
			default:
				return 0
			}
		}
	}
	return strings.Compare(left, right)
}

func devSuffix(v string) string {
	trimmed := strings.TrimSpace(v)
	lower := strings.ToLower(trimmed)
	if strings.HasPrefix(lower, "vdev") {
		trimmed = trimmed[1:]
		lower = lower[1:]
	}
	trimmed = trimmed[len("dev"):]
	trimmed = strings.TrimPrefix(trimmed, "-")
	trimmed = strings.TrimPrefix(trimmed, "_")
	return strings.TrimSpace(trimmed)
}

func parseSemverParts(v string) ([]int, bool) {
	if v == "" {
		return nil, false
	}
	parts := strings.Split(v, ".")
	if len(parts) == 0 {
		return nil, false
	}
	out := make([]int, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			out = append(out, 0)
			continue
		}
		n, err := strconv.Atoi(part)
		if err != nil {
			return nil, false
		}
		out = append(out, n)
	}
	return out, true
}

func compareSemver(a, b []int) int {
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}
	for i := 0; i < maxLen; i++ {
		left := 0
		right := 0
		if i < len(a) {
			left = a[i]
		}
		if i < len(b) {
			right = b[i]
		}
		if left > right {
			return 1
		}
		if left < right {
			return -1
		}
	}
	return 0
}

func parseInt(v string) (int, bool) {
	if v == "" {
		return 0, false
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, false
	}
	return n, true
}

func envBool(key string, fallback bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	switch strings.ToLower(raw) {
	case "1", "true", "yes", "on", "enabled":
		return true
	case "0", "false", "no", "off", "disabled":
		return false
	default:
		return fallback
	}
}
