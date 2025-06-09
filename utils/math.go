package utils

import "math"

const Scale = 1e-7

func RoundFloat(f float64) float64 {
	return math.Round(f*1e7) / 1e7
}

func RoundFloats(fs []float64) []float64 {
	result := make([]float64, len(fs))
	for i, price := range fs {
		result[i] = math.Round(price*1e7) / 1e7
	}
	return result
}
