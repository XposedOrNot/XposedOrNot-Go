package xposedornot

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestCheckEmailFree(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expected := "/v1/check-email/" + url.PathEscape("test@example.com")
		if r.URL.RawPath != "" {
			if r.URL.RawPath != expected {
				t.Errorf("unexpected raw path: %s, expected %s", r.URL.RawPath, expected)
			}
		} else if r.URL.Path != "/v1/check-email/test@example.com" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"breaches":[["Adobe","LinkedIn"]]}`))
	}))
	defer srv.Close()

	c, err := NewClient(WithBaseURL(srv.URL), WithAllowInsecure())
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
	c.apiKey = "" // ensure free API

	free, plus, err := c.CheckEmail(context.Background(), "test@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plus != nil {
		t.Error("expected plus response to be nil for free API")
	}
	if free == nil {
		t.Fatal("expected free response, got nil")
	}
	if len(free.Breaches) != 1 {
		t.Fatalf("expected 1 breach group, got %d", len(free.Breaches))
	}
	if len(free.Breaches[0]) != 2 {
		t.Fatalf("expected 2 breaches, got %d", len(free.Breaches[0]))
	}
	if free.Breaches[0][0] != "Adobe" {
		t.Errorf("expected first breach %q, got %q", "Adobe", free.Breaches[0][0])
	}
}

func TestCheckEmailPlus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expected := "/v3/check-email/" + url.PathEscape("test@example.com")
		if r.URL.RawPath != "" {
			if r.URL.RawPath != expected {
				t.Errorf("unexpected raw path: %s, expected %s", r.URL.RawPath, expected)
			}
		} else if r.URL.Path != "/v3/check-email/test@example.com" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("detailed") != "true" {
			t.Error("expected detailed=true query param")
		}
		if r.Header.Get("x-api-key") != "test-key" {
			t.Errorf("expected x-api-key header, got %q", r.Header.Get("x-api-key"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": "success",
			"email": "test@example.com",
			"breaches": [{
				"breach_id": "Adobe",
				"breached_date": "2013-10-04",
				"logo": "https://example.com/adobe.png",
				"password_risk": "high",
				"searchable": "yes",
				"xposed_data": "email,password",
				"xposed_records": 153000000,
				"xposure_desc": "Adobe breach",
				"domain": "adobe.com"
			}]
		}`))
	}))
	defer srv.Close()

	c, err := NewClient(WithAPIKey("test-key"), WithPlusBaseURL(srv.URL), WithAllowInsecure())
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
	free, plus, err := c.CheckEmail(context.Background(), "test@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if free != nil {
		t.Error("expected free response to be nil for plus API")
	}
	if plus == nil {
		t.Fatal("expected plus response, got nil")
	}
	if plus.Status != "success" {
		t.Errorf("expected status %q, got %q", "success", plus.Status)
	}
	if plus.Email != "test@example.com" {
		t.Errorf("expected email %q, got %q", "test@example.com", plus.Email)
	}
	if len(plus.Breaches) != 1 {
		t.Fatalf("expected 1 breach, got %d", len(plus.Breaches))
	}
	b := plus.Breaches[0]
	if b.BreachID != "Adobe" {
		t.Errorf("expected breach_id %q, got %q", "Adobe", b.BreachID)
	}
	if b.XposedRecords != 153000000 {
		t.Errorf("expected xposed_records %d, got %d", 153000000, b.XposedRecords)
	}
}

func TestCheckEmailInvalidEmail(t *testing.T) {
	c, err := NewClient(WithAPIKey("key"))
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
	_, _, err = c.CheckEmail(context.Background(), "not-an-email")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var valErr *ErrValidation
	if !errors.As(err, &valErr) {
		t.Errorf("expected ErrValidation, got %T: %v", err, err)
	}
}

func TestCheckEmailEmptyEmail(t *testing.T) {
	c, err := NewClient(WithAPIKey("key"))
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
	_, _, err = c.CheckEmail(context.Background(), "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var valErr *ErrValidation
	if !errors.As(err, &valErr) {
		t.Errorf("expected ErrValidation, got %T: %v", err, err)
	}
}

func TestCheckEmailTrimsWhitespace(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expected := "/v1/check-email/" + url.PathEscape("test@example.com")
		if r.URL.RawPath != "" {
			if r.URL.RawPath != expected {
				t.Errorf("unexpected raw path (not trimmed): %s", r.URL.RawPath)
			}
		} else if r.URL.Path != "/v1/check-email/test@example.com" {
			t.Errorf("unexpected path (not trimmed): %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"breaches":[]}`))
	}))
	defer srv.Close()

	c, err := NewClient(WithBaseURL(srv.URL), WithAllowInsecure())
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
	_, _, err = c.CheckEmail(context.Background(), "  test@example.com  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBreachAnalytics(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/breach-analytics" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("email") != "test@example.com" {
			t.Errorf("unexpected email param: %s", r.URL.Query().Get("email"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"ExposedBreaches": {
				"breaches_details": [{
					"breach": "Adobe",
					"domain": "adobe.com",
					"industry": "Technology",
					"xposed_records": 153000000
				}]
			},
			"BreachesSummary": {
				"site": "test@example.com"
			},
			"BreachMetrics": {
				"industry": [],
				"passwords_strength": [],
				"yearwise_details": [],
				"xposed_data": []
			},
			"PastesSummary": {
				"cnt": 0,
				"domain": ""
			},
			"ExposedPastes": []
		}`))
	}))
	defer srv.Close()

	c, err := NewClient(WithBaseURL(srv.URL), WithAllowInsecure())
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
	resp, err := c.BreachAnalytics(context.Background(), "test@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected response, got nil")
	}
	if len(resp.ExposedBreaches.BreachesDetails) != 1 {
		t.Fatalf("expected 1 breach detail, got %d", len(resp.ExposedBreaches.BreachesDetails))
	}
	if resp.ExposedBreaches.BreachesDetails[0].Breach != "Adobe" {
		t.Errorf("expected breach %q, got %q", "Adobe", resp.ExposedBreaches.BreachesDetails[0].Breach)
	}
	if resp.BreachesSummary.Site != "test@example.com" {
		t.Errorf("expected site %q, got %q", "test@example.com", resp.BreachesSummary.Site)
	}
}

func TestBreachAnalyticsInvalidEmail(t *testing.T) {
	c, err := NewClient(WithAPIKey("key"))
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
	_, err = c.BreachAnalytics(context.Background(), "invalid")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var valErr *ErrValidation
	if !errors.As(err, &valErr) {
		t.Errorf("expected ErrValidation, got %T: %v", err, err)
	}
}

func TestCheckEmailNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c, err := NewClient(WithBaseURL(srv.URL), WithAllowInsecure())
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
	_, _, err = c.CheckEmail(context.Background(), "clean@example.com")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *ErrNotFound
	if !errors.As(err, &notFound) {
		t.Errorf("expected ErrNotFound, got %T: %v", err, err)
	}
}
