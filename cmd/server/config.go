package main

// Config holds the configuration values for the server.
// These settings can be configured via environment variables or command-line flags.
type Config struct {
	// BindAddress specifies the address the server will bind to.
	// Format: "host:port" (e.g., "localhost:8080").
	BindAddress string `env:"ADDRESS"`

	// DatabaseDSN is the Data Source Name for connecting to a database.
	// If this is set, the server will use the specified database for metrics storage.
	DatabaseDSN string `env:"DATABASE_DSN"`

	// EnablePprof enables pprof for profiling if set to true.
	// pprof will be available on localhost:6060.
	EnablePprof bool `env:"ENABLE_PPROF"`

	// FileStoragePath specifies the path to the file used for file-based storage.
	// Metrics will be stored and restored from this file.
	FileStoragePath string `env:"FILE_STORAGE_PATH"`

	// Restore determines whether the server should restore previously saved metrics from the file.
	// This is only relevant if FileStoragePath is set.
	Restore bool `env:"RESTORE"`

	// StoreInterval specifies the interval in seconds for periodically saving metrics to the file.
	// A value of 0 disables periodic saving.
	StoreInterval int `env:"STORE_INTERVAL"`
}

// NewConfig returns a new instance of Config with default values.
func NewConfig() *Config {
	return &Config{}
}
