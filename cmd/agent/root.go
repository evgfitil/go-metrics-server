package main

import (
	"fmt"
	"github.com/caarlos0/env/v10"
	"github.com/evgfitil/go-metrics-server.git/internal/agentcore"
	"github.com/evgfitil/go-metrics-server.git/internal/logger"
	"github.com/evgfitil/go-metrics-server.git/internal/metrics"
	"github.com/spf13/cobra"
	"net"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	defaultServerAddress  = "localhost:8080"
	defaultPollInterval   = 2
	defaultReportInterval = 10
)

var (
	collectedMetrics []agentcore.MetricInterface
	pollCount        int64
	cfg              *Config
	rootCmd          = &cobra.Command{
		Use:   "agent",
		Short: "A simple agent for collecting and sending metrics",
		Long:  `Metrics agent is a lightweight and easy-to-use solution for collecting and sending various metrics`,
		Run:   runAgent,
	}
)

func runAgent(cmd *cobra.Command, args []string) {
	wg.Add(1)
	if err := env.Parse(cfg); err != nil {
		logger.Sugar.Fatalf("error to parse environment variables: %v", err)
	}
	if err := validateAddress(cfg.ServerAddress); err != nil {
		logger.Sugar.Fatalf("invalid address format: %v", err)
	}

	serverURL := cfg.GetServerURL()
	pollInterval := time.Duration(cfg.PollInterval) * time.Second
	reportInterval := time.Duration(cfg.ReportInterval) * time.Second
	lastPollTime, lastReportTime := time.Now(), time.Now()

	go func() {
		defer wg.Done()
		logger.Sugar.Infoln("starting agent")
		for {
			now := time.Now()

			if now.Sub(lastPollTime) >= pollInterval {
				pollCount++
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				collectedMetrics = agentcore.CollectMetrics(&m)
				collectedMetrics = append(collectedMetrics, metrics.NewCounter("PollCount", pollCount))

				lastPollTime = now
			}

			if now.Sub(lastReportTime) > reportInterval {
				agentcore.SendMetrics(collectedMetrics, serverURL)
				collectedMetrics = []agentcore.MetricInterface{}

				lastReportTime = now
			}

			time.Sleep(1 * time.Second)
		}
	}()
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
	return rootCmd.Execute()
}

func init() {
	cfg = NewConfig()
	rootCmd.Flags().StringVarP(&cfg.ServerAddress, "address", "a", defaultServerAddress, "metrics server address in the format host:port")
	rootCmd.Flags().IntVarP(&cfg.PollInterval, "poll-interval", "p", defaultPollInterval, "poll interval in seconds")
	rootCmd.Flags().IntVarP(&cfg.ReportInterval, "report-interval", "r", defaultReportInterval, "report interval in seconds")
}
