package config

import (
	"encoding/json"
	"log/slog"
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

// MustInitConfig initializes configuration from .env file or environment variables.
// If configFile exists, it loads from the file. Otherwise, it automatically binds
// environment variables based on the Config struct's mapstructure tags.
func MustInitConfig(configFile string) Config {
	var (
		vpr = viper.New()
		cfg Config
	)

	// Set default values
	vpr.SetDefault("LOG_LEVEL", "info")

	vpr.AutomaticEnv()

	vpr.SetConfigFile(configFile)
	vpr.SetConfigType("env")

	if err := vpr.ReadInConfig(); err != nil {
		slog.Warn("config file not found or cannot be read, using environment variables",
			slog.String("file", configFile),
			slog.String("error", err.Error()))
	} else {
		slog.Info("config file loaded successfully", slog.String("file", configFile))

		vpr.WatchConfig()
	}

	// Automatically bind all environment variables from Config struct
	bindEnvFromStruct(vpr)

	// Unmarshal configuration into struct
	if err := vpr.Unmarshal(&cfg); err != nil {
		slog.Error("cannot unmarshal config", slog.String("error", err.Error()))
		panic(err)
	}

	return cfg
}

// bindEnvFromStruct automatically binds environment variables based on mapstructure tags using reflection
func bindEnvFromStruct(vpr *viper.Viper) {
	bindEnvFromType(vpr, reflect.TypeOf(Config{}))
}

func bindEnvFromType(vpr *viper.Viper, t reflect.Type) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("mapstructure")

		if tag == "" || tag == "-" {
			// If it's an embedded struct without a tag, recurse
			if field.Anonymous && field.Type.Kind() == reflect.Struct {
				bindEnvFromType(vpr, field.Type)
			}
			continue
		}

		parts := strings.Split(tag, ",")
		envVar := parts[0]
		isSquash := false
		for _, p := range parts {
			if strings.TrimSpace(p) == "squash" {
				isSquash = true
				break
			}
		}

		if isSquash && field.Type.Kind() == reflect.Struct {
			bindEnvFromType(vpr, field.Type)
			continue
		}

		if envVar != "" {
			_ = vpr.BindEnv(envVar)

			// If it's an array of struct, check if the value is a JSON string and unmarshal it
			if (field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.Struct) ||
				field.Type.Kind() == reflect.Struct {
				val := vpr.Get(envVar)
				if s, ok := val.(string); ok && s != "" {
					var jsonVal interface{}
					if err := json.Unmarshal([]byte(s), &jsonVal); err == nil {
						vpr.Set(envVar, jsonVal)
					}
				}
			}
		}
	}
}
