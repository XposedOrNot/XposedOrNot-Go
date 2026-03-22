package xposedornot

import "time"

// ClientOption is a functional option for configuring the Client.
type ClientOption func(*Client)

// WithAPIKey sets the API key for the Plus API.
// When set, requests to the Plus API include the x-api-key header
// and client-side rate limiting is disabled.
func WithAPIKey(key string) ClientOption {
	return func(c *Client) {
		c.apiKey = key
	}
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(d time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = d
	}
}

// WithBaseURL overrides the default free API base URL.
func WithBaseURL(url string) ClientOption {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithPlusBaseURL overrides the default Plus API base URL.
func WithPlusBaseURL(url string) ClientOption {
	return func(c *Client) {
		c.plusBaseURL = url
	}
}

// WithPasswordBaseURL overrides the default password API base URL.
func WithPasswordBaseURL(url string) ClientOption {
	return func(c *Client) {
		c.passwordBaseURL = url
	}
}

// WithMaxRetries sets the maximum number of retries on 429 responses.
// Default is 3.
func WithMaxRetries(n int) ClientOption {
	return func(c *Client) {
		c.maxRetries = n
	}
}

// WithCustomHeaders sets additional headers to include in every request.
func WithCustomHeaders(headers map[string]string) ClientOption {
	return func(c *Client) {
		c.customHeaders = headers
	}
}

// WithAllowInsecure disables the HTTPS enforcement check on base URLs.
// This is intended for testing purposes only.
func WithAllowInsecure() ClientOption {
	return func(c *Client) {
		c.allowInsecure = true
	}
}

// --- Response types ---

// CheckEmailFreeResponse is the response from the free check-email endpoint.
type CheckEmailFreeResponse struct {
	Breaches [][]string `json:"breaches"`
}

// PlusBreach represents a single breach entry from the Plus API.
type PlusBreach struct {
	BreachID     string `json:"breach_id"`
	BreachedDate string `json:"breached_date"`
	Logo         string `json:"logo"`
	PasswordRisk string `json:"password_risk"`
	Searchable   string `json:"searchable"`
	XposedData   string `json:"xposed_data"`
	XposedRecords int   `json:"xposed_records"`
	XposureDesc  string `json:"xposure_desc"`
	Domain       string `json:"domain"`
}

// CheckEmailPlusResponse is the response from the Plus API check-email endpoint.
type CheckEmailPlusResponse struct {
	Status   string       `json:"status"`
	Email    string       `json:"email"`
	Breaches []PlusBreach `json:"breaches"`
}

// ExposedBreach represents a single breach from the breaches listing endpoint.
type ExposedBreach struct {
	BreachID       string `json:"breachID"`
	BreachedDate   string `json:"breachedDate"`
	Domain         string `json:"domain"`
	Industry       string `json:"industry"`
	ExposedData    []string `json:"exposedData"`
	ExposedRecords int    `json:"exposedRecords"`
	Verified       bool   `json:"verified"`
	Logo           string `json:"logo"`
	References     string `json:"references"`
	Details        string `json:"details"`
	PasswordRisk   string `json:"password_risk"`
}

// GetBreachesResponse is the response from the breaches listing endpoint.
type GetBreachesResponse struct {
	ExposedBreaches []ExposedBreach `json:"exposedBreaches"`
}

// BreachDetail represents a single breach detail in breach analytics.
type BreachDetail struct {
	Breach       string `json:"breach"`
	Details      string `json:"details"`
	Domain       string `json:"domain"`
	Industry     string `json:"industry"`
	Logo         string `json:"logo"`
	PasswordRisk string `json:"password_risk"`
	References   string `json:"references"`
	Searchable   string `json:"searchable"`
	Verified     string `json:"verified"`
	XposedData   string `json:"xposed_data"`
	XposedDate   string `json:"xposed_date"`
	XposedRecords int   `json:"xposed_records"`
}

// BreachesSummary contains summary information about an email's breaches.
type BreachesSummary struct {
	Site              string `json:"site"`
	TotalBreaches     string `json:"total_breaches,omitempty"`
	MostRecentBreach  string `json:"most_recent_breach,omitempty"`
	FirstBreach       string `json:"first_breach,omitempty"`
	ExposedData       string `json:"exposed_data,omitempty"`
	PasswordRisk      string `json:"password_risk,omitempty"`
}

// BreachMetrics contains metrics derived from breach analytics.
type BreachMetrics struct {
	Industry       []interface{} `json:"industry"`
	PasswordStrength []interface{} `json:"passwords_strength"`
	YearlyBreaches []interface{} `json:"yearwise_details"`
	XposedData     []interface{} `json:"xposed_data"`
}

// PasteSummary contains paste-related summary information.
type PasteSummary struct {
	Count    int    `json:"cnt"`
	Domain   string `json:"domain"`
	FirstSeen string `json:"first_seen,omitempty"`
	LastSeen  string `json:"last_seen,omitempty"`
}

// ExposedBreachesWrapper wraps breach details in the analytics response.
type ExposedBreachesWrapper struct {
	BreachesDetails []BreachDetail `json:"breaches_details"`
}

// BreachAnalyticsResponse is the response from the breach-analytics endpoint.
type BreachAnalyticsResponse struct {
	ExposedBreaches ExposedBreachesWrapper `json:"ExposedBreaches"`
	BreachesSummary BreachesSummary        `json:"BreachesSummary"`
	BreachMetrics   BreachMetrics          `json:"BreachMetrics"`
	PastesSummary   PasteSummary           `json:"PastesSummary"`
	ExposedPastes   []interface{}          `json:"ExposedPastes"`
}

// PasswordAnonResult contains the anonymized password check result.
type PasswordAnonResult struct {
	Anon  string `json:"anon"`
	Char  string `json:"char"`
	Count string `json:"count"`
}

// CheckPasswordResponse is the response from the password check endpoint.
type CheckPasswordResponse struct {
	SearchPassAnon PasswordAnonResult `json:"SearchPassAnon"`
}
