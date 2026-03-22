package xposedornot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

const passwordHashPrefixLength = 10

// CheckPassword checks whether the given password has been exposed in known
// data breaches. The password is hashed locally using the original Keccak-512
// algorithm, and only the first 10 characters of the hex digest are sent to
// the API, preserving privacy through k-anonymity.
func (c *Client) CheckPassword(ctx context.Context, password string) (*CheckPasswordResponse, error) {
	if password == "" {
		return nil, &ErrValidation{Field: "password", Message: "password is required"}
	}

	prefix := keccakHashPrefix(password, passwordHashPrefixLength)
	url := fmt.Sprintf("%s/v1/pass/anon/%s", c.passwordBaseURL, prefix)

	body, err := c.doRequest(ctx, "GET", url)
	if err != nil {
		// 404 means the password hash prefix was not found — password is not exposed
		var notFound *ErrNotFound
		if errors.As(err, &notFound) {
			return &CheckPasswordResponse{
				SearchPassAnon: PasswordAnonResult{
					Anon:  prefix,
					Count: "0",
				},
			}, nil
		}
		return nil, fmt.Errorf("check password: %w", err)
	}

	var resp CheckPasswordResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("check password: parsing response: %w", err)
	}
	return &resp, nil
}
