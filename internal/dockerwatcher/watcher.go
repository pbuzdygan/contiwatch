package dockerwatcher

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"contiwatch/internal/config"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

const policyLabel = "contiwatch.policy"

var ErrSkipContainer = errors.New("container skipped")

// ContainerStatus describes the last known state for a container.
type ContainerStatus struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Image           string    `json:"image"`
	ImageID         string    `json:"image_id"`
	NewImageID      string    `json:"new_image_id,omitempty"`
	Running         bool      `json:"running"`
	Paused          bool      `json:"paused"`
	Policy          string    `json:"policy"`
	UpdateAvailable bool      `json:"update_available"`
	Updated         bool      `json:"updated"`
	LastChecked     time.Time `json:"last_checked"`
	Error           string    `json:"error,omitempty"`
}

type ContainerInfo struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Image       string                 `json:"image"`
	State       string                 `json:"state"`
	UptimeSec   int64                  `json:"uptime_sec"`
	Stack       string                 `json:"stack"`
	IPAddresses []ContainerIP          `json:"ip_addresses,omitempty"`
	Ports       []ContainerPortBinding `json:"ports,omitempty"`
}

type ContainerIP struct {
	Network string `json:"network"`
	IP      string `json:"ip"`
}

type ContainerPortBinding struct {
	Proto         string `json:"proto"`
	ContainerPort int    `json:"container_port"`
	HostIP        string `json:"host_ip"`
	HostPort      int    `json:"host_port"`
}

type ContainerResource struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Image         string                 `json:"image,omitempty"`
	State         string                 `json:"state"`
	UptimeSec     int64                  `json:"uptime_sec"`
	CPUPercent    float64                `json:"cpu_percent"`
	MemUsageBytes uint64                 `json:"mem_usage_bytes"`
	MemLimitBytes uint64                 `json:"mem_limit_bytes"`
	MemPercent    float64                `json:"mem_percent"`
	IPAddresses   []ContainerIP          `json:"ip_addresses,omitempty"`
	Ports         []ContainerPortBinding `json:"ports,omitempty"`
}

type ImageInfo struct {
	ID              string `json:"id"`
	Repository      string `json:"repository"`
	Tag             string `json:"tag"`
	SizeBytes       int64  `json:"size_bytes"`
	CreatedAt       string `json:"created_at"`
	ContainersCount int64  `json:"containers_count"`
	Dangling        bool   `json:"dangling"`
}

type UpdateResult struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Image          string `json:"image"`
	OldImageID     string `json:"old_image_id,omitempty"`
	NewImageID     string `json:"new_image_id,omitempty"`
	AppliedImageID string `json:"applied_image_id,omitempty"`
	Updated        bool   `json:"updated"`
	PreviousState  string `json:"previous_state"`
	CurrentState   string `json:"current_state"`
	Message        string `json:"message,omitempty"`
}

// ScanResult describes a full scan for a single server.
type ScanResult struct {
	ServerName     string            `json:"server_name"`
	ServerURL      string            `json:"server_url"`
	Local          bool              `json:"local"`
	CheckedAt      time.Time         `json:"checked_at"`
	Containers     []ContainerStatus `json:"containers"`
	Error          string            `json:"error,omitempty"`
	ScanState      string            `json:"scan_state,omitempty"`
	ScanStartedAt  *time.Time        `json:"scan_started_at,omitempty"`
	ScanFinishedAt *time.Time        `json:"scan_finished_at,omitempty"`
}

// Watcher performs scans against a local Docker daemon.
type Watcher struct {
	client *client.Client
}

func New() (*Watcher, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &Watcher{client: cli}, nil
}

func NewWithHost(host string) (*Watcher, error) {
	opts := []client.Opt{client.WithAPIVersionNegotiation()}
	if strings.TrimSpace(host) != "" {
		opts = append(opts, client.WithHost(host))
	}
	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, err
	}
	return &Watcher{client: cli}, nil
}

func (w *Watcher) Close() error {
	return w.client.Close()
}

func (w *Watcher) Scan(ctx context.Context, cfg config.Config) (ScanResult, error) {
	return w.scan(ctx, cfg, false)
}

func (w *Watcher) ScanNoUpdate(ctx context.Context, cfg config.Config) (ScanResult, error) {
	return w.scan(ctx, cfg, true)
}

func (w *Watcher) scan(ctx context.Context, cfg config.Config, scanOnly bool) (ScanResult, error) {
	containers, err := w.client.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return ScanResult{}, err
	}
	result := ScanResult{
		ServerName: "local",
		ServerURL:  "local",
		Local:      true,
		CheckedAt:  time.Now(),
	}

	for _, item := range containers {
		if ctx.Err() != nil {
			result.Error = "scan cancelled manually"
			return result, nil
		}
		status := w.scanContainer(ctx, item, cfg, scanOnly)
		result.Containers = append(result.Containers, status)
		if ctx.Err() != nil {
			result.Error = "scan cancelled manually"
			return result, nil
		}
	}

	return result, nil
}

func (w *Watcher) Containers(ctx context.Context) ([]ContainerInfo, error) {
	items, err := w.client.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return nil, err
	}
	containers := make([]ContainerInfo, 0, len(items))
	for _, item := range items {
		if ctx.Err() != nil {
			return containers, ctx.Err()
		}
		info, err := w.containerInfo(ctx, item)
		if err != nil {
			return containers, err
		}
		containers = append(containers, info)
	}
	return containers, nil
}

func (w *Watcher) ContainerResources(ctx context.Context, ids []string) ([]ContainerResource, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	type result struct {
		index int
		value ContainerResource
	}
	trimmed := make([]string, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		trimmed = append(trimmed, id)
	}
	if len(trimmed) == 0 {
		return []ContainerResource{}, nil
	}

	results := make([]result, 0, len(trimmed))
	var resultsMu sync.Mutex
	var firstErr error
	var errMu sync.Mutex

	workers := 6
	if len(trimmed) < workers {
		workers = len(trimmed)
	}
	jobs := make(chan int, len(trimmed))
	var wg sync.WaitGroup
	workerFn := func() {
		defer wg.Done()
		for idx := range jobs {
			if ctx.Err() != nil {
				return
			}
			id := trimmed[idx]
			resource, err := w.containerResource(ctx, id)
			if err != nil {
				errMu.Lock()
				if firstErr == nil {
					firstErr = err
					cancel()
				}
				errMu.Unlock()
				return
			}
			resultsMu.Lock()
			results = append(results, result{index: idx, value: resource})
			resultsMu.Unlock()
		}
	}
	wg.Add(workers)
	for i := 0; i < workers; i += 1 {
		go workerFn()
	}
	for i := 0; i < len(trimmed); i += 1 {
		jobs <- i
	}
	close(jobs)
	wg.Wait()

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if firstErr != nil {
		return nil, firstErr
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].index < results[j].index
	})
	ordered := make([]ContainerResource, 0, len(results))
	for _, item := range results {
		ordered = append(ordered, item.value)
	}
	return ordered, nil
}

func (w *Watcher) containerResource(ctx context.Context, id string) (ContainerResource, error) {
	inspect, err := w.client.ContainerInspect(ctx, id)
	if err != nil {
		return ContainerResource{}, err
	}
	name := strings.TrimPrefix(inspect.Name, "/")
	if name == "" {
		name = id
		if len(name) > 12 {
			name = name[:12]
		}
	}
	state := "unknown"
	uptimeSec := int64(0)
	running := false
	if inspect.State != nil {
		state = normalizeContainerState(inspect.State)
		uptimeSec = uptimeFromState(inspect.State)
		running = inspect.State.Running
	}
	imageRef := ""
	if inspect.Config != nil && inspect.Config.Image != "" {
		imageRef = inspect.Config.Image
	}
	ipAddresses := extractContainerIPs(inspect.NetworkSettings)
	ports := extractContainerPorts(inspect.NetworkSettings)
	stats := types.StatsJSON{}
	if running {
		raw, err := w.client.ContainerStats(ctx, id, false)
		if err != nil {
			return ContainerResource{}, err
		}
		if err := json.NewDecoder(raw.Body).Decode(&stats); err != nil {
			raw.Body.Close()
			return ContainerResource{}, err
		}
		raw.Body.Close()
	}
	cpuPercent := computeContainerCPUPercent(stats)
	memUsage := stats.MemoryStats.Usage
	memLimit := stats.MemoryStats.Limit
	memPercent := computeContainerMemPercent(memUsage, memLimit)
	return ContainerResource{
		ID:            id,
		Name:          name,
		Image:         imageRef,
		State:         state,
		UptimeSec:     uptimeSec,
		CPUPercent:    cpuPercent,
		MemUsageBytes: memUsage,
		MemLimitBytes: memLimit,
		MemPercent:    memPercent,
		IPAddresses:   ipAddresses,
		Ports:         ports,
	}, nil
}

func (w *Watcher) containerInfo(ctx context.Context, item types.Container) (ContainerInfo, error) {
	name := strings.TrimPrefix(firstOrEmpty(item.Names), "/")
	inspect, err := w.client.ContainerInspect(ctx, item.ID)
	if err != nil {
		return ContainerInfo{}, err
	}
	imageRef := item.Image
	if inspect.Config != nil && inspect.Config.Image != "" {
		imageRef = inspect.Config.Image
	}
	state := "unknown"
	uptimeSec := int64(0)
	if inspect.State != nil {
		state = normalizeContainerState(inspect.State)
		uptimeSec = uptimeFromState(inspect.State)
	}
	stack := stackFromLabels(nil)
	if inspect.Config != nil {
		stack = stackFromLabels(inspect.Config.Labels)
	}
	ipAddresses := extractContainerIPs(inspect.NetworkSettings)
	ports := extractContainerPorts(inspect.NetworkSettings)
	return ContainerInfo{
		ID:          item.ID,
		Name:        name,
		Image:       imageRef,
		State:       state,
		UptimeSec:   uptimeSec,
		Stack:       stack,
		IPAddresses: ipAddresses,
		Ports:       ports,
	}, nil
}

func extractContainerIPs(settings *types.NetworkSettings) []ContainerIP {
	if settings == nil || settings.Networks == nil {
		return nil
	}
	ips := make([]ContainerIP, 0, len(settings.Networks))
	for name, network := range settings.Networks {
		if network == nil || network.IPAddress == "" {
			continue
		}
		ips = append(ips, ContainerIP{Network: name, IP: network.IPAddress})
	}
	sort.Slice(ips, func(i, j int) bool {
		return ips[i].Network < ips[j].Network
	})
	return ips
}

func extractContainerPorts(settings *types.NetworkSettings) []ContainerPortBinding {
	if settings == nil || settings.Ports == nil {
		return nil
	}
	ports := make([]ContainerPortBinding, 0, len(settings.Ports))
	for port, bindings := range settings.Ports {
		containerPort := port.Int()
		proto := port.Proto()
		if len(bindings) == 0 {
			ports = append(ports, ContainerPortBinding{
				Proto:         proto,
				ContainerPort: containerPort,
				HostIP:        "",
				HostPort:      0,
			})
			continue
		}
		for _, binding := range bindings {
			if strings.Contains(binding.HostIP, ":") {
				continue
			}
			hostPort := 0
			if binding.HostPort != "" {
				if parsed, err := strconv.Atoi(binding.HostPort); err == nil {
					hostPort = parsed
				}
			}
			ports = append(ports, ContainerPortBinding{
				Proto:         proto,
				ContainerPort: containerPort,
				HostIP:        binding.HostIP,
				HostPort:      hostPort,
			})
		}
	}
	sort.Slice(ports, func(i, j int) bool {
		if ports[i].ContainerPort == ports[j].ContainerPort {
			if ports[i].Proto == ports[j].Proto {
				return ports[i].HostPort < ports[j].HostPort
			}
			return ports[i].Proto < ports[j].Proto
		}
		return ports[i].ContainerPort < ports[j].ContainerPort
	})
	return ports
}

func computeContainerCPUPercent(stats types.StatsJSON) float64 {
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage)
	if cpuDelta <= 0 || systemDelta <= 0 {
		return 0
	}
	percent := (cpuDelta / systemDelta) * 100
	if percent < 0 {
		return 0
	}
	if percent > 100 {
		return 100
	}
	return percent
}

func computeContainerMemPercent(usage, limit uint64) float64 {
	if limit == 0 {
		return 0
	}
	percent := (float64(usage) / float64(limit)) * 100
	if percent < 0 {
		return 0
	}
	if percent > 100 {
		return 100
	}
	return percent
}

func normalizeContainerState(state *types.ContainerState) string {
	if state == nil {
		return "unknown"
	}
	if state.Paused {
		return "paused"
	}
	if state.Restarting {
		return "restarting"
	}
	if state.OOMKilled {
		return "oomkilled"
	}
	if state.Dead {
		return "dead"
	}
	if state.Status != "" {
		return state.Status
	}
	return "unknown"
}

func uptimeFromState(state *types.ContainerState) int64 {
	if state == nil {
		return 0
	}
	if !(state.Running || state.Paused || state.Restarting) {
		return 0
	}
	if state.StartedAt == "" {
		return 0
	}
	started, err := time.Parse(time.RFC3339Nano, state.StartedAt)
	if err != nil {
		return 0
	}
	seconds := int64(time.Since(started).Seconds())
	if seconds < 0 {
		return 0
	}
	return seconds
}

func stackFromLabels(labels map[string]string) string {
	if labels == nil {
		return "-"
	}
	if value := strings.TrimSpace(labels["com.docker.stack.namespace"]); value != "" {
		return value
	}
	if value := strings.TrimSpace(labels["com.docker.compose.project"]); value != "" {
		return value
	}
	if value := strings.TrimSpace(labels["io.portainer.stack.name"]); value != "" {
		return value
	}
	return "-"
}

func (w *Watcher) scanContainer(ctx context.Context, item types.Container, cfg config.Config, scanOnly bool) ContainerStatus {
	status := ContainerStatus{
		ID:          item.ID,
		Name:        strings.TrimPrefix(firstOrEmpty(item.Names), "/"),
		Image:       item.Image,
		ImageID:     item.ImageID,
		Policy:      cfg.GlobalPolicy,
		LastChecked: time.Now(),
	}

	inspect, err := w.client.ContainerInspect(ctx, item.ID)
	if err != nil {
		status.Error = err.Error()
		return status
	}
	if inspect.State != nil {
		status.Running = inspect.State.Running
		status.Paused = inspect.State.Paused
	}
	if inspect.Config != nil && inspect.Config.Labels != nil {
		if override, ok := inspect.Config.Labels[policyLabel]; ok {
			status.Policy = strings.ToLower(strings.TrimSpace(override))
		}
	}

	if status.Policy == config.PolicySkip {
		status.Error = "skipped: policy"
		return status
	}

	imageRef := ""
	if inspect.Config != nil {
		imageRef = inspect.Config.Image
	}
	if imageRef == "" {
		status.Error = "missing image reference"
		return status
	}

	if digest := digestFromRef(imageRef); digest != "" {
		status.ImageID = digest
		status.NewImageID = digest
		status.UpdateAvailable = false
		return status
	}

	localDigest := w.localImageDigest(ctx, inspect.Image, imageRef)
	if localDigest != "" {
		status.ImageID = localDigest
	}
	remoteDigest, err := w.remoteImageDigest(ctx, imageRef)
	if err != nil {
		status.Error = fmt.Sprintf("skipped: digest unknown (registry error: %v)", err)
		return status
	}
	if localDigest == "" {
		status.Error = "skipped: digest unknown (local digest unavailable)"
		return status
	}
	if remoteDigest == "" {
		status.Error = "skipped: digest unknown (registry digest unavailable)"
		return status
	}
	status.NewImageID = remoteDigest
	status.UpdateAvailable = remoteDigest != "" && localDigest != "" && remoteDigest != localDigest
	if !status.UpdateAvailable {
		return status
	}

	if status.Policy == config.PolicyUpdate {
		if scanOnly {
			return status
		}
		if inspect.State != nil && inspect.State.Paused {
			status.Error = "skipped: container paused"
			return status
		}

		wasRunning := inspect.State != nil && inspect.State.Running
		if !wasRunning && !cfg.UpdateStoppedContainers {
			status.Error = "skipped: container not running"
			return status
		}

		newContainerID, err := w.recreateContainer(ctx, inspect, imageRef, wasRunning)
		if err != nil {
			status.Error = err.Error()
			return status
		}
		if newContainerID != "" && status.NewImageID != "" {
			if newInspect, err := w.client.ContainerInspect(ctx, newContainerID); err == nil {
				if newInspect.Image != "" && newInspect.Image != status.NewImageID {
					// Retry once using the resolved image ID to avoid tag-resolution edge cases.
					if retryID, retryErr := w.recreateContainer(ctx, newInspect, status.NewImageID, wasRunning); retryErr == nil {
						if retryInspect, err := w.client.ContainerInspect(ctx, retryID); err == nil && retryInspect.Image == status.NewImageID {
							newContainerID = retryID
						} else if retryErr == nil {
							status.Error = fmt.Sprintf("update applied unexpected image: expected %s got %s", status.NewImageID, retryInspect.Image)
							return status
						}
					} else {
						status.Error = fmt.Sprintf("update applied unexpected image (expected %s got %s); retry failed: %v", status.NewImageID, newInspect.Image, retryErr)
						return status
					}
				}
			}
		}
		status.Updated = true
	}

	return status
}

func (w *Watcher) pullImage(ctx context.Context, imageRef string) error {
	reader, err := w.client.ImagePull(ctx, imageRef, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()
	_, _ = io.Copy(io.Discard, reader)
	return nil
}

func (w *Watcher) PullImage(ctx context.Context, repository, tag string) error {
	repo := strings.TrimSpace(repository)
	if repo == "" {
		return errors.New("repository is required")
	}
	tag = strings.TrimSpace(tag)
	if tag == "" {
		tag = "latest"
	}
	return w.pullImage(ctx, fmt.Sprintf("%s:%s", repo, tag))
}

func (w *Watcher) ListImages(ctx context.Context) ([]ImageInfo, error) {
	images, err := w.client.ImageList(ctx, types.ImageListOptions{All: true})
	if err != nil {
		return nil, err
	}
	containers, err := w.client.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return nil, err
	}
	usage := map[string]int64{}
	for _, container := range containers {
		if container.ImageID == "" {
			continue
		}
		usage[container.ImageID] += 1
	}
	result := make([]ImageInfo, 0)
	for _, image := range images {
		createdAt := time.Unix(image.Created, 0).Format(time.RFC3339)
		containersCount := usage[image.ID]
		if len(image.RepoTags) == 0 {
			result = append(result, ImageInfo{
				ID:              image.ID,
				Repository:      "<none>",
				Tag:             "<none>",
				SizeBytes:       image.Size,
				CreatedAt:       createdAt,
				ContainersCount: containersCount,
				Dangling:        true,
			})
			continue
		}
		for _, ref := range image.RepoTags {
			repo, tag := splitRepoTag(ref)
			dangling := repo == "<none>" || tag == "<none>"
			result = append(result, ImageInfo{
				ID:              image.ID,
				Repository:      repo,
				Tag:             tag,
				SizeBytes:       image.Size,
				CreatedAt:       createdAt,
				ContainersCount: containersCount,
				Dangling:        dangling,
			})
		}
	}
	return result, nil
}

func (w *Watcher) RemoveImage(ctx context.Context, imageID string) error {
	imageID = strings.TrimSpace(imageID)
	if imageID == "" {
		return errors.New("image id is required")
	}
	_, err := w.client.ImageRemove(ctx, imageID, types.ImageRemoveOptions{Force: false, PruneChildren: false})
	return err
}

func (w *Watcher) PruneImages(ctx context.Context, mode string) (int, uint64, error) {
	args := filters.NewArgs()
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "dangling":
		args.Add("dangling", "true")
	case "unused":
		args.Add("dangling", "false")
	default:
		return 0, 0, errors.New("invalid prune mode")
	}
	report, err := w.client.ImagesPrune(ctx, args)
	if err != nil {
		return 0, 0, err
	}
	deleted := 0
	if report.ImagesDeleted != nil {
		deleted = len(report.ImagesDeleted)
	}
	return deleted, report.SpaceReclaimed, nil
}

func (w *Watcher) remoteImageDigest(ctx context.Context, imageRef string) (string, error) {
	inspect, err := w.client.DistributionInspect(ctx, imageRef, "")
	if err != nil {
		return "", err
	}
	digest := inspect.Descriptor.Digest.String()
	return digest, nil
}

func (w *Watcher) localImageDigest(ctx context.Context, imageID, imageRef string) string {
	inspect, _, err := w.client.ImageInspectWithRaw(ctx, imageID)
	if err != nil {
		return ""
	}
	repo := imageRepo(imageRef)
	for _, digestRef := range inspect.RepoDigests {
		if repo == "" || strings.HasPrefix(digestRef, repo+"@") {
			if digest := digestFromRef(digestRef); digest != "" {
				return digest
			}
		}
	}
	if len(inspect.RepoDigests) > 0 {
		return digestFromRef(inspect.RepoDigests[0])
	}
	return ""
}

func imageRepo(imageRef string) string {
	if at := strings.Index(imageRef, "@"); at != -1 {
		return imageRef[:at]
	}
	lastSlash := strings.LastIndex(imageRef, "/")
	lastColon := strings.LastIndex(imageRef, ":")
	if lastColon > lastSlash {
		return imageRef[:lastColon]
	}
	return imageRef
}

func splitRepoTag(imageRef string) (string, string) {
	ref := strings.TrimSpace(imageRef)
	if ref == "" {
		return "<none>", "<none>"
	}
	if ref == "<none>:<none>" {
		return "<none>", "<none>"
	}
	lastSlash := strings.LastIndex(ref, "/")
	lastColon := strings.LastIndex(ref, ":")
	if lastColon > lastSlash {
		return ref[:lastColon], ref[lastColon+1:]
	}
	return ref, "<none>"
}

func digestFromRef(ref string) string {
	if at := strings.Index(ref, "@"); at != -1 && at+1 < len(ref) {
		return ref[at+1:]
	}
	return ""
}

func (w *Watcher) recreateContainer(ctx context.Context, inspect types.ContainerJSON, imageRef string, startAfter bool) (string, error) {
	name := strings.TrimPrefix(inspect.Name, "/")
	if name == "" {
		name = inspect.ID[:12]
	}

	config := inspect.Config
	if config == nil {
		return "", errors.New("missing container config")
	}
	config.Image = imageRef

	hostConfig := inspect.HostConfig
	if hostConfig == nil {
		return "", errors.New("missing host config")
	}

	networking := &network.NetworkingConfig{}
	if inspect.NetworkSettings != nil && inspect.NetworkSettings.Networks != nil {
		endpoints := map[string]*network.EndpointSettings{}
		for netName, netCfg := range inspect.NetworkSettings.Networks {
			endpoints[netName] = &network.EndpointSettings{
				Aliases:    netCfg.Aliases,
				Links:      netCfg.Links,
				IPAMConfig: netCfg.IPAMConfig,
			}
		}
		networking.EndpointsConfig = endpoints
	}

	if startAfter {
		if err := w.client.ContainerStop(ctx, inspect.ID, container.StopOptions{}); err != nil {
			return "", fmt.Errorf("stop: %w", err)
		}
	}
	if err := w.client.ContainerRemove(ctx, inspect.ID, types.ContainerRemoveOptions{Force: true}); err != nil {
		return "", fmt.Errorf("remove: %w", err)
	}

	created, err := w.client.ContainerCreate(ctx, config, hostConfig, networking, nil, name)
	if err != nil {
		return "", fmt.Errorf("create: %w", err)
	}

	if startAfter {
		if err := w.client.ContainerStart(ctx, created.ID, types.ContainerStartOptions{}); err != nil {
			return "", fmt.Errorf("start: %w", err)
		}
	}

	return created.ID, nil
}

func firstOrEmpty(items []string) string {
	if len(items) == 0 {
		return ""
	}
	return items[0]
}

func (w *Watcher) UpdateContainer(ctx context.Context, containerID string, cfg config.Config) (UpdateResult, error) {
	return w.updateContainer(ctx, containerID, cfg, false)
}

func (w *Watcher) UpdateContainerForceStart(ctx context.Context, containerID string, cfg config.Config) (UpdateResult, error) {
	return w.updateContainer(ctx, containerID, cfg, true)
}

func (w *Watcher) StartContainer(ctx context.Context, containerID string) error {
	return w.client.ContainerStart(ctx, strings.TrimSpace(containerID), types.ContainerStartOptions{})
}

func (w *Watcher) StopContainer(ctx context.Context, containerID string, timeoutSec int) error {
	timeout := timeoutSec
	return w.client.ContainerStop(ctx, strings.TrimSpace(containerID), container.StopOptions{Timeout: &timeout})
}

func (w *Watcher) RestartContainer(ctx context.Context, containerID string, timeoutSec int) error {
	timeout := timeoutSec
	return w.client.ContainerRestart(ctx, strings.TrimSpace(containerID), container.StopOptions{Timeout: &timeout})
}

func (w *Watcher) PauseContainer(ctx context.Context, containerID string) error {
	return w.client.ContainerPause(ctx, strings.TrimSpace(containerID))
}

func (w *Watcher) UnpauseContainer(ctx context.Context, containerID string) error {
	return w.client.ContainerUnpause(ctx, strings.TrimSpace(containerID))
}

func (w *Watcher) KillContainer(ctx context.Context, containerID string) error {
	return w.client.ContainerKill(ctx, strings.TrimSpace(containerID), "SIGKILL")
}

func (w *Watcher) ExecShell(ctx context.Context, containerID string, cmd []string) (string, types.HijackedResponse, error) {
	containerID = strings.TrimSpace(containerID)
	if containerID == "" {
		return "", types.HijackedResponse{}, errors.New("container id is required")
	}
	if len(cmd) == 0 {
		cmd = []string{"/bin/sh"}
	}
	execResp, err := w.client.ContainerExecCreate(ctx, containerID, types.ExecConfig{
		Tty:          true,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          cmd,
	})
	if err != nil {
		return "", types.HijackedResponse{}, err
	}
	attach, err := w.client.ContainerExecAttach(ctx, execResp.ID, types.ExecStartCheck{Tty: true})
	if err != nil {
		return "", types.HijackedResponse{}, err
	}
	return execResp.ID, attach, nil
}

func (w *Watcher) ResizeExec(ctx context.Context, execID string, cols, rows uint) error {
	execID = strings.TrimSpace(execID)
	if execID == "" {
		return errors.New("exec id is required")
	}
	return w.client.ContainerExecResize(ctx, execID, types.ResizeOptions{
		Height: rows,
		Width:  cols,
	})
}

func (w *Watcher) InspectExec(ctx context.Context, execID string) (types.ContainerExecInspect, error) {
	execID = strings.TrimSpace(execID)
	if execID == "" {
		return types.ContainerExecInspect{}, errors.New("exec id is required")
	}
	return w.client.ContainerExecInspect(ctx, execID)
}

func (w *Watcher) ContainerLogs(ctx context.Context, containerID string, opts types.ContainerLogsOptions) (io.ReadCloser, bool, error) {
	containerID = strings.TrimSpace(containerID)
	if containerID == "" {
		return nil, false, errors.New("container id is required")
	}
	inspect, err := w.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, false, err
	}
	reader, err := w.client.ContainerLogs(ctx, containerID, opts)
	if err != nil {
		return nil, false, err
	}
	tty := false
	if inspect.Config != nil {
		tty = inspect.Config.Tty
	}
	return reader, tty, nil
}

func (w *Watcher) updateContainer(ctx context.Context, containerID string, cfg config.Config, forceStart bool) (UpdateResult, error) {
	containerID = strings.TrimSpace(containerID)
	if containerID == "" {
		return UpdateResult{}, errors.New("container id is required")
	}

	inspect, err := w.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return UpdateResult{}, err
	}

	name := strings.TrimPrefix(inspect.Name, "/")
	if name == "" {
		name = containerID
	}

	previousState := "unknown"
	if inspect.State != nil {
		if inspect.State.Paused {
			previousState = "paused"
		} else if inspect.State.Running {
			previousState = "online"
		} else {
			previousState = "stopped"
		}
	}

	if inspect.State != nil && inspect.State.Paused {
		return UpdateResult{
			ID:            containerID,
			Name:          name,
			Image:         inspect.Config.Image,
			Updated:       false,
			PreviousState: previousState,
			CurrentState:  previousState,
			Message:       "skipped: container paused",
		}, nil
	}

	imageRef := ""
	if inspect.Config != nil {
		imageRef = inspect.Config.Image
	}
	if imageRef == "" {
		return UpdateResult{}, errors.New("missing image reference")
	}

	if err := w.pullImage(ctx, imageRef); err != nil {
		return UpdateResult{}, err
	}

	newImage, _, err := w.client.ImageInspectWithRaw(ctx, imageRef)
	if err != nil {
		return UpdateResult{}, err
	}

	oldImageID := inspect.Image
	newImageID := newImage.ID
	updateAvailable := newImage.ID != "" && newImage.ID != inspect.Image
	if !updateAvailable {
		return UpdateResult{
			ID:             containerID,
			Name:           name,
			Image:          imageRef,
			OldImageID:     oldImageID,
			NewImageID:     newImageID,
			AppliedImageID: oldImageID,
			Updated:        false,
			PreviousState:  previousState,
			CurrentState:   previousState,
			Message:        "no update available",
		}, nil
	}

	wasRunning := inspect.State != nil && inspect.State.Running
	startAfter := wasRunning || forceStart
	if !wasRunning && !forceStart && !cfg.UpdateStoppedContainers {
		return UpdateResult{
			ID:             containerID,
			Name:           name,
			Image:          imageRef,
			OldImageID:     oldImageID,
			NewImageID:     newImageID,
			AppliedImageID: oldImageID,
			Updated:        false,
			PreviousState:  previousState,
			CurrentState:   previousState,
			Message:        "skipped: container not running",
		}, nil
	}

	newContainerID, err := w.recreateContainer(ctx, inspect, imageRef, startAfter)
	if err != nil {
		return UpdateResult{}, err
	}

	appliedImageID := ""
	if newContainerID != "" {
		if newInspect, err := w.client.ContainerInspect(ctx, newContainerID); err == nil {
			appliedImageID = newInspect.Image
			if newImageID != "" && appliedImageID != "" && appliedImageID != newImageID {
				// Retry once using the resolved image ID to avoid tag-resolution edge cases.
				retryID, retryErr := w.recreateContainer(ctx, newInspect, newImageID, startAfter)
				if retryErr != nil {
					return UpdateResult{}, fmt.Errorf("updated unexpected image (expected %s got %s); retry failed: %w", newImageID, appliedImageID, retryErr)
				}
				retryInspect, err := w.client.ContainerInspect(ctx, retryID)
				if err != nil {
					return UpdateResult{}, fmt.Errorf("updated unexpected image (expected %s got %s); retry inspect failed: %w", newImageID, appliedImageID, err)
				}
				if retryInspect.Image != newImageID {
					return UpdateResult{}, fmt.Errorf("updated unexpected image (expected %s got %s); retry still on %s", newImageID, appliedImageID, retryInspect.Image)
				}
				newContainerID = retryID
				appliedImageID = retryInspect.Image
			}
		}
	}

	currentState := "stopped"
	if startAfter {
		currentState = "online"
	}

	if newContainerID != "" {
		if newInspect, err := w.client.ContainerInspect(ctx, newContainerID); err == nil && newInspect.State != nil {
			if newInspect.State.Running {
				currentState = "online"
			} else {
				currentState = "stopped"
			}
		}
	}

	message := "updated and restored state"
	if cfg.PruneDanglingImages {
		deleted, reclaimed, err := w.pruneDanglingImages(ctx)
		if err != nil {
			message = fmt.Sprintf("updated; prune failed: %v", err)
		} else {
			reclaimedMB := float64(reclaimed) / (1024.0 * 1024.0)
			message = fmt.Sprintf("updated and pruned dangling images (deleted=%d reclaimed=%.1f MB)", deleted, reclaimedMB)
		}
	}

	return UpdateResult{
		ID:             newContainerID,
		Name:           name,
		Image:          imageRef,
		OldImageID:     oldImageID,
		NewImageID:     newImageID,
		AppliedImageID: appliedImageID,
		Updated:        true,
		PreviousState:  previousState,
		CurrentState:   currentState,
		Message:        message,
	}, nil
}

func (w *Watcher) pruneDanglingImages(ctx context.Context) (int, uint64, error) {
	args := filters.NewArgs()
	args.Add("dangling", "true")
	report, err := w.client.ImagesPrune(ctx, args)
	if err != nil {
		return 0, 0, err
	}
	deleted := 0
	if report.ImagesDeleted != nil {
		deleted = len(report.ImagesDeleted)
	}
	return deleted, report.SpaceReclaimed, nil
}

func IsSelfContainer(containerID string) bool {
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		return false
	}
	return strings.HasPrefix(containerID, hostname)
}

func (w *Watcher) TriggerSelfUpdate(ctx context.Context, containerID string) error {
	return w.triggerSelfUpdate(ctx, containerID, nil)
}

func (w *Watcher) TriggerSelfUpdateWithLogs(ctx context.Context, containerID string, logFn func(level, message string)) error {
	return w.triggerSelfUpdate(ctx, containerID, logFn)
}

func (w *Watcher) triggerSelfUpdate(ctx context.Context, containerID string, logFn func(level, message string)) error {
	inspect, err := w.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return err
	}
	imageRef := ""
	if inspect.Config != nil {
		imageRef = inspect.Config.Image
	}
	if imageRef == "" {
		return errors.New("missing image reference")
	}
	if err := w.pullImage(ctx, imageRef); err != nil {
		return err
	}

	nameSuffix := containerID
	if len(nameSuffix) > 12 {
		nameSuffix = nameSuffix[:12]
	}
	helperName := "contiwatch-updater-" + nameSuffix

	env := []string{
		"CONTIWATCH_CONFIG=/data/config.json",
	}
	var binds []string
	binds = append(binds, "/var/run/docker.sock:/var/run/docker.sock")
	for _, mount := range inspect.Mounts {
		if mount.Destination == "/data" {
			binds = append(binds, mount.Source+":/data")
			break
		}
	}

	helperConfig := &container.Config{
		Image:      imageRef,
		Entrypoint: []string{"/app/contiwatch"},
		Cmd:        []string{"--self-update", containerID},
		Env:        env,
	}
	helperHost := &container.HostConfig{
		AutoRemove: true,
		Binds:      binds,
	}

	created, err := w.client.ContainerCreate(ctx, helperConfig, helperHost, nil, nil, helperName)
	if err != nil {
		return err
	}
	if err := w.client.ContainerStart(ctx, created.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}
	if logFn != nil {
		logFn("info", fmt.Sprintf("self-update helper started: %s", helperName))
		go w.streamContainerLogs(created.ID, logFn)
	}
	return nil
}

type logLineWriter struct {
	level string
	logFn func(level, message string)
	buf   string
}

func (w *logLineWriter) Write(p []byte) (int, error) {
	w.buf += string(p)
	for {
		idx := strings.IndexByte(w.buf, '\n')
		if idx < 0 {
			break
		}
		line := strings.TrimRight(w.buf[:idx], "\r")
		w.buf = w.buf[idx+1:]
		if strings.TrimSpace(line) != "" {
			w.logFn(w.level, line)
		}
	}
	return len(p), nil
}

func (w *logLineWriter) Flush() {
	line := strings.TrimSpace(w.buf)
	if line != "" {
		w.logFn(w.level, line)
	}
	w.buf = ""
}

func (w *Watcher) streamContainerLogs(containerID string, logFn func(level, message string)) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	reader, err := w.client.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		logFn("warn", fmt.Sprintf("self-update helper logs unavailable: %v", err))
		return
	}
	defer reader.Close()
	stdoutWriter := &logLineWriter{level: "info", logFn: logFn}
	stderrWriter := &logLineWriter{level: "warn", logFn: logFn}
	_, _ = stdcopy.StdCopy(stdoutWriter, stderrWriter, bufio.NewReader(reader))
	stdoutWriter.Flush()
	stderrWriter.Flush()
}
