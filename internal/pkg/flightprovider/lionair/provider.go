package lionair

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
	ProviderName = "LionAir"
	ProviderCode = "JT"
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

// Search will simulate API call to LionAir flight search API
// it will return always success with delay 100-200ms
func (p *Provider) Search(ctx context.Context,
	criteria dto.SearchCriteria,
) ([]dto.Flight, error) {
	ctx, cancel := context.WithTimeout(ctx, p.Timeout)
	defer cancel()

	var lastErr error
	for attempt := 0; attempt <= p.MaxRetries; attempt++ {
		// simulate delay 100-200ms
		delay := time.Duration(100+rand.Intn(101)) * time.Millisecond
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled or timeout: %w", ctx.Err())
		}

		// always success not simulate error
		if rand.Float64() < 0 {
			lastErr = providerutils.ErrProviderInternalError
			slog.ErrorContext(ctx, "failed to call lionair flight search API", "attempt",
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
		flights := p.flightToDTO(response.Data.AvailableFlights)
		return providerutils.FilterFlights(flights, criteria), nil
	}

	return nil, fmt.Errorf("failed to get flight data after %d attempts", p.MaxRetries)
}

func (p *Provider) flightToDTO(flights []Flight) []dto.Flight {
	results := make([]dto.Flight, len(flights))
	for i, flight := range flights {
		deptTime := p.parseTimeWithLocation(flight.Schedule.Departure, flight.Schedule.DepartureTimezone)
		arrTime := p.parseTimeWithLocation(flight.Schedule.Arrival, flight.Schedule.ArrivalTimezone)
		amenities := p.getAmenities(flight.Services)

		results[i] = dto.Flight{
			ID:       p.generateID(flight.ID, p.Name),
			Provider: p.Name,
			Airline: dto.Airline{
				Name: ProviderName,
				Code: ProviderCode,
			},
			FlightNumber: flight.ID,
			Departure: dto.Departure{
				Airport:   flight.Route.From.Code,
				City:      flight.Route.From.City,
				Datetime:  deptTime.Format(time.RFC3339),
				Timestamp: deptTime.Unix(),
			},
			Arrival: dto.Arrival{
				Airport:   flight.Route.To.Code,
				City:      flight.Route.To.City,
				Datetime:  arrTime.Format(time.RFC3339),
				Timestamp: arrTime.Unix(),
			},
			Duration: dto.Duration{
				TotalMinutes: flight.FlightTime,
				Formatted:    utils.ConvertMinutesToDuration(int64(flight.FlightTime)),
			},
			Stops: flight.StopCount,
			Price: dto.Price{
				Amount:    float64(flight.Pricing.Total),
				Currency:  flight.Pricing.Currency,
				Formatted: utils.FormatRupiah(int64(flight.Pricing.Total)),
			},
			AvailableSeats: flight.SeatsLeft,
			CabinClass:     strings.ToLower(flight.Pricing.FareType),
			Aircraft:       &flight.PlaneType,
			Amenities:      amenities,
			Baggage: dto.Baggage{
				CarryOn: flight.Services.BaggageAllowance.Cabin,
				Checked: flight.Services.BaggageAllowance.Hold,
			},
		}
	}
	return results
}

func (p *Provider) generateID(code, name string) string {
	return fmt.Sprintf("%s_%s", code, name)
}

func (p *Provider) parseTimeWithLocation(timeStr, locationStr string) time.Time {
	loc, err := time.LoadLocation(locationStr)
	if err != nil {
		loc = time.UTC
	}
	t, _ := time.ParseInLocation("2006-01-02T15:04:05", timeStr, loc)
	return t
}

func (p *Provider) getAmenities(services Services) []string {
	amenities := []string{}
	if services.WifiAvailable {
		amenities = append(amenities, "wifi")
	}
	if services.MealsIncluded {
		amenities = append(amenities, "meal")
	}
	return amenities
}
