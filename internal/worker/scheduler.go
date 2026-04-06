package worker

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/pavanrkadave/uptime-monitor/internal/domain"
)

type MonitorProvider interface {
	ListAll(ctx context.Context) ([]*domain.Monitor, error)
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

	log.Printf("Starting background worker, pinging every %s\n", interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("Worker shutting down safely..")
			return
		case <-ticker.C:
			s.runCheckCycle(ctx)
		}
	}
}

func (s *Scheduler) runCheckCycle(ctx context.Context) {
	monitors, err := s.provider.ListAll(ctx)
	if err != nil {
		log.Printf("Worker failed to list all monitors: %v\n", err)
		return
	}

	if len(monitors) == 0 {
		log.Printf("Nothing to monitor..\n")
		return
	}

	log.Printf("Worker waking up: checking %d monitors concurrently...\n", len(monitors))
	var wg sync.WaitGroup

	for _, monitor := range monitors {
		wg.Add(1)

		go func(url string) {
			defer wg.Done()

			result := PingSite(ctx, monitor.URL)

			if result.IsUp {
				log.Printf("✅ UP: %s (Status: %d, Time: %v)\n", result.URL, result.StatusCode, result.Duration)
			} else {
				log.Printf("❌ DOWN: %s (Error: %s)\n", result.URL, result.ErrorMessage)
			}
		}(monitor.URL)
	}
	wg.Wait()
	log.Println("Worker cycle complete. Sleeping until next check...")
}
