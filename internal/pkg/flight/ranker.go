package flight

import (
	"math"

	"github.com/ijalalfrz/flight-search-aggregation-service/internal/app/dto"
)

// weighted scoring using normalization
// ref: https://www.1000minds.com/decision-making/what-is-mcdm-mcda

// weights for each criteria
const (
	WeightPrice             = 0.6
	WeightDurationInMinutes = 0.2
	WeightStops             = 0.15
	WeightAmenities         = 0.05
)

// RankFlights ranks the flights based on the given criteria
// score is calculated using weighted scoring using normalization
// 0 indicates the best flight and 1 indicates the worst flight
func RankFlights(flights []dto.Flight) []dto.Flight {
	priceMin, priceMax := findPriceRange(flights)
	durationMin, durationMax := findDurationRange(flights)
	stopsMin, stopsMax := findStopsRange(flights)
	amenitiesMin, amenitiesMax := findAmenitiesRange(flights)

	for i, flight := range flights {
		priceScore := normalizeValue(flight.Price.Amount, priceMin, priceMax)
		durationScore := normalizeValue(float64(flight.Duration.TotalMinutes),
			float64(durationMin), float64(durationMax))
		stopsScore := normalizeValue(float64(flight.Stops),
			float64(stopsMin), float64(stopsMax))

		// invert amenities score because more amenities is better
		amenitiesScore := 1 - normalizeValue(float64(len(flight.Amenities)),
			float64(amenitiesMin), float64(amenitiesMax))

		flights[i].Score = WeightPrice*priceScore +
			WeightDurationInMinutes*durationScore +
			WeightStops*stopsScore +
			WeightAmenities*amenitiesScore
	}

	return flights
}

func findPriceRange(flights []dto.Flight) (float64, float64) {
	if len(flights) == 0 {
		return 0, 0
	}

	minPrice := math.MaxFloat64
	maxPrice := -math.MaxFloat64
	for _, flight := range flights {
		if flight.Price.Amount < minPrice {
			minPrice = flight.Price.Amount
		}
		if flight.Price.Amount > maxPrice {
			maxPrice = flight.Price.Amount
		}
	}
	return minPrice, maxPrice
}

func findDurationRange(flights []dto.Flight) (int, int) {
	if len(flights) == 0 {
		return 0, 0
	}

	minDuration := math.MaxInt
	maxDuration := -math.MaxInt
	for _, flight := range flights {
		if flight.Duration.TotalMinutes < minDuration {
			minDuration = flight.Duration.TotalMinutes
		}
		if flight.Duration.TotalMinutes > maxDuration {
			maxDuration = flight.Duration.TotalMinutes
		}
	}
	return minDuration, maxDuration
}

func findStopsRange(flights []dto.Flight) (int, int) {
	if len(flights) == 0 {
		return 0, 0
	}

	minStops := math.MaxInt
	maxStops := -math.MaxInt
	for _, flight := range flights {
		if flight.Stops < minStops {
			minStops = flight.Stops
		}
		if flight.Stops > maxStops {
			maxStops = flight.Stops
		}
	}
	return minStops, maxStops
}

func findAmenitiesRange(flights []dto.Flight) (int, int) {
	if len(flights) == 0 {
		return 0, 0
	}

	minAmenities := math.MaxInt
	maxAmenities := -math.MaxInt
	for _, flight := range flights {
		if len(flight.Amenities) < minAmenities {
			minAmenities = len(flight.Amenities)
		}
		if len(flight.Amenities) > maxAmenities {
			maxAmenities = len(flight.Amenities)
		}
	}
	return minAmenities, maxAmenities
}

func normalizeValue(value float64, min float64, max float64) float64 {
	if max == min {
		return 0
	}

	return (value - min) / (max - min)
}
