package xposedornot

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetBreachesAll(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/breaches" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("domain") != "" {
			t.Errorf("unexpected domain param: %s", r.URL.Query().Get("domain"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"exposedBreaches": [{
				"breachID": "Adobe",
				"breachedDate": "2013-10-04",
				"domain": "adobe.com",
				"industry": "Technology",
				"exposedData": ["emails","passwords"],
				"exposedRecords": 153000000,
				"verified": true,
				"logo": "https://example.com/adobe.png",
				"password_risk": "high"
			}]
		}`))
	}))
	defer srv.Close()

	c, err := NewClient(WithAPIKey("key"), WithBaseURL(srv.URL), WithAllowInsecure())
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
	resp, err := c.GetBreaches(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected response, got nil")
	}
	if len(resp.ExposedBreaches) != 1 {
		t.Fatalf("expected 1 breach, got %d", len(resp.ExposedBreaches))
	}
	b := resp.ExposedBreaches[0]
	if b.BreachID != "Adobe" {
		t.Errorf("expected breachID %q, got %q", "Adobe", b.BreachID)
	}
	if b.ExposedRecords != 153000000 {
		t.Errorf("expected exposedRecords %d, got %d", 153000000, b.ExposedRecords)
	}
	if !b.Verified {
		t.Error("expected verified to be true")
	}
}

func TestGetBreachesWithDomain(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/breaches" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("domain") != "adobe.com" {
			t.Errorf("expected domain %q, got %q", "adobe.com", r.URL.Query().Get("domain"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"exposedBreaches": []}`))
	}))
	defer srv.Close()

	c, err := NewClient(WithAPIKey("key"), WithBaseURL(srv.URL), WithAllowInsecure())
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
	resp, err := c.GetBreaches(context.Background(), "adobe.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected response, got nil")
	}
}

func TestGetBreachesTrimsWhitespaceDomain(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("domain") != "adobe.com" {
			t.Errorf("expected trimmed domain %q, got %q", "adobe.com", r.URL.Query().Get("domain"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"exposedBreaches": []}`))
	}))
	defer srv.Close()

	c, err := NewClient(WithAPIKey("key"), WithBaseURL(srv.URL), WithAllowInsecure())
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
	_, err = c.GetBreaches(context.Background(), "  adobe.com  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetBreachesServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer srv.Close()

	c, err := NewClient(WithAPIKey("key"), WithBaseURL(srv.URL), WithAllowInsecure())
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
	_, err = c.GetBreaches(context.Background(), "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetBreachesEmptyDomainNoParam(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ensure no domain param when empty string
		if r.URL.RawQuery != "" {
			t.Errorf("expected no query params, got %q", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"exposedBreaches": []}`))
	}))
	defer srv.Close()

	c, err := NewClient(WithAPIKey("key"), WithBaseURL(srv.URL), WithAllowInsecure())
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
	_, err = c.GetBreaches(context.Background(), "   ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
