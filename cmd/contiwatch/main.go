package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"runtime/debug"

	"contiwatch/internal/config"
	"contiwatch/internal/dockerwatcher"
	"contiwatch/internal/notify"
	"contiwatch/internal/server"
)

var Version = "dev"

func discordNotificationsEnabled(cfg config.Config) bool {
	if cfg.DiscordNotificationsEnabled == nil {
		return true
	}
	return *cfg.DiscordNotificationsEnabled
}

func resolveVersion() string {
	if Version != "" && Version != "dev" {
		return strings.TrimPrefix(Version, "v")
	}
	version := strings.TrimSpace(os.Getenv("CONTIWATCH_VERSION"))
	if version != "" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return strings.TrimPrefix(info.Main.Version, "v")
		}
	}
	return "dev"
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-version", "--version":
			fmt.Println(resolveVersion())
			return
		case "--self-update":
			if len(os.Args) < 3 {
				log.Fatalf("self-update requires container id")
			}
			containerID := strings.TrimSpace(os.Args[2])
			if containerID == "" {
				log.Fatalf("self-update requires container id")
			}
			store, err := config.NewStore(server.ResolveConfigPath())
			if err != nil {
				log.Fatalf("config: %v", err)
			}
			watcher, err := dockerwatcher.New()
			if err != nil {
				log.Fatalf("docker: %v", err)
			}
			defer watcher.Close()
			cfg := store.Get()
			if _, err := watcher.UpdateContainer(context.Background(), containerID, cfg); err != nil {
				log.Fatalf("self-update failed: %v", err)
			}
			log.Printf("self-update completed: %s", containerID)
			return
		}
	}
	configPath := server.ResolveConfigPath()
	_, configStatErr := os.Stat(configPath)
	configExisted := configStatErr == nil

	store, err := config.NewStore(configPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	watcher, err := dockerwatcher.New()
	if err != nil {
		log.Fatalf("docker: %v", err)
	}
	defer watcher.Close()

	agentMode := envBool("CONTIWATCH_AGENT")
	agentToken := strings.TrimSpace(os.Getenv("CONTIWATCH_AGENT_TOKEN"))
	if agentMode && agentToken == "" {
		log.Fatalf("agent mode requires CONTIWATCH_AGENT_TOKEN")
	}

	version := resolveVersion()
	srv := server.New(store, watcher, agentMode, agentToken, version)
	cfg := store.Get()
	if agentMode && len(cfg.LocalServers) == 0 {
		if _, err := os.Stat("/var/run/docker.sock"); err == nil {
			updated, err := store.Update(func(c *config.Config) {
				c.LocalServers = []config.LocalServer{{Name: "local", Socket: "/var/run/docker.sock"}}
			})
			if err == nil {
				cfg = updated
				log.Printf("startup: agent default local server configured (/var/run/docker.sock)")
			} else {
				log.Printf("startup: agent default local server not configured: %v", err)
			}
		} else {
			log.Printf("startup: agent local docker socket not found: %v", err)
		}
	}
	srv.UpdateDiscord(cfg.DiscordWebhookURL)

	addr := resolveAddr()
	schedulerInterval := time.Duration(cfg.ScanIntervalSec) * time.Second

	log.Printf("startup: http addr=%s", addr)
	if configExisted {
		log.Printf("startup: config path=%s (loaded)", configPath)
	} else {
		log.Printf("startup: config path=%s (created defaults)", configPath)
	}
	log.Printf(
		"startup: env CONTIWATCH_ADDR=%q CONTIWATCH_CONFIG=%q",
		os.Getenv("CONTIWATCH_ADDR"),
		os.Getenv("CONTIWATCH_CONFIG"),
	)
	log.Printf("startup: global policy=%s", cfg.GlobalPolicy)
	log.Printf("startup: update stopped containers=%t", cfg.UpdateStoppedContainers)
	log.Printf("startup: discord notifications=%t", discordNotificationsEnabled(cfg))
	log.Printf("startup: release=%s", version)
	log.Printf("startup: agent mode=%t", agentMode)
	if agentMode {
		log.Printf("startup: agent api=enabled (token required)")
	}
	if len(cfg.RemoteServers) == 0 {
		log.Printf("startup: remote servers=none configured")
	} else {
		names := make([]string, 0, len(cfg.RemoteServers))
		for _, remote := range cfg.RemoteServers {
			if remote.Name != "" {
				names = append(names, remote.Name)
			}
		}
		if len(names) == 0 {
			log.Printf("startup: remote servers=%d", len(cfg.RemoteServers))
		} else {
			log.Printf("startup: remote servers=%d (%s)", len(cfg.RemoteServers), strings.Join(names, ", "))
		}
	}
	if cfg.DiscordWebhookURL == "" {
		log.Printf("startup: discord webhook=not configured (startup notification skipped)")
	} else {
		log.Printf("startup: discord webhook=configured")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if !cfg.SchedulerEnabled {
		log.Printf("startup: scheduler=disabled (default)")
	} else {
		if schedulerInterval <= 0 {
			log.Printf("startup: scheduler=enabled (default interval)")
		} else {
			log.Printf("startup: scheduler=enabled (interval=%s)", schedulerInterval)
		}
	}
	srv.StartScheduler(ctx)

	if cfg.DiscordWebhookURL != "" && discordNotificationsEnabled(cfg) {
		description := buildStartupDiscordDescription(addr, configPath, configExisted, cfg, schedulerInterval)
		const discordBlue = 0x3498DB
		discordClient := notify.NewDiscordClient(cfg.DiscordWebhookURL)
		go func() {
			if err := discordClient.SendEmbed("Contiwatch started", description, discordBlue); err != nil {
				log.Printf("startup: discord notification failed: %v", err)
			}
		}()
	}

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           srv,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("contiwatch listening on %s", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	_ = httpServer.Shutdown(shutdownCtx)
}

func resolveAddr() string {
	if addr := os.Getenv("CONTIWATCH_ADDR"); addr != "" {
		return addr
	}
	return ":8080"
}

func envBool(name string) bool {
	value := strings.TrimSpace(os.Getenv(name))
	value = strings.ToLower(value)
	switch value {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

func buildStartupDiscordDescription(addr, configPath string, configExisted bool, cfg config.Config, schedulerInterval time.Duration) string {
	schedulerLine := "Scheduler: disabled"
	if cfg.SchedulerEnabled {
		schedulerLine = "Scheduler: enabled"
		if schedulerInterval > 0 {
			schedulerLine = "Scheduler: enabled (every " + schedulerInterval.String() + ")"
		}
	}

	remoteLine := "Remote servers: none configured"
	remoteNames := []string{}
	if len(cfg.RemoteServers) > 0 {
		remoteLine = "Remote servers: " + fmt.Sprintf("%d", len(cfg.RemoteServers))
		for _, remote := range cfg.RemoteServers {
			if remote.Name != "" {
				remoteNames = append(remoteNames, remote.Name)
			}
		}
	}

	localLine := "Local servers: none configured"
	localNames := []string{}
	if len(cfg.LocalServers) > 0 {
		localLine = "Local servers: " + fmt.Sprintf("%d", len(cfg.LocalServers))
		for _, local := range cfg.LocalServers {
			if local.Name != "" {
				localNames = append(localNames, local.Name)
			}
		}
	}

	lines := []string{
		schedulerLine,
		"Global policy: " + cfg.GlobalPolicy,
		"Update stopped containers: " + fmt.Sprintf("%t", cfg.UpdateStoppedContainers),
		"Discord notifications: " + fmt.Sprintf("%t", discordNotificationsEnabled(cfg)),
		remoteLine,
	}
	if len(remoteNames) > 0 {
		lines = append(lines, "- "+strings.Join(remoteNames, "\n- "))
	}
	lines = append(lines, localLine)
	if len(localNames) > 0 {
		lines = append(lines, "- "+strings.Join(localNames, "\n- "))
	}
	return strings.Join(lines, "\n")
}
