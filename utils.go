package xposedornot

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/crypto/sha3"
)

// emailRegex is a basic email validation pattern.
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// validateEmail checks whether the given string is a valid email address.
// It returns an *ErrValidation if the email is invalid.
func validateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return &ErrValidation{Field: "email", Message: "email address is required"}
	}
	if !emailRegex.MatchString(email) {
		return &ErrValidation{Field: "email", Message: fmt.Sprintf("invalid email address: %s", email)}
	}
	return nil
}

// keccakHash computes the legacy Keccak-512 hash of the input string and
// returns the full hex-encoded digest. This uses the original Keccak algorithm,
// not the FIPS 202 SHA3-512 variant.
func keccakHash(input string) string {
	h := sha3.NewLegacyKeccak512()
	h.Write([]byte(input))
	return hex.EncodeToString(h.Sum(nil))
}

// keccakHashPrefix computes the legacy Keccak-512 hash and returns the first
// n characters of the hex-encoded digest.
func keccakHashPrefix(input string, n int) string {
	hash := keccakHash(input)
	if n > len(hash) {
		return hash
	}
	return hash[:n]
}
