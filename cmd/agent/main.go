package main

import (
	"github.com/evgfitil/go-metrics-server.git/internal/logger"
	"sync"
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
