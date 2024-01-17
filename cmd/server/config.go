package main

import (
	"fmt"
	"github.com/evgfitil/go-metrics-server.git/internal/logger"
	"github.com/spf13/pflag"
	"net"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	bindAddress string
}

func NewConfig() *Config {
	return &Config{}
}

func getEnvOrDefault(envKey, defaultValue string) string {
	if value := os.Getenv(envKey); value != "" {
		return value
	}
	return defaultValue
}

func (c *Config) setAndValidate(key string, defaultValue string) error {
	var value string
	switch key {
	case "ADDRESS":
		value = getEnvOrDefault(key, defaultValue)
		if err := validateAddress(value); err != nil {
			return err
		}
		c.bindAddress = value
	}
	return nil
}

func (c *Config) ParseFlags() error {
	var addr = pflag.StringP("address", "a", "localhost:8080", "Bind address for the server in the format host:port")
	pflag.Parse()

	if err := c.setAndValidate("ADDRESS", *addr); err != nil {
		logger.Sugar.Fatalf("invalid bind address: %v", err)
		return err
	}
	return nil
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
