package main

import (
	"context"
	"fmt"
	"github.com/caarlos0/env/v10"
	"github.com/evgfitil/go-metrics-server.git/internal/agentcore"
	"github.com/evgfitil/go-metrics-server.git/internal/logger"
	"github.com/evgfitil/go-metrics-server.git/internal/metrics"
	"github.com/spf13/cobra"
	"net"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	defaultServerAddress  = "localhost:8080"
	defaultPollInterval   = 2
	defaultReportInterval = 10
	defaultBatchMode      = true
	defaultRateLimit      = 1
)

var (
	cfg     *Config
	rootCmd = &cobra.Command{
		Use:   "agent",
		Short: "A simple agent for collecting and sending metrics",
		Long:  `Metrics agent is a lightweight and easy-to-use solution for collecting and sending various metrics`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			runAgent(ctx)
		},
	}
)

func runAgent(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(2)

	if err := env.Parse(cfg); err != nil {
		logger.Sugar.Fatalf("error to parse environment variables: %v", err)
	}
	if err := validateAddress(cfg.ServerAddress); err != nil {
		logger.Sugar.Fatalf("invalid address format: %v", err)
	}

	serverURL := cfg.GetServerURL()
	pollInterval := time.Duration(cfg.PollInterval) * time.Second
	reportInterval := time.Duration(cfg.ReportInterval) * time.Second
	metricsChan := make(chan []metrics.Metrics)
	go func() {
		defer wg.Done()
		agentcore.StartCollector(ctx, metricsChan, pollInterval)
	}()

	go func() {
		defer wg.Done()
		agentcore.StartSender(ctx, metricsChan, serverURL, reportInterval, cfg.BatchMode, cfg.SecretKey, cfg.RateLimit)
	}()

	wg.Wait()
}

func validateAddress(addr string) error {
	hp := strings.Split(addr, ":")
	if len(hp) != 2 {
		return fmt.Errorf("address must be in the forman host:port")
	}

	host, portString := hp[0], hp[1]
	port, err := strconv.Atoi(portString)
	if err != nil {
		return fmt.Errorf("ivalid port: %v", err)
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	if net.ParseIP(host) == nil {
		if _, err := net.LookupHost(host); err != nil {
			return fmt.Errorf("invalid host: %v", err)
		}
	}
	return nil
}

func Execute() error {
	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		stop()
		logger.Sugar.Infoln("shutdown signal received, graceful shutdown")
	}()

	return rootCmd.ExecuteContext(ctx)
}

func init() {
	cfg = NewConfig()
	rootCmd.Flags().StringVarP(&cfg.ServerAddress, "address", "a", defaultServerAddress, "metrics server address in the format host:port")
	rootCmd.Flags().IntVarP(&cfg.PollInterval, "poll-interval", "p", defaultPollInterval, "poll interval in seconds")
	rootCmd.Flags().IntVarP(&cfg.ReportInterval, "report-interval", "r", defaultReportInterval, "report interval in seconds")
	rootCmd.Flags().BoolVarP(&cfg.BatchMode, "batch-mode", "b", defaultBatchMode, "send batch of metrics")
	rootCmd.Flags().StringVarP(&cfg.SecretKey, "key", "k", "", "data secret key")
	rootCmd.Flags().IntVarP(&cfg.RateLimit, "rate-limit", "l", defaultRateLimit, "rate limit")
}
