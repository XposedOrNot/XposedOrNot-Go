package xposedornot

import (
	"errors"
	"testing"
)

func TestValidateEmailValid(t *testing.T) {
	validEmails := []string{
		"test@example.com",
		"user.name@domain.org",
		"user+tag@example.co.uk",
		"test123@sub.domain.com",
		"a@b.co",
	}
	for _, email := range validEmails {
		if err := validateEmail(email); err != nil {
			t.Errorf("expected %q to be valid, got error: %v", email, err)
		}
	}
}

func TestValidateEmailInvalid(t *testing.T) {
	invalidEmails := []string{
		"",
		"   ",
		"notanemail",
		"@example.com",
		"user@",
		"user@.com",
		"user@domain",
		"user @example.com",
	}
	for _, email := range invalidEmails {
		err := validateEmail(email)
		if err == nil {
			t.Errorf("expected %q to be invalid, got nil error", email)
			continue
		}
		var valErr *ErrValidation
		if !errors.As(err, &valErr) {
			t.Errorf("expected ErrValidation for %q, got %T: %v", email, err, err)
		}
	}
}

func TestValidateEmailTrimsWhitespace(t *testing.T) {
	if err := validateEmail("  test@example.com  "); err != nil {
		t.Errorf("expected trimmed email to be valid, got: %v", err)
	}
}

func TestKeccakHash(t *testing.T) {
	// Known Keccak-512 test vector: empty string
	// Keccak-512("") = 0eab42de4c3ceb9235fc91acffe746b29c29a8c366b7c60e4e67c466f36a4304
	//                   c00fa9caf9d87976ba469bcbe06713b435f091ef2769fb160cdab33d3670680e
	hash := keccakHash("")
	expected := "0eab42de4c3ceb9235fc91acffe746b29c29a8c366b7c60e4e67c466f36a4304c00fa9caf9d87976ba469bcbe06713b435f091ef2769fb160cdab33d3670680e"
	if hash != expected {
		t.Errorf("keccak hash of empty string:\n  got:    %s\n  expect: %s", hash, expected)
	}
}

func TestKeccakHashNonEmpty(t *testing.T) {
	hash := keccakHash("hello")
	// Keccak-512 hash of "hello" should be 128 hex chars
	if len(hash) != 128 {
		t.Errorf("expected 128 hex chars, got %d", len(hash))
	}
	// Should not be all zeros
	allZero := true
	for _, c := range hash {
		if c != '0' {
			allZero = false
			break
		}
	}
	if allZero {
		t.Error("hash should not be all zeros")
	}
}

func TestKeccakHashPrefix(t *testing.T) {
	hash := keccakHash("test")
	prefix := keccakHashPrefix("test", 10)
	if len(prefix) != 10 {
		t.Errorf("expected prefix length 10, got %d", len(prefix))
	}
	if prefix != hash[:10] {
		t.Errorf("prefix %q does not match first 10 chars of hash %q", prefix, hash[:10])
	}
}

func TestKeccakHashPrefixLongerThanHash(t *testing.T) {
	hash := keccakHash("x")
	prefix := keccakHashPrefix("x", 999)
	if prefix != hash {
		t.Error("prefix longer than hash should return full hash")
	}
}

func TestKeccakHashDeterministic(t *testing.T) {
	h1 := keccakHash("password123")
	h2 := keccakHash("password123")
	if h1 != h2 {
		t.Error("same input should produce same hash")
	}
}

func TestKeccakHashDifferentInputs(t *testing.T) {
	h1 := keccakHash("password1")
	h2 := keccakHash("password2")
	if h1 == h2 {
		t.Error("different inputs should produce different hashes")
	}
}

func TestErrorMessages(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"ErrRateLimit with message", &ErrRateLimit{Message: "too fast"}, "rate limit exceeded: too fast"},
		{"ErrRateLimit empty", &ErrRateLimit{}, "rate limit exceeded"},
		{"ErrNotFound with resource", &ErrNotFound{Resource: "/v1/test"}, "not found: /v1/test"},
		{"ErrNotFound empty", &ErrNotFound{}, "not found"},
		{"ErrAuthentication with message", &ErrAuthentication{Message: "bad key"}, "authentication failed: bad key"},
		{"ErrAuthentication empty", &ErrAuthentication{}, "authentication failed"},
		{"ErrValidation with field", &ErrValidation{Field: "email", Message: "invalid"}, "validation error on email: invalid"},
		{"ErrValidation no field", &ErrValidation{Message: "bad input"}, "validation error: bad input"},
		{"ErrAPI", &ErrAPI{StatusCode: 500, Body: "error"}, "api error (status 500): error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestErrNetworkUnwrap(t *testing.T) {
	inner := errors.New("connection refused")
	err := &ErrNetwork{Err: inner}
	if !errors.Is(err, inner) {
		t.Error("ErrNetwork should unwrap to inner error")
	}
}
