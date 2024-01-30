package main

import (
	"time"
)

type Config struct {
	BindAddress     string        `env:"ADDRESS"`
	StoreInterval   time.Duration `env:"STORE_INTERVAL"`
	FileStoragePath string        `env:"FILE_STORAGE_PATH"`
	Restore         bool          `env:"RESTORE"`
}

func NewConfig() *Config {
	return &Config{}
}
