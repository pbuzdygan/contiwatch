package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

type RemoteServer struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Token string `json:"token"`
}

type LocalServer struct {
	Name   string `json:"name"`
	Socket string `json:"socket"`
}

type Config struct {
	ScanIntervalSec          int            `json:"scan_interval_sec"`
	SchedulerEnabled         bool           `json:"scheduler_enabled"`
	GlobalPolicy             string         `json:"global_policy"`
	DiscordWebhookURL           string `json:"discord_webhook_url"`
	DiscordNotificationsEnabled *bool  `json:"discord_notifications_enabled"`
	UpdateStoppedContainers  bool           `json:"update_stopped_containers"`
	LocalServers             []LocalServer  `json:"local_servers"`
	RemoteServers            []RemoteServer `json:"remote_servers"`
}

const (
	PolicyUpdate     = "update"
	PolicyNotifyOnly = "notify_only"
	PolicySkip       = "skip"
)

func DefaultConfig() Config {
	return Config{
		ScanIntervalSec:         60 * 1440,
		SchedulerEnabled:        false,
		GlobalPolicy:            PolicyNotifyOnly,
		DiscordWebhookURL:       "",
		DiscordNotificationsEnabled: boolPtr(true),
		UpdateStoppedContainers: false,
		LocalServers:           []LocalServer{},
		RemoteServers:           []RemoteServer{},
	}
}

type Store struct {
	mu     sync.RWMutex
	path   string
	config Config
}

func NewStore(path string) (*Store, error) {
	if path == "" {
		return nil, errors.New("config path is required")
	}
	store := &Store{path: path}
	if err := store.loadOrCreate(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *Store) loadOrCreate() error {
	if _, err := os.Stat(s.path); err == nil {
		return s.load()
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	s.config = DefaultConfig()
	return s.saveLocked()
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}
	if cfg.ScanIntervalSec <= 0 {
		cfg.ScanIntervalSec = DefaultConfig().ScanIntervalSec
	}
	if cfg.ScanIntervalSec == 300 &&
		!cfg.SchedulerEnabled &&
		cfg.DiscordWebhookURL == "" &&
		cfg.UpdateStoppedContainers == false &&
		cfg.GlobalPolicy == DefaultConfig().GlobalPolicy {
		cfg.ScanIntervalSec = DefaultConfig().ScanIntervalSec
	}
	if cfg.GlobalPolicy == "" {
		cfg.GlobalPolicy = DefaultConfig().GlobalPolicy
	}
	if cfg.DiscordNotificationsEnabled == nil {
		cfg.DiscordNotificationsEnabled = boolPtr(true)
	}
	s.config = cfg
	return nil
}

func (s *Store) saveLocked() error {
	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

func (s *Store) Get() Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

func (s *Store) Update(update func(*Config)) (Config, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	update(&s.config)
	if s.config.ScanIntervalSec <= 0 {
		s.config.ScanIntervalSec = DefaultConfig().ScanIntervalSec
	}
	if s.config.GlobalPolicy == "" {
		s.config.GlobalPolicy = DefaultConfig().GlobalPolicy
	}
	if s.config.DiscordNotificationsEnabled == nil {
		s.config.DiscordNotificationsEnabled = boolPtr(true)
	}
	return s.config, s.saveLocked()
}

func boolPtr(value bool) *bool {
	return &value
}
