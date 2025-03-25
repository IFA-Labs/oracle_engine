package outlier

import (
    "math"
    "oracle_engine/internal/models"
    "sort"
)

func FilterOutliers(prices []models.Price) []models.Price {
    if len(prices) < 3 { // Need enough data for meaningful filtering
        return prices
    }

    // Extract values
    values := make([]float64, len(prices))
    for i, p := range prices {
        values[i] = p.Value
    }

    // Calculate median
    sort.Float64s(values)
    median := values[len(values)/2]

    // Filter: Keep prices within 10% of median (configurable later)
    var filtered []models.Price
    for _, p := range prices {
        if math.Abs(p.Value-median)/median <= 0.1 {
            filtered = append(filtered, p)
        }
    }
    return filtered
}