package flight

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ijalalfrz/flight-search-aggregation-service/internal/app/dto"
	"github.com/redis/go-redis/v9"
)

type RedisClient interface {
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
}

type FlightCache struct {
	redis RedisClient
}

func NewFlightCache(redis RedisClient) *FlightCache {
	return &FlightCache{
		redis: redis,
	}
}

func (c *FlightCache) GetLockKey(req dto.SearchCriteria) string {
	return fmt.Sprintf("flight:lock:%s:%s:%s:%s:%d",
		req.DepartureDate, req.Origin, req.Destination, req.CabinClass, req.Passengers)
}

func (c *FlightCache) GetCacheKey(req dto.SearchCriteria) string {
	return fmt.Sprintf("flight:cache:%s:%s:%s:%s:%d",
		req.DepartureDate, req.Origin, req.Destination, req.CabinClass, req.Passengers)
}

func (c *FlightCache) AcquireLock(ctx context.Context, key string, timeout time.Duration) (bool, error) {
	return c.redis.SetNX(ctx, key, "1", timeout).Result()
}

func (c *FlightCache) ReleaseLock(ctx context.Context, key string) error {
	return c.redis.Del(ctx, key).Err()
}

func (c *FlightCache) SetFlight(ctx context.Context,
	key string,
	flights []dto.Flight,
	metadata dto.Metadata,
	expiration time.Duration,
) error {
	data, err := json.Marshal(flights)
	if err != nil {
		return fmt.Errorf("failed to marshal flights: %w", err)
	}

	err = c.redis.Set(ctx, key, data, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set flights: %w", err)
	}

	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	err = c.redis.Set(ctx, key+":metadata", metadataBytes, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set metadata: %w", err)
	}

	return nil
}

func (c *FlightCache) GetFlight(ctx context.Context, key string) ([]dto.Flight, error) {

	data, err := c.redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var flights []dto.Flight
	if err := json.Unmarshal(data, &flights); err != nil {
		return nil, err
	}

	return flights, nil
}

func (c *FlightCache) GetMetadata(ctx context.Context, key string) (dto.Metadata, error) {
	metadataBytes, err := c.redis.Get(ctx, key+":metadata").Bytes()
	if err != nil {
		return dto.Metadata{}, err
	}

	var metadata dto.Metadata
	if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
		return dto.Metadata{}, err
	}

	return metadata, nil
}
