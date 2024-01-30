package main

import (
	"github.com/caarlos0/env/v10"
	"github.com/evgfitil/go-metrics-server.git/internal/logger"
	"time"
)

type Config struct {
	bindAddress     string        `env:"ADDRESS"`
	storeInterval   time.Duration `env:"STORE_INTERVAL"`
	fileStoragePath string        `env:"FILE_STORAGE_PATH"`
	restore         bool          `env:"RESTORE"`
}

func NewConfig() *Config {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		logger.Sugar.Fatalf("error to parse environment variables: %v", err)
	}
	return cfg
}
