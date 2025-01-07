package main

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Address string
	}
	Logging struct {
		Level string
	}
	API struct {
		Version string
		Routes  map[string]string
	}
	Telemetry struct {
		Enabled         bool   `mapstructure:"enabled"`
		MetricsEndpoint string `mapstructure:"metrics_endpoint"`
	}
	RateLimiting struct {
		Enabled     bool   `mapstructure:"enabled"`
		MaxRequests int    `mapstructure:"max_requests"` // Max requests allowed
		TimeWindow  string `mapstructure:"time_window"`  // Time window (e.g., "1m" for 1 minute)
	}
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %w", err)
	}

	return &config, nil
}
