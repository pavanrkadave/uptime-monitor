package worker

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/pavanrkadave/uptime-monitor/internal/domain"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	monitorUpGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "uptime_monitor_up",
			Help: "1 if the monitored website is up, 0 if it is down",
		},
		[]string{"url"},
	)

	monitorLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "uptime_monitor_response_duration_seconds",
			Help:    "Response duration of monitored websites in seconds",
			Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"url"})

	tlsCertExpiryGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "uptime_monitor_cert_expiry_timestamp_seconds",
			Help: "The UNIX timestamp of the TLS certificate expiration date",
		},
		[]string{"url"},
	)
)

type MonitorProvider interface {
	ListAll(ctx context.Context) ([]*domain.Monitor, error)
	SavePingResult(ctx context.Context, monitorID int64, isUp bool, statusCode int, duration time.Duration, errMsg string) error
}

type Scheduler struct {
	provider MonitorProvider
	log      *slog.Logger
}

func New(provider MonitorProvider, log *slog.Logger) *Scheduler {
	return &Scheduler{
		provider: provider,
		log:      log.With("component", "monitor-scheduler"),
	}
}

func (s *Scheduler) Start(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	s.log.Info("Starting background worker ", slog.Any("interval", interval))

	for {
		select {
		case <-ctx.Done():
			s.log.Info("Worker shutting down safely..")
			return
		case <-ticker.C:
			s.runCheckCycle(ctx)
		}
	}
}

func (s *Scheduler) runCheckCycle(ctx context.Context) {
	monitors, err := s.provider.ListAll(ctx)
	if err != nil {
		s.log.Error("Worker failed to list all monitors ", slog.Any("error", err))
		return
	}

	if len(monitors) == 0 {
		s.log.Info("Nothing to monitor, sleeping until next check...")
		return
	}

	s.log.Info("Worker waking up: checking monitors concurrently.", slog.Any("monitors", len(monitors)))
	var wg sync.WaitGroup

	for _, monitor := range monitors {
		wg.Add(1)

		go func(monitor *domain.Monitor) {
			defer wg.Done()

			result := PingSite(ctx, monitor.URL, monitor.ExpectedKeyword)

			gaugeValue := 0.0
			if result.IsUp {
				gaugeValue = 1.0
				s.log.Info("Link is UP ✅", slog.Any("url", result.URL), slog.Any("status_code", result.StatusCode), slog.Any("duration", result.Duration.Milliseconds()))
			} else {
				s.log.Info("Link is DOWN ❌", slog.Any("url", result.URL), slog.String("error", result.ErrorMessage))
			}

			monitorUpGauge.WithLabelValues(monitor.URL).Set(gaugeValue)
			monitorLatency.WithLabelValues(monitor.URL).Observe(result.Duration.Seconds())

			if result.TLSCertExpiry != nil {
				tlsCertExpiryGauge.WithLabelValues(monitor.URL).Set(float64(result.TLSCertExpiry.Unix()))
			}

			dbCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			err := s.provider.SavePingResult(dbCtx, monitor.ID, result.IsUp, result.StatusCode, result.Duration, result.ErrorMessage)
			if err != nil {
				s.log.Error("⚠️ Failed to save ping result", slog.String("monitor", monitor.URL), slog.Any("error", err))
			}
		}(monitor)
	}
	wg.Wait()
	s.log.Info("Worker cycle complete. Sleeping until next check...")
}
