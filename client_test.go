package xposedornot

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClientDefaults(t *testing.T) {
	c, err := NewClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.baseURL != defaultBaseURL {
		t.Errorf("expected base URL %q, got %q", defaultBaseURL, c.baseURL)
	}
	if c.plusBaseURL != defaultPlusBaseURL {
		t.Errorf("expected plus base URL %q, got %q", defaultPlusBaseURL, c.plusBaseURL)
	}
	if c.passwordBaseURL != defaultPasswordBaseURL {
		t.Errorf("expected password base URL %q, got %q", defaultPasswordBaseURL, c.passwordBaseURL)
	}
	if c.maxRetries != defaultMaxRetries {
		t.Errorf("expected max retries %d, got %d", defaultMaxRetries, c.maxRetries)
	}
	if c.apiKey != "" {
		t.Errorf("expected empty API key, got %q", c.apiKey)
	}
}

func TestNewClientWithOptions(t *testing.T) {
	c, err := NewClient(
		WithAPIKey("test-key"),
		WithTimeout(10*time.Second),
		WithBaseURL("https://custom.api"),
		WithPlusBaseURL("https://custom.plus"),
		WithPasswordBaseURL("https://custom.pass"),
		WithMaxRetries(5),
		WithCustomHeaders(map[string]string{"X-Custom": "value"}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.apiKey != "test-key" {
		t.Errorf("expected API key %q, got %q", "test-key", c.apiKey)
	}
	if c.baseURL != "https://custom.api" {
		t.Errorf("expected base URL %q, got %q", "https://custom.api", c.baseURL)
	}
	if c.plusBaseURL != "https://custom.plus" {
		t.Errorf("expected plus base URL %q, got %q", "https://custom.plus", c.plusBaseURL)
	}
	if c.passwordBaseURL != "https://custom.pass" {
		t.Errorf("expected password base URL %q, got %q", "https://custom.pass", c.passwordBaseURL)
	}
	if c.maxRetries != 5 {
		t.Errorf("expected max retries %d, got %d", 5, c.maxRetries)
	}
	if c.httpClient.Timeout != 10*time.Second {
		t.Errorf("expected timeout %v, got %v", 10*time.Second, c.httpClient.Timeout)
	}
	if c.customHeaders["X-Custom"] != "value" {
		t.Errorf("expected custom header value %q, got %q", "value", c.customHeaders["X-Custom"])
	}
}

func TestNewClientRejectsInsecureURL(t *testing.T) {
	_, err := NewClient(WithBaseURL("http://insecure.api"))
	if err == nil {
		t.Fatal("expected error for insecure base URL, got nil")
	}
}

func TestNewClientAllowsInsecure(t *testing.T) {
	c, err := NewClient(
		WithBaseURL("http://insecure.api"),
		WithAllowInsecure(),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.baseURL != "http://insecure.api" {
		t.Errorf("expected base URL %q, got %q", "http://insecure.api", c.baseURL)
	}
}

func TestDoRequestSetsAPIKeyHeader(t *testing.T) {
	var gotKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("x-api-key")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c, err := NewClient(WithAPIKey("my-key"), WithBaseURL(srv.URL), WithAllowInsecure())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = c.doRequest(context.Background(), "GET", srv.URL+"/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotKey != "my-key" {
		t.Errorf("expected x-api-key header %q, got %q", "my-key", gotKey)
	}
}

func TestDoRequestCustomHeaders(t *testing.T) {
	var gotHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-Custom")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c, err := NewClient(
		WithAPIKey("key"), // skip rate limit
		WithCustomHeaders(map[string]string{"X-Custom": "test-value"}),
		WithAllowInsecure(),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = c.doRequest(context.Background(), "GET", srv.URL+"/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotHeader != "test-value" {
		t.Errorf("expected custom header %q, got %q", "test-value", gotHeader)
	}
}

func TestDoRequest404ReturnsErrNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c, err := NewClient(WithAPIKey("key"), WithAllowInsecure())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = c.doRequest(context.Background(), "GET", srv.URL+"/missing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *ErrNotFound
	if !errors.As(err, &notFound) {
		t.Errorf("expected ErrNotFound, got %T: %v", err, err)
	}
}

func TestDoRequest401ReturnsErrAuthentication(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized"))
	}))
	defer srv.Close()

	c, err := NewClient(WithAPIKey("bad-key"), WithAllowInsecure())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = c.doRequest(context.Background(), "GET", srv.URL+"/auth")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var authErr *ErrAuthentication
	if !errors.As(err, &authErr) {
		t.Errorf("expected ErrAuthentication, got %T: %v", err, err)
	}
}

func TestDoRequest403ReturnsErrAuthentication(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("forbidden"))
	}))
	defer srv.Close()

	c, err := NewClient(WithAPIKey("key"), WithAllowInsecure())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = c.doRequest(context.Background(), "GET", srv.URL+"/forbidden")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var authErr *ErrAuthentication
	if !errors.As(err, &authErr) {
		t.Errorf("expected ErrAuthentication, got %T: %v", err, err)
	}
}

func TestDoRequest500ReturnsErrAPI(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer srv.Close()

	c, err := NewClient(WithAPIKey("key"), WithAllowInsecure())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = c.doRequest(context.Background(), "GET", srv.URL+"/error")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *ErrAPI
	if !errors.As(err, &apiErr) {
		t.Errorf("expected ErrAPI, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 500 {
		t.Errorf("expected status 500, got %d", apiErr.StatusCode)
	}
}

func TestDoRequestRetryOn429(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("slow down"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c, err := NewClient(WithAPIKey("key"), WithMaxRetries(3), WithAllowInsecure())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	body, err := c.doRequest(context.Background(), "GET", srv.URL+"/retry")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(body) != `{"ok":true}` {
		t.Errorf("unexpected body: %s", body)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestDoRequest429ExhaustedRetries(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("rate limited"))
	}))
	defer srv.Close()

	c, err := NewClient(WithAPIKey("key"), WithMaxRetries(0), WithAllowInsecure())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = c.doRequest(context.Background(), "GET", srv.URL+"/limited")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var rlErr *ErrRateLimit
	if !errors.As(err, &rlErr) {
		t.Errorf("expected ErrRateLimit, got %T: %v", err, err)
	}
}

func TestDoRequestContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, err := NewClient(WithAPIKey("key"), WithAllowInsecure())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err = c.doRequest(ctx, "GET", srv.URL+"/slow")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRateLimitFreeAPI(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c, err := NewClient(WithBaseURL(srv.URL), WithAllowInsecure()) // no API key = rate limited
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	start := time.Now()
	for i := 0; i < 3; i++ {
		_, err := c.doRequest(context.Background(), "GET", srv.URL+"/test")
		if err != nil {
			t.Fatalf("request %d: unexpected error: %v", i, err)
		}
	}
	elapsed := time.Since(start)

	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
	// 3 calls with 1s spacing: first is immediate, then ~1s wait each
	// Should take at least ~2s
	if elapsed < 1900*time.Millisecond {
		t.Errorf("rate limiting too fast: 3 calls in %v, expected >=2s", elapsed)
	}
}

func TestRateLimitBypassedWithAPIKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c, err := NewClient(WithAPIKey("key"), WithBaseURL(srv.URL), WithAllowInsecure())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	start := time.Now()
	for i := 0; i < 5; i++ {
		_, err := c.doRequest(context.Background(), "GET", srv.URL+"/test")
		if err != nil {
			t.Fatalf("request %d: unexpected error: %v", i, err)
		}
	}
	elapsed := time.Since(start)

	// With API key, no rate limit, should be fast
	if elapsed > 2*time.Second {
		t.Errorf("API key requests too slow: 5 calls in %v", elapsed)
	}
}
