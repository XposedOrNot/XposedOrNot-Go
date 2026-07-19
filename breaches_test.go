package xposedornot

import (
	"context"
	"errors"
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

func TestGetBreachesWithBreachID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/breaches" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("breach_id") != "Adobe" {
			t.Errorf("expected breach_id %q, got %q", "Adobe", r.URL.Query().Get("breach_id"))
		}
		if r.URL.Query().Get("domain") != "" {
			t.Errorf("unexpected domain param: %s", r.URL.Query().Get("domain"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"exposedBreaches": [{"breachID": "Adobe"}]}`))
	}))
	defer srv.Close()

	c, err := NewClient(WithAPIKey("key"), WithBaseURL(srv.URL), WithAllowInsecure())
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
	resp, err := c.GetBreaches(context.Background(), "", WithBreachID("Adobe"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.ExposedBreaches) != 1 {
		t.Fatalf("expected 1 breach, got %d", len(resp.ExposedBreaches))
	}
	if resp.ExposedBreaches[0].BreachID != "Adobe" {
		t.Errorf("expected breachID %q, got %q", "Adobe", resp.ExposedBreaches[0].BreachID)
	}
}

func TestGetDomainBreaches(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/domain-breaches" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("x-api-key") != "test-api-key" {
			t.Errorf("expected x-api-key header, got %q", r.Header.Get("x-api-key"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": "success",
			"metrics": {
				"Yearly_Metrics": {"2013": 1, "2012": 1},
				"Domain_Summary": {"example.com": 2},
				"Breach_Summary": {"Adobe": 1, "LinkedIn": 1},
				"Breaches_Details": [
					{"email": "alice@example.com", "domain": "example.com", "breach": "Adobe"},
					{"email": "bob@example.com", "domain": "example.com", "breach": "LinkedIn"}
				],
				"Top10_Breaches": {"Adobe": 152000000, "LinkedIn": 164000000},
				"Detailed_Breach_Info": {
					"Adobe": {"breached_date": "2013-10-04", "domain": "adobe.com"}
				}
			}
		}`))
	}))
	defer srv.Close()

	c, err := NewClient(WithAPIKey("test-api-key"), WithBaseURL(srv.URL), WithAllowInsecure())
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
	resp, err := c.GetDomainBreaches(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "success" {
		t.Errorf("expected status %q, got %q", "success", resp.Status)
	}
	if resp.Metrics.YearlyMetrics["2013"] != 1 {
		t.Errorf("expected yearly metric 1 for 2013, got %d", resp.Metrics.YearlyMetrics["2013"])
	}
	if resp.Metrics.DomainSummary["example.com"] != 2 {
		t.Errorf("expected domain summary 2, got %d", resp.Metrics.DomainSummary["example.com"])
	}
	if resp.Metrics.BreachSummary["Adobe"] != 1 {
		t.Errorf("expected breach summary 1 for Adobe, got %d", resp.Metrics.BreachSummary["Adobe"])
	}
	if resp.Metrics.Top10Breaches["Adobe"] != 152000000 {
		t.Errorf("expected top10 count 152000000, got %d", resp.Metrics.Top10Breaches["Adobe"])
	}
	if _, ok := resp.Metrics.DetailedBreachInfo["Adobe"]; !ok {
		t.Error("expected detailed breach info for Adobe")
	}
	if len(resp.Metrics.BreachesDetails) != 2 {
		t.Fatalf("expected 2 breach details, got %d", len(resp.Metrics.BreachesDetails))
	}
	first := resp.Metrics.BreachesDetails[0]
	if first.Email != "alice@example.com" || first.Domain != "example.com" || first.Breach != "Adobe" {
		t.Errorf("unexpected first breach detail: %+v", first)
	}
}

func TestGetDomainBreachesWithoutAPIKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("expected no HTTP request without an API key")
	}))
	defer srv.Close()

	c, err := NewClient(WithBaseURL(srv.URL), WithAllowInsecure())
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
	_, err = c.GetDomainBreaches(context.Background())
	var authErr *ErrAuthentication
	if !errors.As(err, &authErr) {
		t.Fatalf("expected ErrAuthentication, got %v", err)
	}
}

func TestGetDomainBreachesInvalidAPIKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"Error": "Unauthorized"}`))
	}))
	defer srv.Close()

	c, err := NewClient(WithAPIKey("invalid-key"), WithBaseURL(srv.URL), WithAllowInsecure())
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
	_, err = c.GetDomainBreaches(context.Background())
	var authErr *ErrAuthentication
	if !errors.As(err, &authErr) {
		t.Fatalf("expected ErrAuthentication, got %v", err)
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
