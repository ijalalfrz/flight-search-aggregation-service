package load_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"testing"

	"github.com/ijalalfrz/flight-search-aggregation-service/internal/app/dto"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Stats struct {
	CacheHits          int
	CacheMisses        int
	ProvidersSucceeded int
	ProvidersFailed    int
}

func (s *Stats) Add(other Stats) {
	s.CacheHits += other.CacheHits
	s.CacheMisses += other.CacheMisses
	s.ProvidersSucceeded += other.ProvidersSucceeded
	s.ProvidersFailed += other.ProvidersFailed
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func clearRedis(t *testing.T, ctx context.Context, rdb *redis.Client) {
	err := rdb.FlushDB(ctx).Err()
	require.NoError(t, err, "Failed to flush Redis")
}

func searchFlights(ctx context.Context, url string, criteria dto.SearchCriteria) (Stats, error) {
	payload, _ := json.Marshal(criteria)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return Stats{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Stats{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// If 404, it means all providers failed (likely due to rate limit)
		return Stats{CacheMisses: 1, ProvidersFailed: 4}, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return Stats{}, fmt.Errorf("bad status: %d, body: %s", resp.StatusCode, string(body))
	}

	var r dto.SearchFlightResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return Stats{}, err
	}

	stats := Stats{}
	if r.Metadata.CacheHit {
		stats.CacheHits = 1
	} else {
		stats.CacheMisses = 1
	}
	stats.ProvidersSucceeded = r.Metadata.ProvidersSucceeded
	stats.ProvidersFailed = r.Metadata.ProvidersFailed

	return stats, nil
}

func TestFlightSearchLoad(t *testing.T) {
	appHost := getEnv("APP_HOST", "http://localhost:8080")
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	redisPass := getEnv("REDIS_PASSWORD", "redis123")

	url := appHost + "/api/v1/flights/search"
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPass,
		DB:       0,
	})
	defer rdb.Close()

	criteria := dto.SearchCriteria{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "economy",
	}

	rateLimitCriteria := dto.SearchCriteria{
		Origin:        "CGK",
		Destination:   "SUB",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "economy",
	}

	t.Run("Cache Miss Test", func(t *testing.T) {
		clearRedis(t, ctx, rdb)
		vus := 5
		stats := runScenario(t, ctx, url, criteria, vus)

		assert.Equal(t, 0, stats.CacheHits)
		assert.Equal(t, vus, stats.CacheMisses)
		assert.Greater(t, stats.ProvidersSucceeded, 0)
	})

	t.Run("Cache Hit Test", func(t *testing.T) {
		clearRedis(t, ctx, rdb)

		// Populate cache
		_, err := searchFlights(ctx, url, criteria)
		require.NoError(t, err)

		vus := 5
		stats := runScenario(t, ctx, url, criteria, vus)

		assert.Equal(t, vus, stats.CacheHits)
		assert.Equal(t, 0, stats.CacheMisses)
	})

	t.Run("Rate Limit Test", func(t *testing.T) {
		clearRedis(t, ctx, rdb)

		vus := 20
		stats := runScenario(t, ctx, url, rateLimitCriteria, vus)

		fmt.Printf("Rate Limit Test Result: Cache Misses = %d, Failed Providers = %d, Succeeded Providers = %d\n", stats.CacheMisses, stats.ProvidersFailed, stats.ProvidersSucceeded)
		assert.Equal(t, vus, stats.CacheMisses, "Should all be cache misses to trigger providers")
		assert.Greater(t, stats.ProvidersFailed, 0, "Should have triggered some rate limits with 20 concurrent requests")
	})
}

func runScenario(t *testing.T, ctx context.Context, url string, criteria dto.SearchCriteria, vus int) Stats {
	var wg sync.WaitGroup
	var mu sync.Mutex
	scenarioStats := Stats{}

	for i := 0; i < vus; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			stats, err := searchFlights(ctx, url, criteria)
			if err != nil {
				t.Errorf("VU %d failed: %v", id, err)
				return
			}
			mu.Lock()
			scenarioStats.Add(stats)
			mu.Unlock()
		}(i)
	}

	wg.Wait()
	return scenarioStats
}
