package common

import (
	"fmt"
)

// CompareKeys compares two string keys with type-aware comparison
// It handles numbers, booleans, and strings appropriately
func CompareKeys(a, b string) bool {
	// Try to compare as numbers first
	var numA, numB float64
	_, errA := fmt.Sscanf(a, "%f", &numA)
	_, errB := fmt.Sscanf(b, "%f", &numB)

	if errA == nil && errB == nil {
		// Both are numbers
		return numA < numB
	}

	// Try to compare as booleans
	if (a == "true" || a == "false") && (b == "true" || b == "false") {
		// false < true
		return a == "false" && b == "true"
	}

	// Default to string comparison
	return a < b
}