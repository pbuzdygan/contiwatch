package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"contiwatch/internal/config"

	"github.com/robfig/cron/v3"
)

var cronParser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

type schedulerPlan struct {
	mode     string
	expr     string
	interval time.Duration
}

type schedulerPreviewResponse struct {
	TimeZone string   `json:"tz"`
	Now      string   `json:"now"`
	NextRuns []string `json:"next_runs"`
	Error    string   `json:"error,omitempty"`
}

type schedulerStatusResponse struct {
	Enabled  bool     `json:"enabled"`
	Mode     string   `json:"mode"`
	TimeZone string   `json:"tz"`
	NextRuns []string `json:"next_runs"`
	Error    string   `json:"error,omitempty"`
}

func (s *Server) handleSchedulerPreviewBasic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	daysParam := strings.TrimSpace(r.URL.Query().Get("days"))
	timeParam := strings.TrimSpace(r.URL.Query().Get("time"))
	count := parsePreviewCount(r.URL.Query().Get("count"))
	days, err := parseDaysParam(daysParam)
	if err != nil {
		writeJSON(w, http.StatusOK, schedulerPreviewResponse{TimeZone: time.Local.String(), Now: time.Now().In(time.Local).Format(time.RFC3339), Error: err.Error()})
		return
	}
	expr, err := buildCronFromBasic(days, timeParam)
	if err != nil {
		writeJSON(w, http.StatusOK, schedulerPreviewResponse{TimeZone: time.Local.String(), Now: time.Now().In(time.Local).Format(time.RFC3339), Error: err.Error()})
		return
	}
	nextRuns, err := nextRunsFromCron(expr, count)
	resp := schedulerPreviewResponse{TimeZone: time.Local.String(), Now: time.Now().In(time.Local).Format(time.RFC3339)}
	if err != nil {
		resp.Error = err.Error()
	} else {
		resp.NextRuns = nextRuns
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleSchedulerPreviewCron(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	expr := strings.TrimSpace(r.URL.Query().Get("expr"))
	count := parsePreviewCount(r.URL.Query().Get("count"))
	nextRuns, err := nextRunsFromCron(expr, count)
	resp := schedulerPreviewResponse{TimeZone: time.Local.String(), Now: time.Now().In(time.Local).Format(time.RFC3339)}
	if err != nil {
		resp.Error = err.Error()
	} else {
		resp.NextRuns = nextRuns
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleSchedulerStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	count := parsePreviewCount(r.URL.Query().Get("count"))
	s.schedulerMu.Lock()
	enabled := s.schedulerEnabled
	mode := s.schedulerMode
	next := s.schedulerNextRun
	s.schedulerMu.Unlock()
	resp := schedulerStatusResponse{
		Enabled:  enabled,
		Mode:     mode,
		TimeZone: time.Local.String(),
	}
	if enabled && !next.IsZero() && count > 0 {
		resp.NextRuns = []string{next.In(time.Local).Format(time.RFC3339)}
	}
	writeJSON(w, http.StatusOK, resp)
}

func buildSchedulerPlan(cfg config.Config) (schedulerPlan, error) {
	mode := strings.TrimSpace(cfg.SchedulerPlan.Mode)
	if mode == "" {
		mode = config.SchedulerModeLegacy
	}
	switch mode {
	case config.SchedulerModeLegacy:
		interval := time.Duration(cfg.ScanIntervalSec) * time.Second
		if interval <= 0 {
			interval = 5 * time.Minute
		}
		return schedulerPlan{mode: mode, interval: interval}, nil
	case config.SchedulerModeBasic:
		if cfg.SchedulerPlan.Basic == nil {
			return schedulerPlan{}, errors.New("basic schedule is required")
		}
		expr, err := buildCronFromBasic(cfg.SchedulerPlan.Basic.Days, cfg.SchedulerPlan.Basic.Time)
		if err != nil {
			return schedulerPlan{}, err
		}
		return schedulerPlan{mode: mode, expr: expr}, nil
	case config.SchedulerModeCron:
		if cfg.SchedulerPlan.Cron == nil || strings.TrimSpace(cfg.SchedulerPlan.Cron.Expr) == "" {
			return schedulerPlan{}, errors.New("cron expression is required")
		}
		expr := strings.TrimSpace(cfg.SchedulerPlan.Cron.Expr)
		if _, err := cronParser.Parse(expr); err != nil {
			return schedulerPlan{}, fmt.Errorf("invalid cron: %w", err)
		}
		return schedulerPlan{mode: mode, expr: expr}, nil
	default:
		return schedulerPlan{}, fmt.Errorf("unknown scheduler mode: %s", mode)
	}
}

func buildCronFromBasic(days []int, timeValue string) (string, error) {
	if len(days) == 0 {
		return "", errors.New("select at least one day")
	}
	parsed, err := time.Parse("15:04", timeValue)
	if err != nil {
		return "", errors.New("time must be HH:MM")
	}
	hour := parsed.Hour()
	minute := parsed.Minute()
	dowList, err := normalizeBasicDays(days)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d %d * * %s", minute, hour, dowList), nil
}

func normalizeBasicDays(days []int) (string, error) {
	unique := map[int]struct{}{}
	for _, day := range days {
		if day < 1 || day > 7 {
			return "", fmt.Errorf("invalid day: %d", day)
		}
		unique[day] = struct{}{}
	}
	values := make([]int, 0, len(unique))
	for day := range unique {
		if day == 7 {
			values = append(values, 0)
		} else {
			values = append(values, day)
		}
	}
	sort.Ints(values)
	parts := make([]string, 0, len(values))
	for _, day := range values {
		parts = append(parts, strconv.Itoa(day))
	}
	return strings.Join(parts, ","), nil
}

func nextRunsFromCron(expr string, count int) ([]string, error) {
	if strings.TrimSpace(expr) == "" {
		return nil, errors.New("cron expression is required")
	}
	if count <= 0 {
		count = 5
	}
	schedule, err := cronParser.Parse(expr)
	if err != nil {
		return nil, fmt.Errorf("invalid cron: %w", err)
	}
	now := time.Now().In(time.Local)
	nextRuns := make([]string, 0, count)
	next := now
	for i := 0; i < count; i += 1 {
		next = schedule.Next(next)
		nextRuns = append(nextRuns, next.In(time.Local).Format(time.RFC3339))
	}
	return nextRuns, nil
}

func parsePreviewCount(raw string) int {
	count, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || count <= 0 {
		return 5
	}
	if count > 15 {
		return 15
	}
	return count
}

func parseDaysParam(raw string) ([]int, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, errors.New("select at least one day")
	}
	parts := strings.Split(raw, ",")
	days := make([]int, 0, len(parts))
	for _, part := range parts {
		val, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return nil, errors.New("invalid day list")
		}
		days = append(days, val)
	}
	return days, nil
}

func shouldApplySchedulerPlan(cfg config.Config) bool {
	return strings.TrimSpace(cfg.SchedulerPlan.Mode) != "" || cfg.SchedulerPlan.Basic != nil || cfg.SchedulerPlan.Cron != nil
}

func (s *Server) runScheduledScan(runCtx context.Context) {
	currentCfg := s.store.Get()
	s.setAllScanStates(currentCfg, scanStatePending)
	s.addLog("info", "scheduled scan started")
	result, err := s.runScan(runCtx)
	if err != nil && !errors.Is(err, errScanInProgress) {
		s.addLog("error", fmt.Sprintf("scheduled scan failed: %v", err))
	}
	if len(currentCfg.RemoteServers) > 0 {
		s.addLog("info", fmt.Sprintf("scheduled remote scans started: servers=%d", len(currentCfg.RemoteServers)))
		remoteCtx, cancel := context.WithTimeout(runCtx, 2*time.Minute)
		s.triggerRemoteScans(remoteCtx, currentCfg.RemoteServers)
		cancel()
		s.addLog("info", "scheduled remote scans finished")
	}
	if err == nil {
		s.addLog("info", fmt.Sprintf("scheduled scan finished: containers=%d", len(result.Containers)))
	}
}

func (s *Server) runSchedulerLoop(runCtx context.Context, plan schedulerPlan) {
	switch plan.mode {
	case config.SchedulerModeLegacy:
		s.runIntervalScheduler(runCtx, plan.interval, plan.mode)
	default:
		schedule, err := cronParser.Parse(plan.expr)
		if err != nil {
			s.addLog("error", fmt.Sprintf("scheduler cron parse failed: %v", err))
			return
		}
		s.runCronScheduler(runCtx, schedule, plan.mode)
	}
}

func (s *Server) runIntervalScheduler(runCtx context.Context, interval time.Duration, mode string) {
	if interval <= 0 {
		interval = 5 * time.Minute
	}
	for {
		next := time.Now().Add(interval)
		s.setSchedulerNextRun(next, mode)
		timer := time.NewTimer(time.Until(next))
		select {
		case <-runCtx.Done():
			timer.Stop()
			return
		case <-timer.C:
			s.runScheduledScan(runCtx)
		}
	}
}

func (s *Server) runCronScheduler(runCtx context.Context, schedule cron.Schedule, mode string) {
	next := time.Now().In(time.Local)
	for {
		next = schedule.Next(next)
		s.setSchedulerNextRun(next, mode)
		timer := time.NewTimer(time.Until(next))
		select {
		case <-runCtx.Done():
			timer.Stop()
			return
		case <-timer.C:
			s.runScheduledScan(runCtx)
		}
	}
}

func (s *Server) setSchedulerNextRun(next time.Time, mode string) {
	s.schedulerMu.Lock()
	s.schedulerNextRun = next
	s.schedulerMode = mode
	s.schedulerMu.Unlock()
}

func schedulerPlanKey(plan schedulerPlan) string {
	if plan.mode == config.SchedulerModeLegacy {
		return fmt.Sprintf("%s|%s", plan.mode, plan.interval)
	}
	return fmt.Sprintf("%s|%s", plan.mode, plan.expr)
}
