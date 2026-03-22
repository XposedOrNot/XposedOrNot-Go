// Package xposedornot provides a Go client for the XposedOrNot API,
// which allows checking whether email addresses and passwords have been
// exposed in known data breaches.
//
// The client supports both the free API and the commercial Plus API.
// Use [NewClient] with functional options to configure the client:
//
//	// Free API usage
//	client, err := xposedornot.NewClient()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Plus API usage with API key
//	client, err := xposedornot.NewClient(
//	    xposedornot.WithAPIKey("your-api-key"),
//	    xposedornot.WithTimeout(30 * time.Second),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// All methods accept a [context.Context] for cancellation and timeout control.
//
// The free API is rate-limited to 1 request per second on the client side.
// Plus API users with an API key bypass this rate limit.
//
// The client automatically retries on HTTP 429 (Too Many Requests) responses
// with exponential backoff (1s, 2s, 4s) up to 3 retries by default.
//
// Error handling uses typed errors that can be checked with [errors.Is] and
// [errors.As]:
//
//	_, _, err := client.CheckEmail(ctx, "test@example.com")
//	var notFound *xposedornot.ErrNotFound
//	if errors.As(err, &notFound) {
//	    // email not found in any breaches
//	}
package xposedornot
