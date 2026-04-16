package services

import "math"

const (
	screeningPercentileFloor = 0.5
	screeningPercentileCeil  = 99.5
)

type ScreeningIQProfile struct {
	EstimatedIQ         int    `json:"estimatedIq"`
	ClassificationLabel string `json:"classificationLabel"`
}

func BuildScreeningIQProfile(percentage float64) ScreeningIQProfile {
	estimatedIQ := EstimateScreeningIQ(percentage)
	return ScreeningIQProfile{
		EstimatedIQ:         estimatedIQ,
		ClassificationLabel: ClassifyEstimatedIQ(estimatedIQ),
	}
}

func EstimateScreeningIQ(percentage float64) int {
	percentile := clamp(percentage, screeningPercentileFloor, screeningPercentileCeil) / 100
	zScore := -math.Sqrt2 * math.Erfcinv(2*percentile)
	estimatedIQ := int(math.Round(100 + (15 * zScore)))
	return clampInt(estimatedIQ, 40, 160)
}

func ClassifyEstimatedIQ(estimatedIQ int) string {
	switch {
	case estimatedIQ >= 130:
		return "Sangat superior"
	case estimatedIQ >= 120:
		return "Superior"
	case estimatedIQ >= 110:
		return "Di atas rata-rata"
	case estimatedIQ >= 90:
		return "Rata-rata"
	case estimatedIQ >= 80:
		return "Rata-rata bawah"
	default:
		return "Di bawah rata-rata"
	}
}

func clamp(value float64, min float64, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func clampInt(value int, min int, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
