package flight

import (
	"sort"

	"github.com/ijalalfrz/flight-search-aggregation-service/internal/app/dto"
)

func SortFlights(flights []dto.Flight, sortOption *dto.SortOption) []dto.Flight {
	var (
		option = ""
		order  = "asc"
	)
	if sortOption != nil {
		option = sortOption.Field
		order = sortOption.Order
	}

	switch option {
	case "price":
		sort.Slice(flights, func(i, j int) bool {
			if order == "asc" {
				return flights[i].Price.Amount < flights[j].Price.Amount
			} else {
				return flights[i].Price.Amount > flights[j].Price.Amount
			}
		})
	case "duration":
		sort.Slice(flights, func(i, j int) bool {
			if order == "asc" {
				return flights[i].Duration.TotalMinutes < flights[j].Duration.TotalMinutes
			} else {
				return flights[i].Duration.TotalMinutes > flights[j].Duration.TotalMinutes
			}
		})
	case "stops":
		sort.Slice(flights, func(i, j int) bool {
			if order == "asc" {
				return flights[i].Stops < flights[j].Stops
			} else {
				return flights[i].Stops > flights[j].Stops
			}
		})
	case "departure_time":
		sort.Slice(flights, func(i, j int) bool {
			if order == "asc" {
				return flights[i].Departure.Timestamp < flights[j].Departure.Timestamp
			} else {
				return flights[i].Departure.Timestamp > flights[j].Departure.Timestamp
			}
		})
	case "arrival_time":
		sort.Slice(flights, func(i, j int) bool {
			if order == "asc" {
				return flights[i].Arrival.Timestamp < flights[j].Arrival.Timestamp
			} else {
				return flights[i].Arrival.Timestamp > flights[j].Arrival.Timestamp
			}
		})
	default:
		// best score
		sort.Slice(flights, func(i, j int) bool {
			if order == "asc" {
				return flights[i].Score < flights[j].Score
			} else {
				return flights[i].Score > flights[j].Score
			}
		})
	}

	return flights
}
