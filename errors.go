package xposedornot

import "fmt"

// ErrRateLimit is returned when the API returns a 429 Too Many Requests
// response after all retries have been exhausted.
type ErrRateLimit struct {
	Message string
}

func (e *ErrRateLimit) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("rate limit exceeded: %s", e.Message)
	}
	return "rate limit exceeded"
}

// ErrNotFound is returned when the requested resource is not found (404).
type ErrNotFound struct {
	Resource string
}

func (e *ErrNotFound) Error() string {
	if e.Resource != "" {
		return fmt.Sprintf("not found: %s", e.Resource)
	}
	return "not found"
}

// ErrAuthentication is returned when the API returns a 401 or 403 response,
// typically due to a missing or invalid API key.
type ErrAuthentication struct {
	Message string
}

func (e *ErrAuthentication) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("authentication failed: %s", e.Message)
	}
	return "authentication failed"
}

// ErrValidation is returned when input validation fails on the client side,
// such as an invalid email address format.
type ErrValidation struct {
	Field   string
	Message string
}

func (e *ErrValidation) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error on %s: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// ErrNetwork is returned when a network-level error occurs, such as a
// connection timeout or DNS resolution failure.
type ErrNetwork struct {
	Err error
}

func (e *ErrNetwork) Error() string {
	return fmt.Sprintf("network error: %v", e.Err)
}

func (e *ErrNetwork) Unwrap() error {
	return e.Err
}

// ErrAPI is returned when the API returns an unexpected error response
// that does not map to a more specific error type.
type ErrAPI struct {
	StatusCode int
	Body       string
}

func (e *ErrAPI) Error() string {
	return fmt.Sprintf("api error (status %d): %s", e.StatusCode, e.Body)
}
