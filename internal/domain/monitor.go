package domain

import (
	"errors"
	"net/url"
	"strings"
	"time"
)

var (
	ErrMonitorNotFound = errors.New("monitor not found")
	ErrEmptyURL        = errors.New("URL cannot be empty")
	ErrInvalidURL      = errors.New("URL is not valid")
	ErrMissingScheme   = errors.New("URL must include a scheme (http:// or https://)")
)

type Monitor struct {
	ID        int64      `json:"id"`
	URL       string     `json:"url"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

func (m Monitor) Validate() error {
	trimmedURL := strings.TrimSpace(m.URL)
	if trimmedURL == "" {
		return ErrEmptyURL
	}

	parsedURL, err := url.ParseRequestURI(trimmedURL)
	if err != nil {
		return ErrInvalidURL
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return ErrMissingScheme
	}
	return nil
}

type MonitorStats struct {
	TotalPings       int     `json:"total_pings"`
	UptimePercentage float64 `json:"uptime_percentage"`
	AvgLatencyMs     float64 `json:"avg_latency_ms"`
	P95LatencyMs     float64 `json:"p95_latency_ms"`
}
