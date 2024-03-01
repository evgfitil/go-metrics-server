package main

import (
	"github.com/evgfitil/go-metrics-server.git/internal/logger"
	"os"
)

func main() {
	logger.InitLogger()
	defer logger.Sugar.Sync()

	if err := Execute(); err != nil {
		logger.Sugar.Fatalf("error starting agent: %v", err)
		os.Exit(1)
	}
}
