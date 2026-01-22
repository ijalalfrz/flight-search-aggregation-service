package batikair

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/app/dto"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/pkg/flightprovider"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/pkg/flightprovider/providerutils"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/pkg/utils"
)

const (
	ProviderName = "BatikAir"
	ProviderCode = "ID"
)

type Provider struct {
	Name         string
	SearchAPIURL string
	Timeout      time.Duration
	MaxRetries   int
	Limiter      *redis_rate.Limiter
	RateLimitRPS int
}

func NewProvider(config flightprovider.FlightProviderConfig) *Provider {
	return &Provider{
		Name:         ProviderName,
		SearchAPIURL: config.SearchAPIURL,
		Timeout:      config.Timeout,
		MaxRetries:   config.MaxRetries,
		Limiter:      config.Limiter,
		RateLimitRPS: config.RateLimitRPS,
	}
}

// flightToDTO converts a slice of Flight to a slice of dto.Flight
// it will normalize the data from the provider to the dto.Flight struct
func (p *Provider) flightToDTO(flights []Flight) []dto.Flight {
	results := make([]dto.Flight, len(flights))
	for i, flight := range flights {

		departureTime, err := p.parseTime(flight.DepartureDateTime)
		if err != nil {
			slog.Debug("failed to parse departure time", "error", err)
			continue
		}
		arrivalTime, err := p.parseTime(flight.ArrivalDateTime)
		if err != nil {
			slog.Debug("failed to parse arrival time", "error", err)
			continue
		}

		durationFormat, duration := p.getActualDuration(flight.TravelTime,
			flight.Connections)

		results[i] = dto.Flight{
			ID:       p.generateID(flight.FlightNumber, ProviderName),
			Provider: p.Name,
			Airline: dto.Airline{
				Name: ProviderName,
				Code: ProviderCode,
			},
			FlightNumber: flight.FlightNumber,
			Departure: dto.Departure{
				City:      providerutils.MapAirportToCity[flight.Origin],
				Airport:   flight.Origin,
				Datetime:  departureTime.Format(time.RFC3339),
				Timestamp: departureTime.Unix(),
			},
			Arrival: dto.Arrival{
				City:      providerutils.MapAirportToCity[flight.Destination],
				Airport:   flight.Destination,
				Datetime:  arrivalTime.Format(time.RFC3339),
				Timestamp: arrivalTime.Unix(),
			},
			Duration: dto.Duration{
				TotalMinutes: duration,
				Formatted:    durationFormat,
			},
			Stops: flight.NumberOfStops,
			Price: dto.Price{
				Amount:   float64(p.getTotalPrice(flight.Fare.BasePrice, flight.Fare.Taxes)),
				Currency: flight.Fare.CurrencyCode,
				Formatted: utils.FormatRupiah(p.getTotalPrice(flight.Fare.BasePrice,
					flight.Fare.Taxes)),
			},
			AvailableSeats: flight.SeatsAvailable,
			CabinClass:     p.getCabinClass(flight.Fare.Class),
			Aircraft:       &flight.AircraftModel,
			Amenities:      p.getAmenities(flight.OnboardServices),
			Baggage:        p.parseBaggage(flight.BaggageInfo),
		}
	}

	return results
}

func (p *Provider) getActualDuration(travelTime string, connection []Connection) (string, int) {
	totalMinutes := utils.ConvertDurationToMinutes(travelTime)
	for _, conn := range connection {
		totalMinutes += utils.ConvertDurationToMinutes(conn.StopDuration)
	}

	return utils.ConvertMinutesToDuration(totalMinutes), int(totalMinutes)
}

// Search will simulate API call to AirAsia flight search API
// it will return always success with delay 200-400ms
func (p *Provider) Search(ctx context.Context, criteria dto.SearchCriteria) ([]dto.Flight, error) {
	ctx, cancel := context.WithTimeout(ctx, p.Timeout)
	defer cancel()

	var lastErr error
	for attempt := 0; attempt <= p.MaxRetries; attempt++ {
		// simulate delay 200-400ms
		delay := time.Duration(200+rand.Intn(201)) * time.Millisecond
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled or timeout: %w", ctx.Err())
		}

		// always success not simulate error
		if rand.Float64() < 0 {
			lastErr = providerutils.ErrProviderInternalError
			slog.ErrorContext(ctx, "failed to call batikair flight search API", "attempt",
				attempt+1, "error", lastErr)

			if attempt < p.MaxRetries {
				// Exponential backoff: 200ms * 2^attempt
				backoff := time.Duration(200*(1<<attempt)) * time.Millisecond
				slog.InfoContext(ctx, "retrying with exponential backoff", "backoff",
					backoff, "next_attempt", attempt+2)
				select {
				case <-time.After(backoff):
					continue
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}
			continue
		}

		// rate limit
		res, err := p.Limiter.Allow(ctx, fmt.Sprintf("limit:%s", p.Name),
			redis_rate.PerSecond(p.RateLimitRPS))
		if err != nil {
			return nil, fmt.Errorf("failed to rate limit: %w", err)
		}

		if res.Allowed == 0 {
			return nil, providerutils.ErrProviderRateLimitExceeded
		}

		// get response from mock file
		flightData, err := os.ReadFile(p.SearchAPIURL)
		if err != nil {
			return nil, fmt.Errorf("failed to read mock file: %w", err)
		}

		var response SearchFlightResponse
		if err := json.Unmarshal(flightData, &response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal mock file: %w", err)
		}

		// convert to dto.Flight and filter
		flights := p.flightToDTO(response.Results)

		return providerutils.FilterFlights(flights, criteria), nil
	}

	return nil, fmt.Errorf("failed to get flight data after %d attempts", p.MaxRetries)
}

func (p *Provider) parseTime(timeString string) (time.Time, error) {
	time, err := time.Parse("2006-01-02T15:04:05-0700", timeString)
	if err != nil {
		return time, err
	}

	return time, nil
}

func (p *Provider) getAmenities(amenities []string) []string {
	for i := range amenities {
		amenities[i] = strings.ToLower(amenities[i])
	}

	return amenities
}

func (p *Provider) parseBaggage(note string) dto.Baggage {
	baggage := dto.Baggage{}
	noteSplit := strings.Split(note, ",")
	for _, n := range noteSplit {
		if strings.Contains(n, "cabin") {
			baggage.CarryOn = n
		}

		if strings.Contains(n, "checked") {
			baggage.Checked = n
		}
	}

	return baggage
}
func (p *Provider) generateID(code, name string) string {
	return fmt.Sprintf("%s_%s", code, name)
}

func (p *Provider) getCabinClass(code string) string {
	code = strings.ToUpper(strings.TrimSpace(code))

	classMap := map[string]string{
		"Y": "economy",
		"W": "premium_economy",
		"C": "business",
		"J": "business",
		"F": "first",
	}

	if class, ok := classMap[code]; ok {
		return class
	}

	return "economy"
}

func (p *Provider) getTotalPrice(basePrice int, taxes int) int64 {
	return int64(basePrice + taxes)
}
