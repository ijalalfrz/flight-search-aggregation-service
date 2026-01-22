# Flight Search Aggregation Service

## Prerequisites

- **Docker Desktop**: Required to run the application containers and Redis.
- **Go 1.25+**: (Optional) If you want to run tests locally without Docker.
- **Make**: For running Makefile commands.

## How to Run

### Automatic
This command will setup:
- .env file
- Start all applications in docker-compose
- Run the service with Redis

```bash
make setup
```

### Manual Step by Step

1.  **Start the Application**:
    ```bash
    make start
    ```
    This command uses `docker-compose` to build and start the application service and Redis.

2.  **Restart the Application**:
    ```bash
    make restart
    ```

3.  **Stop the Application**:
    ```bash
    make stop
    ```

4.  **Generate API Documentation**:
    ```bash
    make api-docs
    ```

## How to Test

### Load Tests

To run the load tests simulating concurrent flight searches:

```bash
make tests-load
```

This runs scenarios for:
- Cache miss test (5 concurrent users)
- Cache hit test (5 concurrent users)
- Rate limit test (20 concurrent users)

### Unit Tests

Run all unit tests:

```bash
make tests-unit
```

### Manual API Testing

**Search Flights:**
```bash
curl -X POST http://localhost:8080/api/v1/flights/search \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "CGK",
    "destination": "DPS",
    "departure_date": "2025-12-15",
    "passengers": 1,
    "cabin_class": "economy",
    "sort_option": {
      "sort_by": "price",
      "order": "asc"
    },
    "filter_option": {
      "max_price": 2000000,
      "max_stops": 1
    }
  }'
```

**Available Filter and Sort Options:**

```json
{
  "cabin_class": "string", // economy, business, first
  "departure_date": "string", // YYYY-MM-DD format, example: 2025-12-15
  "origin": "string", // IATA Airport Code, example: CGK
  "destination": "string", // IATA Airport Code, example: DPS
  "filter_option": { // OPTIONAL
    "airline": "string", // airline code, example: GA, JT
    "arrival_time_start": "string", // example: 08:00, overnight flight not supported
    "arrival_time_end": "string", // example: 10:00, overnight flight not supported
    "departure_time_start": "string", // example: 08:00, overnight flight not supported
    "departure_time_end": "string", // example: 10:00, overnight flight not supported
    "min_duration_minutes": 0, // minimum duration in minutes
    "max_duration_minutes": 0, // maximum duration in minutes
    "min_price": 0, 
    "max_price": 0, 
    "min_stops": 0, 
    "max_stops": 0 
  },
  "passengers": 10, // number of passengers max 10
  "sort_option": { // OPTIONAL, default by sort by recommended (best value)
    "field": "string", // price, duration, stops, departure_time, arrival_time, recommended
    "order": "string" // asc, desc
  }
}
```

**Example Request Response**

Request:
```bash
curl --location 'http://localhost:8080/api/v1/flights/search' \
--header 'Content-Type: application/json' \
--data '{
    "origin": "CGK",
    "destination": "DPS",
    "departure_date": "2025-12-15",
    "return_date": null,
    "passengers": 10,
    "cabin_class": "economy",
    "filter_option": {
        "min_price": 400000,
        "max_price": 800001,
        "airline": "JT"
    }
}'
```

Response:
```json
{
    "search_criteria": {
        "origin": "CGK",
        "destination": "DPS",
        "departure_date": "2025-12-15",
        "passengers": 10,
        "cabin_class": "economy",
        "filter_option": {
            "min_price": 400000,
            "max_price": 800001,
            "airline": "JT"
        }
    },
    "metadata": {
        "total_results": 1,
        "providers_queried": 4,
        "providers_succeeded": 4,
        "providers_failed": 0,
        "search_time_ms": 2,
        "cache_hit": true
    },
    "flights": [
        {
            "id": "JT650_LionAir",
            "provider": "LionAir",
            "airline": {
                "name": "LionAir",
                "code": "JT"
            },
            "flight_number": "JT650",
            "departure": {
                "airport": "CGK",
                "city": "Jakarta",
                "datetime": "2025-12-15T16:20:00+07:00",
                "timestamp": 1765790400
            },
            "arrival": {
                "airport": "DPS",
                "city": "Denpasar",
                "datetime": "2025-12-15T21:10:00+08:00",
                "timestamp": 1765804200
            },
            "duration": {
                "total_minutes": 230,
                "formatted": "3h 50m"
            },
            "stops": 1,
            "price": {
                "amount": 780000,
                "currency": "IDR",
                "formatted": "Rp780.000"
            },
            "available_seats": 52,
            "cabin_class": "economy",
            "aircraft": "Boeing 737-800",
            "amenities": [],
            "baggage": {
                "carry_on": "7 kg",
                "checked": "20 kg"
            },
            "score": 0.05
        }
    ]
}
```

Error Response:

- No Flights Found
```json
Status Code: 404
{
    "error": "no flights found"
}
```

- Validation Error
```json
Status Code: 400
{
    "error": "max_price must be greater than min_price"
}
```

**Health Check:**
```bash
curl http://localhost:8080/health
```

## API Documentation

Swagger UI is available at: [http://localhost:8444/docs](http://localhost:8444/docs)
OpenAPI spec JSON: `docs/flight-search-aggregation_swagger.json`

## Architecture Notes

### System Overview

The Flight Search Aggregation Service is a high-performance API that:
- Aggregates flight data from multiple airline providers (LionAir, BatikAir, Garuda, AirAsia)
- Implements caching to reduce provider load
- Applies rate limiting per provider to prevent abuse
- Filters, ranks, and sorts results based on user preferences
- Rank best value flights based on weighted criteria using weighted sum method
- Provides distributed request tracing via Request IDs

### Key Components

**Provider Layer:**
Each airline provider implements:
- Simulated API calls with configurable delays and failure rates
- Exponential backoff retry logic
- Provider-specific rate limiting using Redis (GCRA algorithm) `redis_rate` package.
- Data normalization to common DTO format

**Caching Layer:**
- Redis-based caching with distributed locking
- Cache key based on search criteria
- TTL-based expiration
- Prevents thundering herd via lock acquisition
- Flight cache key use all search criteria:
    - `flight:cache:{departure_date}:{origin}:{destination}:{cabin_class}:{passengers}`
    - `flight:cache:{departure_date}:{origin}:{destination}:{cabin_class}:{passengers}:metadata`


**Aggregation Layer:**
- Concurrent provider queries using goroutines
- Result merging and deduplication
- Filter, rank, and sort pipeline
- Metadata tracking (cache hits, provider success/failure counts)

**Observability:**
- Structured logging with `slog`
- Request ID injection via middleware
- Automatic request ID propagation through context

### Concurrency & Rate Limiting

**Provider Rate Limiting:**
Each provider has a configurable RPS (Requests Per Second) limit enforced via `redis_rate`. This prevents overwhelming individual provider APIs.

**Distributed Locking:**
When multiple concurrent requests arrive for the same search criteria:
1. First request acquires a distributed lock (Redis)
2. Fetches data from providers and saves to cache
3. Releases lock
4. Next requests read from cache


### C4 Diagrams
#### Container Diagram
![Container Diagram](./docs/c4.png)



### Sequence Diagram (Flight Search Flow)

```mermaid
sequenceDiagram
    participant User
    participant Middleware
    participant Endpoint
    participant Aggregator
    participant Cache
    participant Redis
    participant Providers

    User->>Middleware: POST /api/v1/flights/search
    activate Middleware
    Middleware->>Middleware: Generate Request ID
    Middleware->>Endpoint: Forward with Request ID in Context
    activate Endpoint
    
    Endpoint->>Aggregator: SearchFlights(ctx, criteria)
    activate Aggregator
    
    Aggregator->>Cache: GetFlight(cacheKey)
    activate Cache
    Cache->>Redis: GET flight:CGK:DPS:2025-12-15
    
    alt Cache Hit
        Redis-->>Cache: Flight Data
        Cache-->>Aggregator: Cached Flights
        Aggregator->>Aggregator: Filter, Rank, Sort
        Aggregator-->>Endpoint: SearchFlightResponse (cache_hit=true)
    else Cache Miss
        Redis-->>Cache: nil
        Cache-->>Aggregator: nil
        
        par Concurrent Provider Queries
            Aggregator->>Providers: LionAir.Search(ctx, criteria)
            Aggregator->>Providers: BatikAir.Search(ctx, criteria)
            Aggregator->>Providers: Garuda.Search(ctx, criteria)
            Aggregator->>Providers: AirAsia.Search(ctx, criteria)
            
            Providers->>Redis: Check Rate Limit (GCRA)
            alt Rate Limit OK
                Providers->>Providers: Fetch Flight Data
                Providers-->>Aggregator: Flight Results
            else Rate Limit Exceeded
                Providers-->>Aggregator: ErrProviderRateLimitExceeded
            end
        end
        
        Aggregator->>Aggregator: Merge Provider Results
        
        Note over Aggregator,Cache: Try to acquire lock to save to cache
        Aggregator->>Cache: AcquireLock(lockKey, 3s)
        Cache->>Redis: SET lock:CGK:DPS NX EX 3
        
        alt Lock Acquired
            Redis-->>Cache: OK (lock acquired)
            Cache-->>Aggregator: true
            
            Aggregator->>Cache: SetFlight(cacheKey, flights, metadata, TTL=1m)
            Cache->>Redis: SETEX flight:CGK:DPS:2025-12-15, flights
            Redis-->>Cache: OK
            
            Aggregator->>Cache: ReleaseLock(lockKey)
            Cache->>Redis: DEL lock:CGK:DPS
        else Lock Not Acquired
            Redis-->>Cache: nil (another request holds lock)
            Cache-->>Aggregator: false
            Note over Aggregator: Skip cache save, return results anyway
        end
        
        Aggregator->>Aggregator: Filter, Rank, Sort
        Aggregator-->>Endpoint: SearchFlightResponse (cache_hit=false)
    end
    
    deactivate Cache
    deactivate Aggregator
    
    Endpoint-->>Middleware: JSON Response
    deactivate Endpoint
    Middleware-->>User: 200 OK + X-Request-Id Header
    deactivate Middleware
```

## Environment Variables

```env
LOG_LEVEL=debug

# Redis Configuration
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=redis123
REDIS_DB=0

# Cache Configuration
PROVIDER_LOCK_TIMEOUT=3s
PROVIDER_CACHE_EXPIRATION=1m

# Provider config
# Rate limit: assuming lion air provider have rate limit 20 rps and here we will
# define rate limit lower than 20, because we don't want to get rate limit error from provider it self
# or provider might ban you from calling over than limit so we prevent with own rate limit

# Provider Configuration - LionAir
LION_AIR_PROVIDER_SEARCH_URL=tests/mockprovider/lion_air_search_response.json
LION_AIR_PROVIDER_TIMEOUT=5s
LION_AIR_PROVIDER_MAX_RETRIES=3
LION_AIR_PROVIDER_RATE_LIMIT=10

# Provider Configuration - BatikAir
BATIK_AIR_PROVIDER_SEARCH_URL=tests/mockprovider/batik_air_search_response.json
BATIK_AIR_PROVIDER_TIMEOUT=5s
BATIK_AIR_PROVIDER_MAX_RETRIES=3
BATIK_AIR_PROVIDER_RATE_LIMIT=10

# Provider Configuration - Garuda
GARUDA_PROVIDER_SEARCH_URL=tests/mockprovider/garuda_indonesia_search_response.json
GARUDA_PROVIDER_TIMEOUT=5s
GARUDA_PROVIDER_MAX_RETRIES=3
GARUDA_PROVIDER_RATE_LIMIT=10

# Provider Configuration - AirAsia
AIR_ASIA_PROVIDER_SEARCH_URL=tests/mockprovider/airasia_search_response.json
AIR_ASIA_PROVIDER_TIMEOUT=5s
AIR_ASIA_PROVIDER_MAX_RETRIES=3
AIR_ASIA_PROVIDER_RATE_LIMIT=10

# HTTP Server
HTTP_PORT=8080
HTTP_TIMEOUT=30s
```

## Log Level

- `info`: JSON formatted logs
- `debug`: JSON formatted logs with function name, line number, and stack trace on errors
- All logs include `request_id` for distributed tracing

## Performance Characteristics

- **Cache Hit Latency**: ~2-10ms (Redis read + filtering/sorting)
- **Cache Miss Latency**: ~200-400ms (concurrent provider queries + caching)
- **Concurrent Provider Queries**: All providers queried in parallel using goroutines

## Rate Limiting Behavior

The system uses `redis_rate` (https://github.com/go-redis/redis_rate):
- Providers are rate-limited individually
- Failed requests due to rate limiting are tracked in metadata
- Rate limits are enforced across all service instances (distributed)