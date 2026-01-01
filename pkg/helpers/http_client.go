package helpers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

var (
	// ErrNonOKStatusCode is returned when the HTTP response status code is not 200 OK.
	ErrNonOKStatusCode = errors.New("received non-OK status code")
	// ErrNoLDJSONFound is returned when no LD+JSON script tags are found in the HTML.
	ErrNoLDJSONFound = errors.New("no LD+JSON script tags found")
	client           = &http.Client{}
	defaultHeaders   = map[string]string{
		"Accept":          "*/*",
		"Accept-Language": "en-US,en;q=0.9",
		"Content-Type":    "application/json",
		"User-Agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36",
	}
)

// PostJSON performs an HTTP POST request with a JSON payload and returns the response body.
func PostJSON(ctx context.Context, url string, payload io.Reader, headers map[string]string) ([]byte, error) {
	slog.DebugContext(ctx, "POST JSON", slog.String("url", url))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, payload)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create HTTP request", slog.String("url", url), slog.Any("error", err))

		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	if headers == nil {
		headers = defaultHeaders
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Let's do it!
	resp, err := client.Do(req)
	if err != nil {
		// oh no
		slog.ErrorContext(ctx, "Failed to perform HTTP request", slog.String("url", url), slog.Any("error", err))
		return nil, fmt.Errorf("failed to perform HTTP request: %w", err)
	}

	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			slog.ErrorContext(ctx, "Error closing response body", slog.String("url", url), slog.Any("error", closeErr))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", ErrNonOKStatusCode, resp.StatusCode)
	}

	return io.ReadAll(resp.Body) //nolint:wrapcheck // we want to return the original error
}

// GetJSON performs an HTTP GET request and returns the response body.
func GetJSON(ctx context.Context, url string, headers map[string]string) ([]byte, error) {
	slog.DebugContext(ctx, "GET JSON", slog.String("url", url))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create HTTP request", slog.String("url", url), slog.Any("error", err))
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	if headers == nil {
		headers = defaultHeaders
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Let's do it!
	resp, err := client.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to perform HTTP request", slog.String("url", url), slog.Any("error", err))
		return nil, fmt.Errorf("failed to perform HTTP request: %w", err)
	}

	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			slog.ErrorContext(ctx, "Error closing response body", slog.String("url", url), slog.Any("error", closeErr))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		slog.ErrorContext(ctx, "Received non-OK HTTP status", slog.String("url", url), slog.Int("status_code", resp.StatusCode))
		return nil, fmt.Errorf("%w: %d", ErrNonOKStatusCode, resp.StatusCode)
	}

	return io.ReadAll(resp.Body) //nolint:wrapcheck // we want to return the original error
}

// GetLDJSON fetches a URL and extracts the LD+JSON structured data from it.
func GetLDJSON(ctx context.Context, url string) (map[string]any, error) {
	result := make(map[string]any)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create HTTP request", slog.String("url", url), slog.Any("error", err))
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	for key, value := range defaultHeaders {
		req.Header.Set(key, value)
	}

	// Let's do it!
	resp, err := client.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to perform HTTP request", slog.String("url", url), slog.Any("error", err))
		return nil, fmt.Errorf("failed to perform HTTP request: %w", err)
	}

	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			slog.ErrorContext(ctx, "Error closing response body", slog.String("url", url), slog.Any("error", closeErr))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		slog.ErrorContext(ctx, "Received non-OK HTTP status", slog.String("url", url), slog.Int("status_code", resp.StatusCode))
		return nil, fmt.Errorf("%w: %d", ErrNonOKStatusCode, resp.StatusCode)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		slog.ErrorContext(ctx, "Error parsing HTML", slog.String("url", url), slog.Any("error", err))
		return nil, fmt.Errorf("error parsing HTML: %w", err)
	}

	var walk func(*html.Node)

	var ldjson string

	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "script" {
			var isLDJSON bool

			for _, a := range n.Attr {
				if a.Key == "type" && a.Val == "application/ld+json" {
					isLDJSON = true
					break
				}
			}

			if isLDJSON && n.FirstChild != nil {
				ldjson = strings.TrimSpace(n.FirstChild.Data)
				return
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(doc)

	if ldjson != "" {
		err = json.Unmarshal([]byte(ldjson), &result)
		if err != nil {
			slog.ErrorContext(ctx, "Error unmarshaling LD+JSON", slog.String("url", url), slog.Any("error", err))
			return nil, fmt.Errorf("error unmarshaling LD+JSON: %w", err)
		}
	} else {
		slog.ErrorContext(ctx, "No LD+JSON script tags found", slog.String("url", url))
		return nil, fmt.Errorf("%w", ErrNoLDJSONFound)
	}

	return result, nil
}

// SetHTTPClient sets the HTTP client for testing purposes.
func SetHTTPClient(c *http.Client) {
	client = c
}

// ResetHTTPClient resets the HTTP client to the default client.
func ResetHTTPClient() {
	client = &http.Client{}
}
