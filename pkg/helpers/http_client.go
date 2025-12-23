package helpers

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

var client = &http.Client{}

func PostJSON(ctx context.Context, url string, payload io.Reader) ([]byte, error) {
	slog.DebugContext(ctx, "POST JSON", slog.String("url", url))

	req, err := http.NewRequestWithContext(ctx, "POST", url, payload)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create HTTP request", slog.String("url", url), slog.Any("error", err))
		return nil, err
	}
	req.Header.Set("accept", "*/*")
	req.Header.Set("accept-language", "en-US,en;q=0.9")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36")

	// Let's do it!
	resp, err := client.Do(req)
	if err != nil {
		// oh no
		slog.ErrorContext(ctx, "Failed to perform HTTP request", slog.String("url", url), slog.Any("error", err))
		return nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			slog.ErrorContext(ctx, "Error closing response body", slog.String("url", url), slog.Any("error", closeErr))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: received status code %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func GetJSON(ctx context.Context, url string) ([]byte, error) {
	slog.DebugContext(ctx, "GET JSON", slog.String("url", url))
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create HTTP request", slog.String("url", url), slog.Any("error", err))
		return nil, err
	}
	req.Header.Set("accept", "*/*")
	req.Header.Set("accept-language", "en-US,en;q=0.9")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36")

	// Let's do it!
	resp, err := client.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to perform HTTP request", slog.String("url", url), slog.Any("error", err))
		return nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			slog.ErrorContext(ctx, "Error closing response body", slog.String("url", url), slog.Any("error", closeErr))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		slog.ErrorContext(ctx, "Received non-OK HTTP status", slog.String("url", url), slog.Int("status_code", resp.StatusCode))
		return nil, fmt.Errorf("error: received status code %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
