// Package exponential provides functions for exponential backoff.
package exponential

import (
	"math"
)

// GetNth calculates the exponential backoff for given base and exponent.
func GetNth(n uint, b uint) uint {
	return uint(math.Pow(float64(b), float64(n)))
}
