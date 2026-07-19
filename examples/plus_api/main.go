package main

import (
	"context"
	"fmt"
	"log"
	"os"

	xon "github.com/XposedOrNot/XposedOrNot-Go"
)

func main() {
	apiKey := os.Getenv("XON_API_KEY")
	if apiKey == "" {
		log.Fatal("Set XON_API_KEY environment variable")
	}

	client, err := xon.NewClient(xon.WithAPIKey(apiKey))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	_, plus, err := client.CheckEmail(ctx, "test@example.com")
	if err != nil {
		log.Fatal(err)
	}
	if plus != nil {
		fmt.Printf("Status: %s\n", plus.Status)
		for _, b := range plus.Breaches {
			fmt.Printf("  Breach: %s (Domain: %s, Records: %d)\n", b.BreachID, b.Domain, b.XposedRecords)
		}
	}

	report, err := client.GetDomainBreaches(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Domain summary: %v\n", report.Metrics.DomainSummary)
	fmt.Printf("Yearly metrics: %v\n", report.Metrics.YearlyMetrics)
	for _, record := range report.Metrics.BreachesDetails {
		fmt.Printf("  %s (%s): %s\n", record.Email, record.Domain, record.Breach)
	}
}
