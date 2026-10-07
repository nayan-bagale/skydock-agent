package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const maxResponseBytes = 8 << 20

// Error is a non-2xx response from the API.
type Error struct {
	StatusCode int
	Body       string
}

func (e *Error) Error() string {
	return fmt.Sprintf("api status %d: %s", e.StatusCode, e.Body)
}

// Client calls the SkyDock API with JSON bodies and an optional bearer token.
type Client struct {
	baseURL string
	http    *http.Client

	mu    sync.RWMutex
	token string
}

// New builds a client for baseURL. timeout is the http.Client deadline for each request.
func New(baseURL string, timeout time.Duration) (*Client, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return nil, fmt.Errorf("api base url is invalid: %q", baseURL)
	}
	if timeout <= 0 {
		return nil, fmt.Errorf("api timeout must be positive")
	}

	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: timeout},
	}, nil
}

// BaseURL returns the API origin and prefix this client calls.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// SetAccessToken stores the bearer token sent on later requests.
// An empty token clears the header.
func (c *Client) SetAccessToken(token string) {
	c.mu.Lock()
	c.token = strings.TrimSpace(token)
	c.mu.Unlock()
}

func (c *Client) accessToken() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.token
}

// DoJSON sends method to path under the base URL.
// body is JSON-encoded when non-nil. A 2xx body is decoded into dest when dest is non-nil.
func (c *Client) DoJSON(ctx context.Context, method, path string, body, dest any) error {
	endpoint, err := c.resolve(path)
	if err != nil {
		return err
	}

	var payload io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode api request: %w", err)
		}
		payload = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, payload)
	if err != nil {
		return fmt.Errorf("create api request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token := c.accessToken(); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("api request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return fmt.Errorf("read api response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return &Error{StatusCode: resp.StatusCode, Body: string(raw)}
	}

	if dest == nil || len(bytes.TrimSpace(raw)) == 0 {
		return nil
	}

	if err := json.Unmarshal(raw, dest); err != nil {
		return fmt.Errorf("decode api response: %w", err)
	}

	return nil
}

func (c *Client) resolve(path string) (string, error) {
	rel, err := url.Parse(path)
	if err != nil {
		return "", fmt.Errorf("api path: %w", err)
	}
	if rel.IsAbs() && rel.Host != "" {
		return "", fmt.Errorf("api path must be relative: %q", path)
	}

	joined, err := url.JoinPath(c.baseURL, strings.TrimPrefix(rel.Path, "/"))
	if err != nil {
		return "", fmt.Errorf("api path: %w", err)
	}

	endpoint, err := url.Parse(joined)
	if err != nil {
		return "", fmt.Errorf("api path: %w", err)
	}
	endpoint.RawQuery = rel.RawQuery
	endpoint.Fragment = rel.Fragment
	return endpoint.String(), nil
}
