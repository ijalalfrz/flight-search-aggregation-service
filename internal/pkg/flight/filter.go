package flight

import (
	"context"
	"log/slog"
	"time"

	"github.com/ijalalfrz/flight-search-aggregation-service/internal/app/dto"
)

func FilterFlights(ctx context.Context, flights []dto.Flight, filterOpts *dto.FilterOption) []dto.Flight {
	if filterOpts == nil {
		return flights
	}

	results := make([]dto.Flight, 0, len(flights))

	for _, flight := range flights {
		if filterOpts.Airline != nil && *filterOpts.Airline != flight.Airline.Code {
			continue
		}

		if filterOpts.MaxPrice != nil && flight.Price.Amount > *filterOpts.MaxPrice {
			continue
		}

		if filterOpts.MinPrice != nil && flight.Price.Amount < *filterOpts.MinPrice {
			continue
		}

		if filterOpts.MaxStops != nil && flight.Stops > *filterOpts.MaxStops {
			continue
		}

		if filterOpts.MinStops != nil && flight.Stops < *filterOpts.MinStops {
			continue
		}

		if filterOpts.MaxDurationMinutes != nil && flight.Duration.TotalMinutes > *filterOpts.MaxDurationMinutes {
			continue
		}

		if filterOpts.MinDurationMinutes != nil && flight.Duration.TotalMinutes < *filterOpts.MinDurationMinutes {
			continue
		}

		if filterOpts.DepartureTimeStart != nil && filterOpts.DepartureTimeEnd != nil {
			if !isWithinTimeRange(ctx, flight.Departure.Datetime, *filterOpts.DepartureTimeStart, *filterOpts.DepartureTimeEnd) {
				continue
			}
		}

		if filterOpts.ArrivalTimeStart != nil && filterOpts.ArrivalTimeEnd != nil {
			if !isWithinTimeRange(ctx, flight.Arrival.Datetime, *filterOpts.ArrivalTimeStart, *filterOpts.ArrivalTimeEnd) {
				continue
			}
		}

		results = append(results, flight)
	}

	return results
}

// startTime and endTime will be time only, without date and will depend on targetTime timezone
// e.g. arrival at gmt+8 at 14:00, so will be checked if 14:00 is between startTime and endTime
func isWithinTimeRange(ctx context.Context, targetTime string, startTime string, endTime string) bool {
	targetTimeParsed, err := time.Parse(time.RFC3339, targetTime)
	if err != nil {
		return false
	}

	startTimeParsed, err := time.ParseInLocation("15:04", startTime, targetTimeParsed.Location())
	if err != nil {
		slog.ErrorContext(ctx, "failed to parse start time", slog.String("time", startTime), slog.Any("error", err))
		return false
	}

	endTimeParsed, err := time.ParseInLocation("15:04", endTime, targetTimeParsed.Location())
	if err != nil {
		slog.ErrorContext(ctx, "failed to parse end time", slog.String("time", endTime), slog.Any("error", err))
		return false
	}

	return targetTimeParsed.Hour() >= startTimeParsed.Hour() &&
		targetTimeParsed.Hour() <= endTimeParsed.Hour()
}
