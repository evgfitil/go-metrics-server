package main

// Config holds the configuration values for the agent.
// These settings can be configured via environment variables or command-line flags.
type Config struct {
	// BatchMode determines whether metrics are sent in batch or individually.
	BatchMode bool `env:"BATCH_MODE"`

	// PollInterval specifies the interval in seconds for polling system metrics.
	PollInterval int `env:"POLL_INTERVAL"`

	// ReportInterval specifies the interval in seconds for reporting metrics to the server.
	ReportInterval int `env:"REPORT_INTERVAL"`

	// ServerAddress specifies the address of the metrics server.
	// Format: "host:port" (e.g., "localhost:8080").
	ServerAddress string `env:"ADDRESS"`

	// UseHTTPS determines whether to use HTTPS for communication with the server.
	UseHTTPS bool `env:"USE_HTTPS"`
}

// NewConfig returns a new instance of Config with default values.
func NewConfig() *Config {
	return &Config{}
}

// GetServerURL constructs the server URL based on the configuration.
func (c Config) GetServerURL() string {
	proto := "http://"
	if c.UseHTTPS {
		proto = "https://"
	}
	return proto + c.ServerAddress
}
