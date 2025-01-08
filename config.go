package main

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Address string `mapstructure:"address"`
	} `mapstructure:"server"`
	Logging struct {
		Level string `mapstructure:"level"`
	} `mapstructure:"logging"`
	API struct {
		Version string            `mapstructure:"version"`
		Routes  map[string]string `mapstructure:"routes"`
	} `mapstructure:"api"`
	Telemetry struct {
		Enabled         bool   `mapstructure:"enabled"`
		MetricsEndpoint string `mapstructure:"metrics_endpoint"`
	} `mapstructure:"telemetry"`
	RateLimiting struct {
		Enabled     bool   `mapstructure:"enabled"`
		MaxRequests int    `mapstructure:"max_requests"`
		TimeWindow  string `mapstructure:"time_window"`
	} `mapstructure:"rate_limiting"`
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
