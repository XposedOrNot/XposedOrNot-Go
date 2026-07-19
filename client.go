package xposedornot

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	defaultBaseURL         = "https://api.xposedornot.com"
	defaultPlusBaseURL     = "https://plus-api.xposedornot.com"
	defaultPasswordBaseURL = "https://passwords.xposedornot.com/api"
	defaultTimeout         = 30 * time.Second
	defaultMaxRetries      = 3
)

// Client is an API client for the XposedOrNot service.
type Client struct {
	baseURL         string
	plusBaseURL     string
	passwordBaseURL string
	apiKey          string
	httpClient      *http.Client
	maxRetries      int
	customHeaders   map[string]string

	allowInsecure bool

	// Rate limiting: 1 request per second for free API users.
	rateMu   sync.Mutex
	lastCall time.Time
}

// NewClient creates a new XposedOrNot API client with the given options.
// It returns an error if any base URL does not use HTTPS, unless
// WithAllowInsecure() is set.
func NewClient(opts ...ClientOption) (*Client, error) {
	c := &Client{
		baseURL:         defaultBaseURL,
		plusBaseURL:     defaultPlusBaseURL,
		passwordBaseURL: defaultPasswordBaseURL,
		httpClient:      &http.Client{Timeout: defaultTimeout},
		maxRetries:      defaultMaxRetries,
	}
	for _, opt := range opts {
		opt(c)
	}
	if !c.allowInsecure {
		for _, u := range []struct{ name, val string }{
			{"baseURL", c.baseURL},
			{"plusBaseURL", c.plusBaseURL},
			{"passwordBaseURL", c.passwordBaseURL},
		} {
			if !strings.HasPrefix(u.val, "https://") {
				return nil, fmt.Errorf("insecure base URL for %s: %q must start with https:// (use WithAllowInsecure() for testing)", u.name, u.val)
			}
		}
	}
	return c, nil
}

// rateLimit enforces a minimum interval of 1 second between requests
// for free API users (no API key). Plus API users bypass rate limiting.
func (c *Client) rateLimit(ctx context.Context) error {
	if c.apiKey != "" {
		return nil
	}

	c.rateMu.Lock()
	var wait time.Duration
	if !c.lastCall.IsZero() {
		elapsed := time.Since(c.lastCall)
		if elapsed < time.Second {
			wait = time.Second - elapsed
		}
	}
	// Reserve the slot by setting lastCall to when the request will happen
	c.lastCall = time.Now().Add(wait)
	c.rateMu.Unlock()

	// Sleep outside the lock
	if wait > 0 {
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

// doRequest executes an HTTP request with retry logic for 429 responses.
// It applies rate limiting, custom headers, and the API key header when set.
func (c *Client) doRequest(ctx context.Context, method, url string) ([]byte, error) {
	if err := c.rateLimit(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, url, nil)
		if err != nil {
			return nil, &ErrNetwork{Err: fmt.Errorf("creating request: %w", err)}
		}

		req.Header.Set("Accept", "application/json")
		if c.apiKey != "" {
			req.Header.Set("x-api-key", c.apiKey)
		}
		for k, v := range c.customHeaders {
			req.Header.Set(k, v)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, &ErrNetwork{Err: fmt.Errorf("executing request: %w", err)}
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, &ErrNetwork{Err: fmt.Errorf("reading response body: %w", err)}
		}

		switch {
		case resp.StatusCode >= 200 && resp.StatusCode < 300:
			return body, nil
		case resp.StatusCode == 429:
			lastErr = &ErrRateLimit{Message: string(body)}
			if attempt < c.maxRetries {
				delay := time.Duration(1<<uint(attempt)) * time.Second // 1s, 2s, 4s
				select {
				case <-time.After(delay):
					continue
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}
		case resp.StatusCode == 404:
			return nil, &ErrNotFound{Resource: req.URL.Path}
		case resp.StatusCode == 401 || resp.StatusCode == 403:
			return nil, &ErrAuthentication{Message: string(body)}
		default:
			return nil, &ErrAPI{StatusCode: resp.StatusCode, Body: string(body)}
		}
	}

	return nil, lastErr
}
