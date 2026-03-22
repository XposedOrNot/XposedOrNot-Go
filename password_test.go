package xposedornot

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCheckPassword(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify path starts with /v1/pass/anon/ and has a 10-char hex prefix
		if !strings.HasPrefix(r.URL.Path, "/v1/pass/anon/") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		prefix := strings.TrimPrefix(r.URL.Path, "/v1/pass/anon/")
		if len(prefix) != 10 {
			t.Errorf("expected 10-char hash prefix, got %d chars: %q", len(prefix), prefix)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"SearchPassAnon": {
				"anon": "` + prefix + `abc123",
				"char": "D:3;A:8;S:0;L:11",
				"count": "62703"
			}
		}`))
	}))
	defer srv.Close()

	c, err := NewClient(WithAPIKey("key"), WithPasswordBaseURL(srv.URL+"/api"), WithAllowInsecure())
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
	// Override password base URL to match test server
	c.passwordBaseURL = srv.URL

	resp, err := c.CheckPassword(context.Background(), "test-password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected response, got nil")
	}
	if resp.SearchPassAnon.Count != "62703" {
		t.Errorf("expected count %q, got %q", "62703", resp.SearchPassAnon.Count)
	}
	if resp.SearchPassAnon.Char != "D:3;A:8;S:0;L:11" {
		t.Errorf("expected char %q, got %q", "D:3;A:8;S:0;L:11", resp.SearchPassAnon.Char)
	}
}

func TestCheckPasswordEmptyPassword(t *testing.T) {
	c, err := NewClient(WithAPIKey("key"))
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
	_, err = c.CheckPassword(context.Background(), "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var valErr *ErrValidation
	if !errors.As(err, &valErr) {
		t.Errorf("expected ErrValidation, got %T: %v", err, err)
	}
}

func TestCheckPasswordHashPrefix(t *testing.T) {
	// Verify that the hash prefix sent to the server is the first 10 chars
	// of the Keccak-512 hash
	var receivedPrefix string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPrefix = strings.TrimPrefix(r.URL.Path, "/v1/pass/anon/")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"SearchPassAnon":{"anon":"","char":"","count":"0"}}`))
	}))
	defer srv.Close()

	c, err := NewClient(WithAPIKey("key"), WithAllowInsecure())
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
	c.passwordBaseURL = srv.URL

	password := "hello123"
	_, err = c.CheckPassword(context.Background(), password)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedPrefix := keccakHashPrefix(password, 10)
	if receivedPrefix != expectedPrefix {
		t.Errorf("expected prefix %q, got %q", expectedPrefix, receivedPrefix)
	}
}

func TestCheckPasswordNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c, err := NewClient(WithAPIKey("key"), WithAllowInsecure())
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
	c.passwordBaseURL = srv.URL

	_, err = c.CheckPassword(context.Background(), "unique-password")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *ErrNotFound
	if !errors.As(err, &notFound) {
		t.Errorf("expected ErrNotFound, got %T: %v", err, err)
	}
}

func TestCheckPasswordServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error"))
	}))
	defer srv.Close()

	c, err := NewClient(WithAPIKey("key"), WithAllowInsecure())
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
	c.passwordBaseURL = srv.URL

	_, err = c.CheckPassword(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
