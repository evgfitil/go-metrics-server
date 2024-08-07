// Package main provides the entry point for the metrics agent application.
// This agent collects metrics from the system and sends them to a configured metrics server.
// The agent can operate in batch mode or single metric mode, depending on the configuration.
// Configuration options are available via environment variables or command-line flags.

package main

import (
	"fmt"
	"sync"

	"go.uber.org/zap"

	"github.com/evgfitil/go-metrics-server.git/internal/logger"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

var wg sync.WaitGroup

func main() {
	logger.InitLogger()
	defer func(Sugar *zap.SugaredLogger) {
		err := Sugar.Sync()
		if err != nil {
			fmt.Printf("error syncin logger: %v", err)
		}
	}(logger.Sugar)

	logger.Sugar.Infof("Build version: %s\n", buildVersion)
	logger.Sugar.Infof("Build date: %s\n", buildDate)
	logger.Sugar.Infoln("Build commit: %s\n", buildCommit)

	if err := Execute(); err != nil {
		logger.Sugar.Fatalf("error starting agent: %v", err)
	}

	wg.Wait()
}
