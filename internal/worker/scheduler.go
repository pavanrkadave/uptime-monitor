package worker

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/pavanrkadave/uptime-monitor/internal/domain"
)

type MonitorProvider interface {
	ListAll(ctx context.Context) ([]*domain.Monitor, error)
	SavePingResult(ctx context.Context, monitorID int64, isUp bool, statusCode int, duration time.Duration, errMsg string) error
}

type Scheduler struct {
	provider MonitorProvider
}

func NewScheduler(provider MonitorProvider) *Scheduler {
	return &Scheduler{
		provider: provider,
	}
}

func (s *Scheduler) Start(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	slog.Info("Starting background worker ", slog.Any("interval", interval))

	for {
		select {
		case <-ctx.Done():
			slog.Info("Worker shutting down safely..")
			return
		case <-ticker.C:
			s.runCheckCycle(ctx)
		}
	}
}

func (s *Scheduler) runCheckCycle(ctx context.Context) {
	monitors, err := s.provider.ListAll(ctx)
	if err != nil {
		slog.Error("Worker failed to list all monitors ", slog.Any("error", err))
		return
	}

	if len(monitors) == 0 {
		slog.Info("Nothing to monitor, sleeping until next check...")
		return
	}

	slog.Info("Worker waking up: checking monitors concurrently.", slog.Any("monitors", len(monitors)))
	var wg sync.WaitGroup

	for _, monitor := range monitors {
		wg.Add(1)

		go func(monitor *domain.Monitor) {
			defer wg.Done()

			result := PingSite(ctx, monitor.URL)

			if result.IsUp {
				slog.Info("Link is UP ✅", slog.Any("url", result.URL), slog.Any("status_code", result.StatusCode), slog.Any("duration", result.Duration.Milliseconds()))
			} else {
				slog.Info("Link is DOWN ❌", slog.Any("url", result.URL), slog.String("error", result.ErrorMessage))
			}

			dbCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			err := s.provider.SavePingResult(dbCtx, monitor.ID, result.IsUp, result.StatusCode, result.Duration, result.ErrorMessage)
			if err != nil {
				slog.Error("⚠️ Failed to save ping result", slog.String("monitor", monitor.URL), slog.Any("error", err))
			}
		}(monitor)
	}
	wg.Wait()
	slog.Info("Worker cycle complete. Sleeping until next check...")
}
