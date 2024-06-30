// Package main provides the entry point for the metrics agent application.
// This agent collects metrics from the system and sends them to a configured metrics server.
// The agent can operate in batch mode or single metric mode, depending on the configuration.
// Configuration options are available via environment variables or command-line flags.

package main

import (
	"sync"

	"github.com/evgfitil/go-metrics-server.git/internal/logger"
)

var wg sync.WaitGroup

func main() {
	logger.InitLogger()
	defer logger.Sugar.Sync()

	if err := Execute(); err != nil {
		logger.Sugar.Fatalf("error starting agent: %v", err)
	}

	wg.Wait()
}
