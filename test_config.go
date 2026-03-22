//go:build ignore
// +build ignore

package main

import (
	"fmt"

	"oracle_engine/internal/config"
)

func main() {
	cfg := config.Load()

	fmt.Println("Loaded subscription plans from config:")
	for name, plan := range cfg.SubscriptionPlans {
		fmt.Printf("- %s: %d req/month, %d req/hour, %d req/day, $%.2f/month\n",
			name, plan.APIRequests, plan.RateLimitPerHour, plan.RateLimitPerDay, plan.Price)
	}
}
