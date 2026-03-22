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
func (c *Client) GetBreaches(ctx context.Context, domain string) (*GetBreachesResponse, error) {
	reqURL := fmt.Sprintf("%s/v1/breaches", c.baseURL)
	domain = strings.TrimSpace(domain)
	if domain != "" {
		params := url.Values{"domain": {domain}}
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
