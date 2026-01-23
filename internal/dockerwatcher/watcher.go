package dockerwatcher

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
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
	if err != nil || remoteDigest == "" || localDigest == "" {
		status.Error = "skipped: digest unknown (registry or local digest unavailable)"
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
