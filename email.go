package xposedornot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// CheckEmail checks whether the given email address has been exposed in any
// known data breaches. If the client has an API key configured, it uses the
// Plus API and returns a detailed response; otherwise it uses the free API.
//
// For the free API, only CheckEmailFreeResponse is populated (second return value is nil).
// For the Plus API, only CheckEmailPlusResponse is populated (first return value is nil).
func (c *Client) CheckEmail(ctx context.Context, email string) (*CheckEmailFreeResponse, *CheckEmailPlusResponse, error) {
	email = strings.TrimSpace(email)
	if err := validateEmail(email); err != nil {
		return nil, nil, fmt.Errorf("check email: %w", err)
	}

	if c.apiKey != "" {
		return c.checkEmailPlus(ctx, email)
	}
	return c.checkEmailFree(ctx, email)
}

func (c *Client) checkEmailFree(ctx context.Context, email string) (*CheckEmailFreeResponse, *CheckEmailPlusResponse, error) {
	reqURL := fmt.Sprintf("%s/v1/check-email/%s", c.baseURL, url.PathEscape(email))
	body, err := c.doRequest(ctx, "GET", reqURL)
	if err != nil {
		// 404 means email not found in any breaches — this is a valid result
		var notFound *ErrNotFound
		if errors.As(err, &notFound) {
			return &CheckEmailFreeResponse{}, nil, nil
		}
		return nil, nil, fmt.Errorf("check email (free): %w", err)
	}

	var resp CheckEmailFreeResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, nil, fmt.Errorf("check email (free): parsing response: %w", err)
	}
	return &resp, nil, nil
}

func (c *Client) checkEmailPlus(ctx context.Context, email string) (*CheckEmailFreeResponse, *CheckEmailPlusResponse, error) {
	params := url.Values{"detailed": {"true"}}
	reqURL := fmt.Sprintf("%s/v3/check-email/%s?%s", c.plusBaseURL, url.PathEscape(email), params.Encode())
	body, err := c.doRequest(ctx, "GET", reqURL)
	if err != nil {
		return nil, nil, fmt.Errorf("check email (plus): %w", err)
	}

	var resp CheckEmailPlusResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, nil, fmt.Errorf("check email (plus): parsing response: %w", err)
	}
	return nil, &resp, nil
}

// BreachAnalytics retrieves detailed breach analytics for the given email
// address, including breach details, summary, metrics, and paste information.
func (c *Client) BreachAnalytics(ctx context.Context, email string) (*BreachAnalyticsResponse, error) {
	email = strings.TrimSpace(email)
	if err := validateEmail(email); err != nil {
		return nil, fmt.Errorf("breach analytics: %w", err)
	}

	params := url.Values{"email": {email}}
	reqURL := fmt.Sprintf("%s/v1/breach-analytics?%s", c.baseURL, params.Encode())
	body, err := c.doRequest(ctx, "GET", reqURL)
	if err != nil {
		return nil, fmt.Errorf("breach analytics: %w", err)
	}

	var resp BreachAnalyticsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("breach analytics: parsing response: %w", err)
	}
	return &resp, nil
}
