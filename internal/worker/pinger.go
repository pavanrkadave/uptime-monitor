package worker

import (
	"context"
	"net/http"
	"time"
)

type PingResult struct {
	URL           string
	IsUp          bool
	StatusCode    int
	Duration      time.Duration
	ErrorMessage  string
	TLSCertExpiry *time.Time
}

func PingSite(ctx context.Context, url string) PingResult {

	client := &http.Client{
		Timeout: time.Second * 10,
	}
	startTime := time.Now()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return PingResult{
			URL:          url,
			IsUp:         false,
			StatusCode:   0,
			Duration:     time.Since(startTime),
			ErrorMessage: "failed to build request: " + err.Error(),
		}
	}

	response, err := client.Do(request)
	duration := time.Since(startTime)

	if err != nil {
		return PingResult{
			URL:          url,
			IsUp:         false,
			StatusCode:   0,
			Duration:     duration,
			ErrorMessage: "network error: " + err.Error(),
		}
	}
	defer response.Body.Close()

	var certExpiry *time.Time
	if response.TLS != nil && len(response.TLS.PeerCertificates) > 0 {
		certExpiry = new(response.TLS.PeerCertificates[0].NotAfter)
	}

	isUp := response.StatusCode >= 200 && response.StatusCode < 400
	return PingResult{
		URL:           url,
		IsUp:          isUp,
		StatusCode:    response.StatusCode,
		Duration:      duration,
		TLSCertExpiry: certExpiry,
	}
}
