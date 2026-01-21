package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/ijalalfrz/flight-search-aggregation-service/internal/app/dto"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/pkg/flight"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/pkg/flightprovider"
)

type FlightCacher interface {
	GetLockKey(req dto.SearchCriteria) string
	GetCacheKey(req dto.SearchCriteria) string
	AcquireLock(ctx context.Context, key string, timeout time.Duration) (bool, error)
	ReleaseLock(ctx context.Context, key string) error
	GetFlight(ctx context.Context, key string) ([]dto.Flight, error)
	GetMetadata(ctx context.Context, key string) (dto.Metadata, error)
	SetFlight(ctx context.Context,
		key string,
		flights []dto.Flight,
		metadata dto.Metadata,
		expiration time.Duration,
	) error
}

type providerResult struct {
	Provider string
	Flights  []dto.Flight
	Error    error
}

type AggregatorService struct {
	ProviderFactory       *flightprovider.FlightProviderFactory
	Cache                 FlightCacher
	FlightCacheExpiration time.Duration
	FlightLockTimeout     time.Duration
}

func NewAggregatorService(providerFactory *flightprovider.FlightProviderFactory,
	cache FlightCacher, flightCacheExpiration time.Duration,
	flightLockTimeout time.Duration) *AggregatorService {
	return &AggregatorService{
		ProviderFactory:       providerFactory,
		Cache:                 cache,
		FlightCacheExpiration: flightCacheExpiration,
	}
}

// SearchFlights aggregates flights from all providers and returns the best flights
// It uses filter, rank, and sort functions to process the flights
// SearchFlights godoc
// @Summary      Search flights
// @Tags         Flights
// @Description  Search flights from all providers and return the best flights
// @Param        request  body      dto.SearchCriteria  true  "Search Criteria"
// @Success      200      {object}  dto.SearchFlightResponse
// @Failure      404      {object}  dto.ErrorResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Failure      500      {object}  dto.ErrorResponse
// @Router       /api/v1/flights/search [post]
func (s *AggregatorService) SearchFlights(
	ctx context.Context,
	req dto.SearchCriteria,
) (dto.SearchFlightResponse, error) {
	var (
		flights                 []dto.Flight
		metadata                dto.Metadata
		numberOfProviders       int
		numberOfFailedProviders int
	)

	startTime := time.Now()
	cacheHit := false

	// get from cache first
	cacheKey := s.Cache.GetCacheKey(req)
	lockKey := s.Cache.GetLockKey(req)

	flights, err := s.Cache.GetFlight(ctx, cacheKey)
	if err == nil {
		cacheHit = true
	} else {
		slog.WarnContext(ctx, "failed to get flight from cache", slog.String("error", err.Error()))
	}

	metadata, err = s.Cache.GetMetadata(ctx, cacheKey)
	if err != nil {
		slog.WarnContext(ctx, "failed to get metadata from cache", slog.String("error", err.Error()))
	}

	// cache miss get from provider and store to cache
	if !cacheHit {
		// if there is concurrent request with same criteria, only one will be processed
		// to save to cache so next request with same criterial will hit the cache
		// e.g. 3 request
		// request 1 -> acquire lock, process, save to cache, release lock
		// request 2 -> acquire lock, lock not aquired, get from provider
		// request 3 -> acquire lock, lock not aquired, get from provider
		// this ensure only 1 operation that fetch from provider and save to cache

		// get fligt and metadata
		flights, numberOfProviders,
			numberOfFailedProviders, err = s.getFromProvider(ctx, req)
		if err != nil && !errors.Is(err, ErrNoFlightsFound) {
			return dto.SearchFlightResponse{}, fmt.Errorf("failed to get flights from providers: %w", err)
		}

		metadata = dto.Metadata{
			ProvidersQueried:   numberOfProviders,
			ProvidersSucceeded: numberOfProviders - numberOfFailedProviders,
			ProvidersFailed:    numberOfFailedProviders,
		}

		// lock to process cache
		acquired, err := s.Cache.AcquireLock(ctx, lockKey, s.FlightLockTimeout)
		if err != nil {
			return dto.SearchFlightResponse{}, fmt.Errorf("failed to acquire lock: %w", err)
		}
		defer s.Cache.ReleaseLock(ctx, lockKey)

		if acquired {
			// if lock is acquired, process the request and save to cache
			err = s.Cache.SetFlight(ctx, cacheKey, flights, metadata, s.FlightCacheExpiration)
			if err != nil {
				return dto.SearchFlightResponse{}, fmt.Errorf("failed to set flights to cache: %w", err)
			}

		}
	}

	// filter, rank, and sort flights
	filteredFlights := flight.FilterFlights(ctx, flights, req.FilterOption)
	rankedFlights := flight.RankFlights(filteredFlights)
	sortedFlights := flight.SortFlights(rankedFlights, req.SortOption)

	// metadata
	metadata.TotalResults = len(sortedFlights)
	metadata.SearchTimeMs = int(time.Since(startTime).Milliseconds())
	metadata.CacheHit = cacheHit

	if len(sortedFlights) == 0 {
		return dto.SearchFlightResponse{}, ErrNoFlightsFound
	}

	return dto.SearchFlightResponse{
		Flights:        sortedFlights,
		SearchCriteria: req,
		Metadata:       metadata,
	}, nil
}

func (s *AggregatorService) getFromProvider(ctx context.Context,
	req dto.SearchCriteria,
) ([]dto.Flight, int, int, error) {
	providers := s.ProviderFactory.GetAllProviders()
	results := make(chan providerResult, len(providers))
	var wg sync.WaitGroup

	// concurrently call all providers
	// timeout for each provider is set in the provider itself
	wg.Add(len(providers))
	for key, provider := range providers {
		go func(key string, p flightprovider.FlightProvider) {
			defer wg.Done()
			flights, err := p.Search(ctx, req)
			results <- providerResult{
				Provider: key,
				Flights:  flights,
				Error:    err,
			}
		}(key, provider)
	}

	// wait all go routine finish
	go func() {
		wg.Wait()
		close(results)
	}()

	numberOfFailedProviders := 0
	var allFlights []dto.Flight
	for result := range results {
		if result.Error != nil {
			slog.WarnContext(ctx, "provider failed",
				slog.String("provider", result.Provider),
				slog.Any("error", result.Error))
			numberOfFailedProviders++
			continue
		}
		allFlights = append(allFlights, result.Flights...)
	}

	if len(allFlights) == 0 {
		return []dto.Flight{}, len(providers), numberOfFailedProviders, ErrNoFlightsFound
	}

	return allFlights, len(providers), numberOfFailedProviders, nil
}
