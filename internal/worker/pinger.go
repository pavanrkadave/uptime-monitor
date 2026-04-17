package worker

import (
	"context"
	"io"
	"net/http"
	"strings"
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

func PingSite(ctx context.Context, url, expectedKeyword string) PingResult {

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

	bodyBytes, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		return PingResult{
			URL:          url,
			IsUp:         false,
			StatusCode:   response.StatusCode,
			Duration:     duration,
			ErrorMessage: "failed to read response body: " + readErr.Error(),
		}
	}

	var certExpiry *time.Time
	if response.TLS != nil && len(response.TLS.PeerCertificates) > 0 {
		expiry := response.TLS.PeerCertificates[0].NotAfter
		certExpiry = &expiry
	}

	isUp := response.StatusCode >= 200 && response.StatusCode < 400
	errorMessage := ""

	if isUp && expectedKeyword != "" {
		if !strings.Contains(string(bodyBytes), expectedKeyword) {
			isUp = false
			errorMessage = "keyword '" + expectedKeyword + "' not found in response"
		}
	}

	return PingResult{
		URL:           url,
		IsUp:          isUp,
		StatusCode:    response.StatusCode,
		Duration:      duration,
		ErrorMessage:  errorMessage,
		TLSCertExpiry: certExpiry,
	}
}
