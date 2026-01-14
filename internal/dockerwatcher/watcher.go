package dockerwatcher

import (
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
	ID            string `json:"id"`
	Name          string `json:"name"`
	Image         string `json:"image"`
	Updated       bool   `json:"updated"`
	PreviousState string `json:"previous_state"`
	CurrentState  string `json:"current_state"`
	Message       string `json:"message,omitempty"`
}

// ScanResult describes a full scan for a single server.
type ScanResult struct {
	ServerName string            `json:"server_name"`
	ServerURL  string            `json:"server_url"`
	Local      bool              `json:"local"`
	CheckedAt  time.Time         `json:"checked_at"`
	Containers []ContainerStatus `json:"containers"`
	Error      string            `json:"error,omitempty"`
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
		status := w.scanContainer(ctx, item, cfg)
		result.Containers = append(result.Containers, status)
	}

	return result, nil
}

func (w *Watcher) scanContainer(ctx context.Context, item types.Container, cfg config.Config) ContainerStatus {
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
		status.Error = ErrSkipContainer.Error()
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

	if err := w.pullImage(ctx, imageRef); err != nil {
		status.Error = err.Error()
		return status
	}

	newImage, _, err := w.client.ImageInspectWithRaw(ctx, imageRef)
	if err != nil {
		status.Error = err.Error()
		return status
	}

	status.NewImageID = newImage.ID
	status.UpdateAvailable = newImage.ID != "" && newImage.ID != inspect.Image
	if !status.UpdateAvailable {
		return status
	}

	if status.Policy == config.PolicyUpdate {
		if inspect.State != nil && inspect.State.Paused {
			status.Error = "skipped: container paused"
			return status
		}

		wasRunning := inspect.State != nil && inspect.State.Running
		if !wasRunning && !cfg.UpdateStoppedContainers {
			status.Error = "skipped: container not running"
			return status
		}

		if _, err := w.recreateContainer(ctx, inspect, imageRef, wasRunning); err != nil {
			status.Error = err.Error()
			return status
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

	updateAvailable := newImage.ID != "" && newImage.ID != inspect.Image
	if !updateAvailable {
		return UpdateResult{
			ID:            containerID,
			Name:          name,
			Image:         imageRef,
			Updated:       false,
			PreviousState: previousState,
			CurrentState:  previousState,
			Message:       "no update available",
		}, nil
	}

	wasRunning := inspect.State != nil && inspect.State.Running
	if !wasRunning && !cfg.UpdateStoppedContainers {
		return UpdateResult{
			ID:            containerID,
			Name:          name,
			Image:         imageRef,
			Updated:       false,
			PreviousState: previousState,
			CurrentState:  previousState,
			Message:       "skipped: container not running",
		}, nil
	}

	newContainerID, err := w.recreateContainer(ctx, inspect, imageRef, wasRunning)
	if err != nil {
		return UpdateResult{}, err
	}

	currentState := "stopped"
	if wasRunning {
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
		if err := w.pruneDanglingImages(ctx); err != nil {
			message = fmt.Sprintf("updated; prune failed: %v", err)
		} else {
			message = "updated and pruned dangling images"
		}
	}

	return UpdateResult{
		ID:            newContainerID,
		Name:          name,
		Image:         imageRef,
		Updated:       true,
		PreviousState: previousState,
		CurrentState:  currentState,
		Message:       message,
	}, nil
}

func (w *Watcher) pruneDanglingImages(ctx context.Context) error {
	args := filters.NewArgs()
	args.Add("dangling", "true")
	_, err := w.client.ImagesPrune(ctx, args)
	return err
}

func IsSelfContainer(containerID string) bool {
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		return false
	}
	return strings.HasPrefix(containerID, hostname)
}

func (w *Watcher) TriggerSelfUpdate(ctx context.Context, containerID string) error {
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
		Image: imageRef,
		Cmd:   []string{"/app/contiwatch", "--self-update", containerID},
		Env:   env,
	}
	helperHost := &container.HostConfig{
		AutoRemove: true,
		Binds:      binds,
	}

	created, err := w.client.ContainerCreate(ctx, helperConfig, helperHost, nil, nil, helperName)
	if err != nil {
		return err
	}
	return w.client.ContainerStart(ctx, created.ID, types.ContainerStartOptions{})
}
