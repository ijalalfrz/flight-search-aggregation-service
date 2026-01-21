//go:build unit

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/app/dto"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/pkg/flightprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAggregatorService_SearchFlights(t *testing.T) {
	type mockField struct {
		cache    *MockFlightCacher
		provider *flightprovider.MockFlightProvider
	}

	searchFlightRequest := func(
		criteria dto.SearchCriteria,
		setupMock func(m mockField),
		want dto.SearchFlightResponse,
		wantErr error,
	) func(t *testing.T) {
		return func(t *testing.T) {
			m := mockField{
				cache:    NewMockFlightCacher(t),
				provider: flightprovider.NewMockFlightProvider(t),
			}
			setupMock(m)

			factory := flightprovider.NewFlightProviderFactory()
			factory.AddProvider("test-provider", m.provider)

			s := &AggregatorService{
				ProviderFactory:       factory,
				Cache:                 m.cache,
				FlightCacheExpiration: 10 * time.Minute,
				FlightLockTimeout:     5 * time.Second,
			}

			got, err := s.SearchFlights(context.Background(), criteria)

			if wantErr != nil {
				assert.Error(t, err)
				if !errors.Is(err, wantErr) && err.Error() != wantErr.Error() {
					t.Fatalf("expected error %v, got %v", wantErr, err)
				}
				return
			}

			assert.NoError(t, err)
			// Reset SearchTimeMs to 0 for comparison as it's dynamic
			got.Metadata.SearchTimeMs = 0
			want.Metadata.SearchTimeMs = 0

			diff := cmp.Diff(want, got)
			if diff != "" {
				t.Fatalf("SearchFlights() mismatch (-want +got):\n%s", diff)
			}
		}
	}

	criteria := dto.SearchCriteria{
		Origin:        "JKT",
		Destination:   "DPS",
		DepartureDate: "2024-01-01",
		Passengers:    1,
		CabinClass:    "ECONOMY",
	}

	flights := []dto.Flight{
		{
			ID:       "flight-1",
			Provider: "test-provider",
			Price: dto.Price{
				Amount:   1000000,
				Currency: "IDR",
			},
		},
	}

	t.Run("cache_hit", searchFlightRequest(
		criteria,
		func(m mockField) {
			m.cache.On("GetCacheKey", criteria).Return("cache-key")
			m.cache.On("GetLockKey", criteria).Return("lock-key")
			m.cache.On("GetFlight", mock.Anything, "cache-key").Return(flights, nil)
			m.cache.On("GetMetadata", mock.Anything, "cache-key").Return(dto.Metadata{
				ProvidersQueried:   1,
				ProvidersSucceeded: 1,
			}, nil)
		},
		dto.SearchFlightResponse{
			Flights:        flights,
			SearchCriteria: criteria,
			Metadata: dto.Metadata{
				ProvidersQueried:   1,
				ProvidersSucceeded: 1,
				TotalResults:       1,
				CacheHit:           true,
			},
		},
		nil,
	))

	t.Run("cache_miss_success", searchFlightRequest(
		criteria,
		func(m mockField) {
			m.cache.On("GetCacheKey", criteria).Return("cache-key")
			m.cache.On("GetLockKey", criteria).Return("lock-key")
			m.cache.On("GetFlight", mock.Anything, "cache-key").Return(nil, errors.New("miss"))
			m.cache.On("GetMetadata", mock.Anything, "cache-key").Return(dto.Metadata{}, errors.New("miss"))
			m.provider.On("Search", mock.Anything, criteria).Return(flights, nil)
			m.cache.On("AcquireLock", mock.Anything, "lock-key", 5*time.Second).Return(true, nil)
			m.cache.On("SetFlight", mock.Anything, "cache-key", flights, mock.Anything, 10*time.Minute).Return(nil)
			m.cache.On("ReleaseLock", mock.Anything, "lock-key").Return(nil)
		},
		dto.SearchFlightResponse{
			Flights:        flights,
			SearchCriteria: criteria,
			Metadata: dto.Metadata{
				ProvidersQueried:   1,
				ProvidersSucceeded: 1,
				TotalResults:       1,
				CacheHit:           false,
			},
		},
		nil,
	))

	t.Run("no_flights_found", searchFlightRequest(
		criteria,
		func(m mockField) {
			m.cache.On("GetCacheKey", criteria).Return("cache-key")
			m.cache.On("GetLockKey", criteria).Return("lock-key")
			m.cache.On("GetFlight", mock.Anything, "cache-key").Return(nil, errors.New("miss"))
			m.cache.On("GetMetadata", mock.Anything, "cache-key").Return(dto.Metadata{}, errors.New("miss"))
			m.provider.On("Search", mock.Anything, criteria).Return([]dto.Flight{}, nil)
			m.cache.On("AcquireLock", mock.Anything, "lock-key", 5*time.Second).Return(true, nil)
			m.cache.On("SetFlight", mock.Anything, "cache-key", []dto.Flight{}, mock.Anything, 10*time.Minute).Return(nil)
			m.cache.On("ReleaseLock", mock.Anything, "lock-key").Return(nil)
		},
		dto.SearchFlightResponse{},
		ErrNoFlightsFound,
	))
}
