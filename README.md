<p align="center">
  <a href="https://xposedornot.com">
    <img src="https://xposedornot.com/static/logos/xon.png" alt="XposedOrNot" width="200">
  </a>
</p>

<h1 align="center">xposedornot-go</h1>

<p align="center">
  Official Go SDK for the <a href="https://xposedornot.com">XposedOrNot</a> API<br>
  <em>Check if your email or password has been exposed in data breaches</em>
</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/XposedOrNot/XposedOrNot-Go"><img src="https://pkg.go.dev/badge/github.com/XposedOrNot/XposedOrNot-Go.svg" alt="Go Reference"></a>
  <a href="https://github.com/XposedOrNot/XposedOrNot-Go/actions"><img src="https://img.shields.io/github/actions/workflow/status/XposedOrNot/XposedOrNot-Go/build.yml?branch=main" alt="Build Status"></a>
  <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License: MIT"></a>
  <a href="https://go.dev"><img src="https://img.shields.io/badge/Go-1.21+-00ADD8.svg" alt="Go Version"></a>
</p>

---

> **Note:** This SDK uses the free public API from [XposedOrNot.com](https://xposedornot.com) - a free service to check if your email has been compromised in data breaches. A Plus API with enhanced features is available with an API key. Visit the [XposedOrNot website](https://xposedornot.com) to learn more.

---

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Requirements](#requirements)
- [Quick Start](#quick-start)
- [API Reference](#api-reference)
  - [CheckEmail](#checkemailctx-email)
  - [GetBreaches](#getbreachesctx-domain)
  - [BreachAnalytics](#breachanalyticsctx-email)
  - [CheckPassword](#checkpasswordctx-password)
- [Error Handling](#error-handling)
- [Rate Limits](#rate-limits)
- [Configuration Options](#configuration-options)
- [Contributing](#contributing)
- [License](#license)
- [Links](#links)

---

## Features

- **Simple API** - Idiomatic Go client with functional options pattern
- **Context Support** - All methods accept `context.Context` for cancellation and timeouts
- **Free and Plus APIs** - Use the free API or unlock detailed results with an API key
- **Detailed Analytics** - Get breach details, risk scores, metrics, and paste exposures
- **Error Handling** - Typed error values compatible with `errors.As`
- **Password Privacy** - Passwords are hashed locally; only a short prefix is sent (k-anonymity)
- **Rate Limiting** - Built-in client-side rate limiting with automatic retry on 429 responses
- **Secure** - HTTPS enforced on all base URLs by default

## Installation

```bash
go get github.com/XposedOrNot/XposedOrNot-Go
```

## Requirements

- Go 1.21 or higher

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"

	xon "github.com/XposedOrNot/XposedOrNot-Go"
)

func main() {
	client, err := xon.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	// Check if an email has been breached
	free, _, err := client.CheckEmail(context.Background(), "test@example.com")
	if err != nil {
		log.Fatal(err)
	}

	if len(free.Breaches) > 0 {
		fmt.Printf("Email found in %d breaches:\n", len(free.Breaches))
		for _, b := range free.Breaches {
			fmt.Printf("  - %s\n", b)
		}
	} else {
		fmt.Println("Good news! Email not found in any known breaches.")
	}
}
```

## API Reference

### Constructor

```go
client, err := xon.NewClient(opts ...xon.ClientOption)
```

Create a new client with functional options. See [Configuration Options](#configuration-options) for all available options.

### Methods

#### `CheckEmail(ctx, email)`

Check if an email address has been exposed in any data breaches. Returns the free API response or the Plus API response depending on whether an API key is configured.

```go
// Free API (no API key)
free, _, err := client.CheckEmail(ctx, "user@example.com")
if err != nil {
	log.Fatal(err)
}
fmt.Println("Breaches:", free.Breaches)
```

```go
// Plus API (with API key)
client, _ := xon.NewClient(xon.WithAPIKey("your-api-key"))

_, plus, err := client.CheckEmail(ctx, "user@example.com")
if err != nil {
	log.Fatal(err)
}
for _, b := range plus.Breaches {
	fmt.Printf("Breach: %s, Date: %s, Records: %d\n",
		b.BreachID, b.BreachedDate, b.XposedRecords)
}
```

**Signature:**

```go
func (c *Client) CheckEmail(ctx context.Context, email string) (*CheckEmailFreeResponse, *CheckEmailPlusResponse, error)
```

| Return Value | Type | Description |
|---|---|---|
| `free` | `*CheckEmailFreeResponse` | Populated when no API key is set (Plus response is `nil`) |
| `plus` | `*CheckEmailPlusResponse` | Populated when an API key is set (Free response is `nil`) |
| `err` | `error` | Non-nil on validation failure, network error, or API error |

---

#### `GetBreaches(ctx, domain)`

Retrieve a list of all known data breaches. Pass a non-empty `domain` string to filter results by domain.

```go
// Get all breaches
all, err := client.GetBreaches(ctx, "")
if err != nil {
	log.Fatal(err)
}
fmt.Println("Total known breaches:", len(all.ExposedBreaches))

// Filter by domain
adobe, err := client.GetBreaches(ctx, "adobe.com")
if err != nil {
	log.Fatal(err)
}
fmt.Println("Adobe breaches:", len(adobe.ExposedBreaches))
```

**Signature:**

```go
func (c *Client) GetBreaches(ctx context.Context, domain string) (*GetBreachesResponse, error)
```

| Parameter | Type | Description |
|---|---|---|
| `ctx` | `context.Context` | Context for cancellation and timeouts |
| `domain` | `string` | Filter by domain (pass `""` for all breaches) |

**Returns:** `*GetBreachesResponse` containing a slice of `ExposedBreach` with fields including `BreachID`, `BreachedDate`, `Domain`, `Industry`, `ExposedData`, `ExposedRecords`, and `Verified`.

---

#### `BreachAnalytics(ctx, email)`

Get detailed breach analytics for an email address, including breach details, summary statistics, metrics breakdowns, and paste exposures.

```go
analytics, err := client.BreachAnalytics(ctx, "user@example.com")
if err != nil {
	log.Fatal(err)
}

fmt.Println("Breach details:", len(analytics.ExposedBreaches.BreachesDetails))
fmt.Println("Site:", analytics.BreachesSummary.Site)
fmt.Println("Total breaches:", analytics.BreachesSummary.TotalBreaches)
fmt.Println("Most recent:", analytics.BreachesSummary.MostRecentBreach)
```

**Signature:**

```go
func (c *Client) BreachAnalytics(ctx context.Context, email string) (*BreachAnalyticsResponse, error)
```

| Parameter | Type | Description |
|---|---|---|
| `ctx` | `context.Context` | Context for cancellation and timeouts |
| `email` | `string` | Email address to analyze |

**Returns:** `*BreachAnalyticsResponse` with the following fields:

| Field | Type | Description |
|---|---|---|
| `ExposedBreaches` | `ExposedBreachesWrapper` | Detailed list of each breach |
| `BreachesSummary` | `BreachesSummary` | Summary stats (total breaches, date range, risk) |
| `BreachMetrics` | `BreachMetrics` | Breakdowns by industry, password strength, year, and data type |
| `PastesSummary` | `PasteSummary` | Paste exposure summary |
| `ExposedPastes` | `[]interface{}` | Raw paste exposure entries |

---

#### `CheckPassword(ctx, password)`

Check whether a password has appeared in known data breaches. The password is hashed locally using Keccak-512 and only the first 10 characters of the hex digest are sent to the API. This **k-anonymity** approach ensures the full password never leaves your machine.

```go
resp, err := client.CheckPassword(ctx, "password123")
if err != nil {
	log.Fatal(err)
}

if resp.SearchPassAnon.Count != "0" {
	fmt.Printf("Password has been exposed %s time(s)\n", resp.SearchPassAnon.Count)
} else {
	fmt.Println("Password not found in known breaches")
}
```

**Signature:**

```go
func (c *Client) CheckPassword(ctx context.Context, password string) (*CheckPasswordResponse, error)
```

| Parameter | Type | Description |
|---|---|---|
| `ctx` | `context.Context` | Context for cancellation and timeouts |
| `password` | `string` | The plaintext password to check (hashed locally before sending) |

## Error Handling

All errors returned by the client are typed. Use `errors.As` to check for specific error conditions:

```go
import "errors"

free, _, err := client.CheckEmail(ctx, "test@example.com")
if err != nil {
	var notFound *xon.ErrNotFound
	var validationErr *xon.ErrValidation
	var rateLimitErr *xon.ErrRateLimit
	var authErr *xon.ErrAuthentication
	var networkErr *xon.ErrNetwork
	var apiErr *xon.ErrAPI

	switch {
	case errors.As(err, &notFound):
		fmt.Println("Email not found in any breaches")
	case errors.As(err, &validationErr):
		fmt.Printf("Invalid input: %s\n", validationErr.Message)
	case errors.As(err, &rateLimitErr):
		fmt.Println("Rate limit exceeded, try again later")
	case errors.As(err, &authErr):
		fmt.Println("Invalid API key")
	case errors.As(err, &networkErr):
		fmt.Printf("Network error: %v\n", networkErr.Unwrap())
	case errors.As(err, &apiErr):
		fmt.Printf("API error (status %d): %s\n", apiErr.StatusCode, apiErr.Body)
	default:
		fmt.Printf("Unexpected error: %v\n", err)
	}
}
```

### Error Types

| Error Type | Description |
|---|---|
| `*ErrValidation` | Invalid input (e.g., malformed email address) |
| `*ErrNotFound` | Requested resource not found (HTTP 404) |
| `*ErrRateLimit` | API rate limit exceeded after all retries (HTTP 429) |
| `*ErrAuthentication` | Missing or invalid API key (HTTP 401/403) |
| `*ErrNetwork` | Network-level failure (DNS, connection, timeout) |
| `*ErrAPI` | Unexpected API error with status code and body |

## Rate Limits

| | Free API | Plus API (with API key) |
|---|---|---|
| **Client-side limiting** | 1 request per second (automatic) | None |
| **Server-side limiting** | Standard limits apply | Higher limits |
| **429 retry behavior** | Exponential backoff (1s, 2s, 4s) | Exponential backoff (1s, 2s, 4s) |

The client automatically retries on HTTP 429 responses with exponential backoff, up to the configured maximum retries (default: 3).

## Configuration Options

All options are passed to `NewClient` as functional options:

```go
client, err := xon.NewClient(
	xon.WithAPIKey("your-api-key"),
	xon.WithTimeout(15 * time.Second),
	xon.WithMaxRetries(5),
)
```

| Option | Default | Description |
|---|---|---|
| `WithAPIKey(key)` | `""` | Set API key for Plus API access |
| `WithTimeout(d)` | `30s` | HTTP client timeout |
| `WithBaseURL(url)` | `https://api.xposedornot.com` | Override the free API base URL |
| `WithPlusBaseURL(url)` | `https://plus-api.xposedornot.com` | Override the Plus API base URL |
| `WithPasswordBaseURL(url)` | `https://passwords.xposedornot.com/api` | Override the password API base URL |
| `WithMaxRetries(n)` | `3` | Maximum retries on HTTP 429 responses |
| `WithCustomHeaders(h)` | `nil` | Additional headers included in every request |
| `WithAllowInsecure()` | `false` | Disable HTTPS enforcement (testing only) |

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Setup

```bash
# Clone the repository
git clone https://github.com/XposedOrNot/XposedOrNot-Go.git
cd XposedOrNot-Go

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detector
go test -race ./...

# Vet
go vet ./...
```

## License

MIT - see the [LICENSE](LICENSE) file for details.

## Links

- [XposedOrNot Website](https://xposedornot.com)
- [API Documentation](https://xposedornot.com/api_doc)
- [GitHub Repository](https://github.com/XposedOrNot/XposedOrNot-Go)
- [Go Package Reference](https://pkg.go.dev/github.com/XposedOrNot/XposedOrNot-Go)
- [XposedOrNot API Repository](https://github.com/XposedOrNot/XposedOrNot-API)

---

<p align="center">
  Made with care by <a href="https://xposedornot.com">XposedOrNot</a>
</p>
