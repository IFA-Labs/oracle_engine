package weighted

import (
	"math"
	"oracle_engine/internal/logging"
	"oracle_engine/internal/models"
	"sort"

	"go.uber.org/zap"
)

func CalculateWeightedAveragePrice(
	id string,
	currPrice models.UnifiedPrice,
	pastXPrices []models.UnifiedPrice,
) models.Issuance {

	// Sort past prices by Timestamp descending (latest first)
	sort.Slice(pastXPrices, func(i, j int) bool {
		return pastXPrices[i].Timestamp.After(pastXPrices[j].Timestamp)
	})

	// Assign weights: highest for currPrice, descending for pastXPrices
	totalWeight := 0.0
	weightedSum := 0.0

	// Base weight for currPrice
	currWeight := float64(len(pastXPrices) + 1) // Highest weight
	totalWeight += currWeight
	weightedSum += currPrice.Value * currWeight

	// Assign descending weights to past prices
	for i, price := range pastXPrices {
		weight := float64(len(pastXPrices) - i)
		totalWeight += weight
		weightedSum += price.Value * weight
	}

	// Compute weighted average
	weightedAvg := weightedSum / totalWeight
	isDeviated := false

	// Check for deviation
	prices := append(pastXPrices, currPrice)
	mean := 0.0
	for _, p := range prices {
		mean += p.Value
	}
	mean /= float64(len(prices))

	deviationThreshold := 0.4 * mean
	isDeviated = math.Abs(weightedAvg-mean) > deviationThreshold
	state := models.Approved
	if isDeviated {
		state = models.Denied
	}

	modPrice := currPrice
	modPrice.Value = weightedAvg
	logging.Logger.Info("issuance value",
		zap.Any("val", modPrice.Value),
		zap.Any("exp", modPrice.Expo),
		zap.Any("nor", modPrice.Number()),
	)

	return models.Issuance{
		ID:    id,
		State: state,
		Price: modPrice,
	}
}
