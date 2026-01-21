package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/go-redis/redis_rate/v10"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/app/config"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/app/dto"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/app/endpoints"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/app/service"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/app/transport"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/pkg/flight"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/pkg/flightprovider"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/pkg/flightprovider/airasia"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/pkg/flightprovider/batikair"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/pkg/flightprovider/garuda"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/pkg/flightprovider/lionair"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// @title           Flight Search Aggregation Service API
// @version         0.0.1
// @description     flight-search-aggregation-service
// @host      localhost:8080
// @BasePath  /
// @license.name Rizal Alfarizi
// @license.url https://github.com/ijalalfrz
func main() {

	cfg := config.MustInitConfig(".env")
	logger.InitStructuredLogger(cfg.LogLevel)

	slog.Debug("config loaded successfully", slog.Any("config", cfg))
	runApp(cfg)
}

func runApp(cfg config.Config) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	slog.InfoContext(ctx, "starting...", slog.String("log_level", string(cfg.LogLevel)))

	var waitGroup sync.WaitGroup
	// Starts the server in a go routine
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		startHTTPServer(ctx, cfg)
	}()

	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case sig := <-sigChannel:
		cancel()
		slog.InfoContext(ctx, "received OS signal. Exiting...", slog.String("signal", sig.String()))
	case <-ctx.Done():
		slog.ErrorContext(ctx, "failed to start HTTP server")
	}

	waitGroup.Wait()
	slog.InfoContext(ctx, "All service closed...")
}

func startHTTPServer(ctx context.Context, cfg config.Config) {
	endpts := makeEndpoints(ctx, &cfg)
	router := transport.MakeHTTPRouter(&cfg, endpts)
	server := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf(":%d", cfg.HTTP.Port),
		WriteTimeout: cfg.HTTP.Timeout,
		ReadTimeout:  cfg.HTTP.Timeout,
	}

	slog.Info("running HTTP server...", slog.Int("port", cfg.HTTP.Port))

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.ErrorContext(ctx, "failed to start HTTP server", slog.String("error", err.Error()))
		}
	}()

	<-ctx.Done()

	if err := server.Shutdown(ctx); err != nil {
		slog.ErrorContext(ctx, "failed to shutdown HTTP server", slog.String("error", err.Error()))
	}

	slog.InfoContext(ctx, "HTTP server shutdown gracefully")
}

func makeEndpoints(ctx context.Context, cfg *config.Config) endpoints.Endpoints {
	// init redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// init validator
	if err := dto.InitValidator(); err != nil {
		slog.ErrorContext(ctx, "failed to init validator", slog.String("error", err.Error()))
		panic(err)
	}

	// init factory
	flightProviderFactory := initFlightProviderFactory(cfg, redisClient)

	// init service endpoint
	return endpoints.Endpoints{
		AggregatorEndpoint: makeAggregatorEndpoint(flightProviderFactory, redisClient, cfg),
	}
}

// register flight provider
func initFlightProviderFactory(cfg *config.Config, redisClient *redis.Client) *flightprovider.FlightProviderFactory {

	limiter := redis_rate.NewLimiter(redisClient)

	factory := flightprovider.NewFlightProviderFactory()
	factory.AddProvider(lionair.ProviderName, lionair.NewProvider(flightprovider.FlightProviderConfig{
		SearchAPIURL: cfg.Providers.LionAirProvider.SearchAPIURL,
		Timeout:      cfg.Providers.LionAirProvider.Timeout,
		MaxRetries:   cfg.Providers.LionAirProvider.MaxRetries,
		RateLimitRPS: cfg.Providers.LionAirProvider.RateLimitRPS,
		Limiter:      limiter,
	}))
	factory.AddProvider(batikair.ProviderName, batikair.NewProvider(flightprovider.FlightProviderConfig{
		SearchAPIURL: cfg.Providers.BatikAirProvider.SearchAPIURL,
		Timeout:      cfg.Providers.BatikAirProvider.Timeout,
		MaxRetries:   cfg.Providers.BatikAirProvider.MaxRetries,
		RateLimitRPS: cfg.Providers.BatikAirProvider.RateLimitRPS,
		Limiter:      limiter,
	}))
	factory.AddProvider(airasia.ProviderName, airasia.NewProvider(flightprovider.FlightProviderConfig{
		SearchAPIURL: cfg.Providers.AirAsiaProvider.SearchAPIURL,
		Timeout:      cfg.Providers.AirAsiaProvider.Timeout,
		MaxRetries:   cfg.Providers.AirAsiaProvider.MaxRetries,
		RateLimitRPS: cfg.Providers.AirAsiaProvider.RateLimitRPS,
		Limiter:      limiter,
	}))
	factory.AddProvider(garuda.ProviderName, garuda.NewProvider(flightprovider.FlightProviderConfig{
		SearchAPIURL: cfg.Providers.GarudaProvider.SearchAPIURL,
		Timeout:      cfg.Providers.GarudaProvider.Timeout,
		MaxRetries:   cfg.Providers.GarudaProvider.MaxRetries,
		RateLimitRPS: cfg.Providers.GarudaProvider.RateLimitRPS,
		Limiter:      limiter,
	}))

	return factory
}

func makeAggregatorEndpoint(factory *flightprovider.FlightProviderFactory,
	redisClient *redis.Client, cfg *config.Config) endpoints.AggregatorEndpoint {

	// cache
	flightCache := flight.NewFlightCache(redisClient)

	// service
	aggregatorService := service.NewAggregatorService(factory, flightCache,
		cfg.Providers.CacheExpiration, cfg.Providers.LockTimeout)

	// endpoint
	return endpoints.MakeAggregatorEndpoint(aggregatorService)
}
