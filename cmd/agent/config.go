package main

type Config struct {
	BatchMode      bool   `env:"BATCH_MODE"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	RateLimit      int    `env:"RATE_LIMIT"`
	ServerAddress  string `env:"ADDRESS"`
	SecretKey      string `env:"KEY"`
	UseHTTPS       bool   `env:"USE_HTTPS"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c Config) GetServerURL() string {
	proto := "http://"
	if c.UseHTTPS {
		proto = "https://"
	}
	return proto + c.ServerAddress
}
