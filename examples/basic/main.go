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

	// Check if an email has been exposed
	ctx := context.Background()
	free, _, err := client.CheckEmail(ctx, "test@example.com")
	if err != nil {
		log.Fatal(err)
	}
	if free != nil {
		fmt.Printf("Found in %d breaches\n", len(free.Breaches))
	}

	// Get all known breaches
	breaches, err := client.GetBreaches(ctx, "")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Total known breaches: %d\n", len(breaches.ExposedBreaches))

	// Check a password (hashed locally, never sent in clear text)
	passResult, err := client.CheckPassword(ctx, "password123")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Password found %s times in breaches\n", passResult.SearchPassAnon.Count)
}
