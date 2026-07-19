package xposedornot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// GetBreaches retrieves a list of known data breaches. If domain is non-empty,
// results are filtered to breaches affecting that domain.
func (c *Client) GetBreaches(ctx context.Context, domain string, opts ...RequestOption) (*GetBreachesResponse, error) {
	reqURL := fmt.Sprintf("%s/v1/breaches", c.baseURL)
	params := url.Values{}
	domain = strings.TrimSpace(domain)
	if domain != "" {
		params.Set("domain", domain)
	}
	for _, opt := range opts {
		opt(params)
	}
	if len(params) > 0 {
		reqURL += "?" + params.Encode()
	}

	body, err := c.doRequest(ctx, "GET", reqURL)
	if err != nil {
		return nil, fmt.Errorf("get breaches: %w", err)
	}

	var resp GetBreachesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("get breaches: parsing response: %w", err)
	}
	return &resp, nil
}

// GetDomainBreaches retrieves breach information for the domains verified
// against the configured API key. An API key with domains verified at the
// CXO dashboard (https://xposedornot.com/dashboard) is required.
func (c *Client) GetDomainBreaches(ctx context.Context) (*DomainBreachesResponse, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("get domain breaches: %w", &ErrAuthentication{
			Message: "an API key is required for domain breach monitoring; verify your domains at https://xposedornot.com/dashboard",
		})
	}

	reqURL := fmt.Sprintf("%s/v1/domain-breaches", c.baseURL)
	body, err := c.doRequest(ctx, "POST", reqURL)
	if err != nil {
		return nil, fmt.Errorf("get domain breaches: %w", err)
	}

	var resp DomainBreachesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("get domain breaches: parsing response: %w", err)
	}
	return &resp, nil
}
