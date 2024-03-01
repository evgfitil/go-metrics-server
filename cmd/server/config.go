package main

type Config struct {
	BindAddress     string `env:"ADDRESS"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	SecretKey       string `env:"KEY"`
}

func NewConfig() *Config {
	return &Config{}
}
