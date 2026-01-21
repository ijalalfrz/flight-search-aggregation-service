package flight

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/app/dto"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
)

func TestFlightCache_GetLockKey_Closure(t *testing.T) {
	getLockKeyRequest := func(req dto.SearchCriteria, want string) func(t *testing.T) {
		return func(t *testing.T) {
			c := &FlightCache{}
			got := c.GetLockKey(req)
			if got != want {
				t.Fatalf("expected %s, got %s", want, got)
			}
		}
	}

	req := dto.SearchCriteria{
		Origin:        "JKT",
		Destination:   "DPS",
		DepartureDate: "2024-01-01",
		CabinClass:    "ECONOMY",
		Passengers:    1,
	}
	t.Run("basic_lock_key", getLockKeyRequest(req, "flight:lock:2024-01-01:JKT:DPS:ECONOMY:1"))
}

func TestFlightCache_AcquireLock_Closure(t *testing.T) {
	acquireLockRequest := func(key string, timeout time.Duration, mockSetup func(m *MockRedisClient), want bool) func(t *testing.T) {
		return func(t *testing.T) {
			m := NewMockRedisClient(t)
			mockSetup(m)
			c := NewFlightCache(m)

			got, err := c.AcquireLock(context.Background(), key, timeout)
			if err != nil {
				t.Fatalf("AcquireLock returned error: %v", err)
			}
			if got != want {
				t.Fatalf("expected %v, got %v", want, got)
			}
		}
	}

	t.Run("lock_acquired", acquireLockRequest("test-key", 5*time.Second, func(m *MockRedisClient) {
		m.On("SetNX", mock.Anything, "test-key", "1", 5*time.Second).Return(redis.NewBoolResult(true, nil))
	}, true))

	t.Run("lock_not_acquired", acquireLockRequest("test-key", 5*time.Second, func(m *MockRedisClient) {
		m.On("SetNX", mock.Anything, "test-key", "1", 5*time.Second).Return(redis.NewBoolResult(false, nil))
	}, false))
}

func TestFlightCache_SetFlight_Closure(t *testing.T) {
	setFlightRequest := func(key string, flights []dto.Flight, meta dto.Metadata, exp time.Duration, mockSetup func(m *MockRedisClient)) func(t *testing.T) {
		return func(t *testing.T) {
			m := NewMockRedisClient(t)
			mockSetup(m)
			c := NewFlightCache(m)

			err := c.SetFlight(context.Background(), key, flights, meta, exp)
			if err != nil {
				t.Fatalf("SetFlight returned error: %v", err)
			}
		}
	}

	flights := []dto.Flight{{ID: "1"}}
	meta := dto.Metadata{TotalResults: 1}

	t.Run("success", setFlightRequest("test-cache", flights, meta, 10*time.Minute, func(m *MockRedisClient) {
		m.On("Set", mock.Anything, "test-cache", mock.Anything, 10*time.Minute).Return(redis.NewStatusResult("OK", nil))
		m.On("Set", mock.Anything, "test-cache:metadata", mock.Anything, 10*time.Minute).Return(redis.NewStatusResult("OK", nil))
	}))
}

func TestFlightCache_GetFlight_Closure(t *testing.T) {
	getFlightRequest := func(key string, mockSetup func(m *MockRedisClient), want []dto.Flight, wantErr bool) func(t *testing.T) {
		return func(t *testing.T) {
			m := NewMockRedisClient(t)
			mockSetup(m)
			c := NewFlightCache(m)

			got, err := c.GetFlight(context.Background(), key)
			if (err != nil) != wantErr {
				t.Fatalf("GetFlight error = %v, wantErr %v", err, wantErr)
			}
			if !wantErr {
				diff := cmp.Diff(want, got)
				if diff != "" {
					t.Fatalf("GetFlight mismatch (-want +got):\n%s", diff)
				}
			}
		}
	}

	flights := []dto.Flight{{ID: "1"}}
	t.Run("success", getFlightRequest("test-cache", func(m *MockRedisClient) {
		m.On("Get", mock.Anything, "test-cache").Return(redis.NewStringResult(`[{"id":"1"}]`, nil))
	}, flights, false))

	t.Run("cache_miss", getFlightRequest("test-cache", func(m *MockRedisClient) {
		m.On("Get", mock.Anything, "test-cache").Return(redis.NewStringResult("", redis.Nil))
	}, nil, true))
}
