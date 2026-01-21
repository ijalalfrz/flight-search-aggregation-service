package flightprovider

import (
	"context"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/app/dto"
)

// config for flight provider
type FlightProviderConfig struct {
	SearchAPIURL string
	Timeout      time.Duration
	MaxRetries   int
	RateLimitRPS int
	Limiter      *redis_rate.Limiter
}

type FlightProvider interface {
	Search(ctx context.Context, criteria dto.SearchCriteria) ([]dto.Flight, error)
}

type FlightProviderFactory struct {
	Provider map[string]FlightProvider
}

func NewFlightProviderFactory() *FlightProviderFactory {
	return &FlightProviderFactory{
		Provider: make(map[string]FlightProvider),
	}
}

func (f *FlightProviderFactory) AddProvider(name string, provider FlightProvider) {
	f.Provider[name] = provider
}

func (f *FlightProviderFactory) GetProvider(name string) FlightProvider {
	return f.Provider[name]
}

func (f *FlightProviderFactory) GetAllProviders() map[string]FlightProvider {
	return f.Provider
}
