package config

import (
	"log/slog"
	"time"
)

type LogLeveler string

func (l LogLeveler) Level() slog.Level {
	var level slog.Level

	_ = level.UnmarshalText([]byte(l))

	return level
}

// Config holds the server configuration.
type Config struct {
	LogLevel  LogLeveler `mapstructure:"LOG_LEVEL"`
	DB        DB         `mapstructure:",squash"`
	HTTP      HTTP       `mapstructure:",squash"`
	Providers Provider   `mapstructure:",squash"`
	Redis     Redis      `mapstructure:",squash"`
}
type DB struct {
	DSN                   string        `mapstructure:"DB_DSN"`
	MaxOpenConnections    int           `mapstructure:"DB_MAX_OPEN_CONNECTIONS"`
	MaxIdleConnections    int           `mapstructure:"DB_MAX_IDLE_CONNECTIONS"`
	MaxConnectionLifetime time.Duration `mapstructure:"DB_MAX_CONNECTIONS_LIFETIME"`
	MaxConnectionIdleTime time.Duration `mapstructure:"DB_MAX_CONNECTION_IDLE_TIME"`
}

type HTTP struct {
	Port    int           `mapstructure:"HTTP_PORT"`
	Timeout time.Duration `mapstructure:"HTTP_TIMEOUT"`
}

type Redis struct {
	Addr     string        `mapstructure:"REDIS_ADDR"`
	Password string        `mapstructure:"REDIS_PASSWORD"`
	DB       int           `mapstructure:"REDIS_DB"`
	Timeout  time.Duration `mapstructure:"REDIS_TIMEOUT"`
}

// Provider holds the provider configuration. url will route to mock provider
type LionAirProvider struct {
	SearchAPIURL string        `mapstructure:"LION_AIR_PROVIDER_SEARCH_API_URL"`
	Timeout      time.Duration `mapstructure:"LION_AIR_PROVIDER_TIMEOUT"`
	MaxRetries   int           `mapstructure:"LION_AIR_PROVIDER_MAX_RETRIES"`
	RateLimitRPS int           `mapstructure:"LION_AIR_PROVIDER_RATE_LIMIT"`
}

type BatikAirProvider struct {
	SearchAPIURL string        `mapstructure:"BATIK_AIR_PROVIDER_SEARCH_API_URL"`
	Timeout      time.Duration `mapstructure:"BATIK_AIR_PROVIDER_TIMEOUT"`
	MaxRetries   int           `mapstructure:"BATIK_AIR_PROVIDER_MAX_RETRIES"`
	RateLimitRPS int           `mapstructure:"BATIK_AIR_PROVIDER_RATE_LIMIT"`
}

type AirAsiaProvider struct {
	SearchAPIURL string        `mapstructure:"AIRASIA_PROVIDER_SEARCH_API_URL"`
	Timeout      time.Duration `mapstructure:"AIRASIA_PROVIDER_TIMEOUT"`
	MaxRetries   int           `mapstructure:"AIRASIA_PROVIDER_MAX_RETRIES"`
	RateLimitRPS int           `mapstructure:"AIRASIA_PROVIDER_RATE_LIMIT"`
}

type GarudaProvider struct {
	SearchAPIURL string        `mapstructure:"GARUDA_PROVIDER_SEARCH_API_URL"`
	Timeout      time.Duration `mapstructure:"GARUDA_PROVIDER_TIMEOUT"`
	MaxRetries   int           `mapstructure:"GARUDA_PROVIDER_MAX_RETRIES"`
	RateLimitRPS int           `mapstructure:"GARUDA_PROVIDER_RATE_LIMIT"`
}

type Provider struct {
	LionAirProvider  LionAirProvider  `mapstructure:",squash"`
	BatikAirProvider BatikAirProvider `mapstructure:",squash"`
	AirAsiaProvider  AirAsiaProvider  `mapstructure:",squash"`
	GarudaProvider   GarudaProvider   `mapstructure:",squash"`
	LockTimeout      time.Duration    `mapstructure:"PROVIDER_LOCK_TIMEOUT"`
	CacheExpiration  time.Duration    `mapstructure:"PROVIDER_CACHE_EXPIRATION"`
}
