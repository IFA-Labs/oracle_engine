package main

import (
	"fmt"

	"oracle_engine/internal/config"
)

func main() {
	cfg := config.Load()

	fmt.Println("Loaded subscription plans from config:")
	for name, plan := range cfg.SubscriptionPlans {
		fmt.Printf("- %s: %d req/month, %d hours rate limit, $%.2f/month\n", 
			name, plan.APIRequests, plan.RateLimit, plan.Price)
	}
}
