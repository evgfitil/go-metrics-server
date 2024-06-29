package main

type Config struct {
	BindAddress     string `env:"ADDRESS"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	EnablePprof     bool   `env:"ENABLE_PPROF"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
}

func NewConfig() *Config {
	return &Config{}
}
