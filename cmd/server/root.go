package main

import (
	"fmt"
	"github.com/caarlos0/env/v10"
	"github.com/evgfitil/go-metrics-server.git/internal/logger"
	"github.com/evgfitil/go-metrics-server.git/internal/storage"
	"github.com/spf13/cobra"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	defaultBindAddress     = "localhost:8080"
	defaultStoreInterval   = 300 * time.Second
	defaultFileStoragePath = "/tmp/metrics-db.json"
	defaultRestore         = true
)

var (
	cfg     *Config
	rootCmd = &cobra.Command{
		Use:   "metrics-server",
		Short: "A simple server for collecting and storing metrics",
		Long:  `Metrics Server is a lightweight and easy-to-use solution for collecting and storing various metrics.`,
		Run:   runServer,
	}
)

func runServer(cmd *cobra.Command, args []string) {
	if err := env.Parse(cfg); err != nil {
		logger.Sugar.Fatalf("error to parse environment variables: %v", err)
	}
	if err := validateAddress(cfg.BindAddress); err != nil {
		logger.Sugar.Fatalf("invalid bind address: %v", err)
	}

	var saveSignal chan struct{}
	if cfg.StoreInterval == 0 {
		saveSignal = make(chan struct{})
	}

	s := storage.NewMemStorage(saveSignal)
	var fileStorage *storage.FileStorage
	if cfg.FileStoragePath != "" {
		fileStorage, err := storage.NewFileStorage(
			cfg.FileStoragePath, s, cfg.StoreInterval, saveSignal)
		if err != nil {
			logger.Sugar.Fatalf("error initializing file storage: %v", err)
		}
		defer fileStorage.Close()
		if cfg.Restore {
			if err := fileStorage.LoadMetrics(); err != nil {
				logger.Sugar.Errorf("error loading metrics: %v", err)
			}
		}
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		logger.Sugar.Infoln("starting server")
		err := http.ListenAndServe(cfg.BindAddress, logger.WithLogging(MetricsRouter(s)))
		if err != nil {
			logger.Sugar.Fatalf("error starting server: %v", err)
		}
	}()

	<-quit
	logger.Sugar.Info("shutting down server")

	if fileStorage != nil {
		if err := fileStorage.Close(); err != nil {
			logger.Sugar.Errorf("error closing file storage: %v", err)
		}
	}
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
	rootCmd.Flags().StringVarP(&cfg.BindAddress, "address", "a",
		defaultBindAddress, "bind address for the server in the format host:port")
	rootCmd.Flags().DurationVarP(
		&cfg.StoreInterval, "StoreInterval", "i",
		defaultStoreInterval, "interval in seconds for storage data to a file")
	rootCmd.Flags().StringVarP(&cfg.FileStoragePath, "FileStoragePath", "f",
		defaultFileStoragePath, "file path where the server writes its data")
	rootCmd.Flags().BoolP("Restore", "r", defaultRestore,
		"loading previously saved data from a file at startup")
}
