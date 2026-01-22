package providerutils

import (
	"time"

	"github.com/ijalalfrz/flight-search-aggregation-service/internal/app/dto"
)

func FilterFlights(flights []dto.Flight, criteria dto.SearchCriteria) []dto.Flight {
	results := make([]dto.Flight, 0, len(flights))

	for _, flight := range flights {
		if criteria.Origin != "" && flight.Departure.Airport != criteria.Origin {
			continue
		}

		if criteria.Destination != "" && flight.Arrival.Airport != criteria.Destination {
			continue
		}

		if criteria.DepartureDate != "" {
			departureTime, _ := time.Parse(time.RFC3339, flight.Departure.Datetime)
			if departureTime.Format("2006-01-02") != criteria.DepartureDate {
				continue
			}
		}

		if criteria.CabinClass != "" && flight.CabinClass != criteria.CabinClass {
			continue
		}

		if criteria.Passengers != 0 && flight.AvailableSeats < criteria.Passengers {
			continue
		}

		results = append(results, flight)
	}

	return results
}
