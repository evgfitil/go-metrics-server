package main

type Config struct {
	ServerAddress  string `env:"ADDRESS"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
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
